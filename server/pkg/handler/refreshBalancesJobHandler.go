package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/plaid"
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"sort"
	"strconv"
	"time"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
	plaidSDK "github.com/plaid/plaid-go/v34/plaid"
	"github.com/riverqueue/river"
)

// @Summary Start Balance Refresh Job
// @Description Initiates a background job to refresh balance information for the user's Plaid accounts that haven't been updated in the last 24 hrs. Returns immediately without waiting for job completion.
// @Tags plaid,balance,ach
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} BalanceRefreshResponse "No refresh needed"
// @Success 201 {object} BalanceRefreshResponse "Job created successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized access"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /account/balance/refresh [post]
func (h *Handler) BalanceRefresh(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.ErrorResponse{
			ErrorCode:       constant.UNAUTHORIZED_ACCESS_ERROR,
			Message:         constant.UNAUTHORIZED_ACCESS_ERROR_MSG,
			StatusCode:      http.StatusUnauthorized,
			LogMessage:      "Failed to get user Id from custom context",
			MaybeInnerError: errtrace.New(""),
		}
	}
	userId := cc.UserId
	logger := logging.GetEchoContextLogger(c).WithGroup("BalanceRefresh").With("userId", userId)

	staleAccounts, err := dao.PlaidAccountDao{}.FindNonErrorAccountsWithStaleBalanceForUser(userId, nil)
	if err != nil {
		logger.Error("Failed to query stale accounts", "error", err)
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Failed to query stale accounts: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	if len(staleAccounts) == 0 {
		logger.Debug("No stale accounts found, skipping refresh")
		resp := BalanceRefreshResponse{}
		return c.JSON(http.StatusOK, resp)
	}

	// Accounts can share Plaid Items; make a set to eliminate duplicates:
	itemIds := make(map[string]bool)
	for _, account := range staleAccounts {
		itemIds[account.PlaidItemID] = true
	}
	plaidItemIds := make([]string, 0, len(itemIds))
	for itemId := range itemIds {
		plaidItemIds = append(plaidItemIds, itemId)
	}
	// It's important that we keep plaidItemIds sorted; otherwise
	// `byArgs` uniqueness constraint for river won't detect a duplicate
	sort.Strings(plaidItemIds)

	ctx := context.Background()
	jobInsertResult, err := h.RiverClient.Insert(ctx, RefreshBalancesArgs{
		UserID:       userId,
		PlaidItemIds: plaidItemIds,
	}, nil)
	if err != nil {
		logger.Error("Failed to enqueue refresh balances job", "error", err)
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Failed to enqueue refresh balances job: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	jobId := jobInsertResult.Job.ID

	logger.Debug("Enqueued refresh balances job successfully", "jobId", jobId, "itemCount", len(plaidItemIds))

	resp := BalanceRefreshResponse{
		JobId: &jobId,
	}

	return c.JSON(http.StatusCreated, resp)
}

// @Summary Get Balance Refresh Status
// @Description Check if the user has any running balance refresh jobs and return current balance information
// @Tags plaid,balance,ach
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Param jobId path int true "Job ID to check status for"
// @Success 200 {object} BalanceRefreshStatusResponse "Refresh status and current balances"
// @Failure 401 {object} response.ErrorResponse "Unauthorized access"
// @Failure 400 {object} response.ErrorResponse "Invalid job ID"
// @Failure 404 {object} response.ErrorResponse "Could not find job"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /account/balance/refresh/status/{jobId} [get]
func (h *Handler) BalanceRefreshStatus(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.ErrorResponse{
			ErrorCode:       constant.UNAUTHORIZED_ACCESS_ERROR,
			Message:         constant.UNAUTHORIZED_ACCESS_ERROR_MSG,
			StatusCode:      http.StatusUnauthorized,
			LogMessage:      "Failed to get user Id from custom context",
			MaybeInnerError: errtrace.New(""),
		}
	}
	userId := cc.UserId
	logger := logging.GetEchoContextLogger(c).WithGroup("BalanceRefreshStatus").With("userId", userId)

	jobIdParam := c.Param("jobId")
	jobId, err := strconv.ParseInt(jobIdParam, 10, 64)
	if err != nil {
		logger.Error("Invalid jobId parameter", "jobId", jobIdParam, "error", err)
		return response.ErrorResponse{
			Message:         "Invalid job ID",
			StatusCode:      http.StatusBadRequest,
			LogMessage:      "Failed to parse jobId parameter",
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	isRefreshing := false
	ctx := context.Background()
	job, err := h.RiverClient.JobGet(ctx, jobId)
	if errors.Is(err, river.ErrNotFound) {
		logger.Error("Failed to get credit score job status", "jobId", jobId, "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, MaybeInnerError: errtrace.Wrap(err)}
	}
	if job.Kind != (RefreshBalancesArgs{}).Kind() {
		logger.Error("Job is not a RefreshBalances job", "jobId", jobId, "kind", job.Kind)
		return response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, MaybeInnerError: errtrace.New("")}
	}
	var args RefreshBalancesArgs
	if err := json.Unmarshal(job.EncodedArgs, &args); err != nil {
		logger.Error("Job args failed to Unmarshal", "jobId", jobId, "err", err)
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Job args failed to Unmarshal: jobId: %d, err: %s", jobId, err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}
	if args.UserID != userId {
		logger.Error("Job userId does not match requesting user", "jobId", jobId, "userId", args.UserID)
		return response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, MaybeInnerError: errtrace.New("")}
	}
	switch job.State {
	// If the job is about to run, or is running, report that we're refreshing
	case "available", "running":
		isRefreshing = true
	// If the job's done working, or won't be worked, don't report we're refreshing
	case "completed", "cancelled", "discarded":
		isRefreshing = false
	// If it's going to be retried, or it's scheduled to run later, don't report
	// that we're refreshing, otherwise the client will continue to poll for updates
	case "retryable", "scheduled":
		isRefreshing = false
	default:
		isRefreshing = false
	}

	accounts, err := dao.PlaidAccountDao{}.FindAccountsForUser(userId)
	if err != nil {
		logger.Error("Failed to fetch user accounts", "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Failed to fetch user accounts: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	balances := make([]AccountBalance, 0, len(accounts))
	for _, account := range accounts {
		balances = append(balances, AccountBalance{
			ID:                    account.ID,
			AvailableBalanceCents: account.AvailableBalanceCents,
		})
	}

	resp := BalanceRefreshStatusResponse{
		IsRefreshing: isRefreshing,
		Balances:     balances,
	}

	return c.JSON(http.StatusOK, resp)
}

type RefreshBalancesOutput struct {
	RefreshedAccounts []AccountBalance `json:"refreshedAccounts"`
}

type BalanceRefreshResponse struct {
	JobId *int64 `json:"jobId,omitempty"`
}

type BalanceRefreshStatusResponse struct {
	IsRefreshing bool             `json:"isRefreshing" validate:"required"`
	Balances     []AccountBalance `json:"balances" validate:"required"`
}

type AccountBalance struct {
	ID                    string `json:"id" validate:"required"`
	AvailableBalanceCents *int64 `json:"availableBalanceCents,omitempty"`
}

type RefreshBalancesArgs struct {
	UserID       string   `json:"userId"`
	PlaidItemIds []string `json:"plaidItemIds"`
}

func (RefreshBalancesArgs) Kind() string { return "refresh_balances" }

func (RefreshBalancesArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: "plaid",
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 24 * time.Hour,
		},
	}
}

type RefreshBalancesWorker struct {
	river.WorkerDefaults[RefreshBalancesArgs]
	Plaid *plaidSDK.APIClient
}

func (w *RefreshBalancesWorker) Work(ctx context.Context, job *river.Job[RefreshBalancesArgs]) error {
	logger := logging.Logger.WithGroup("RefreshBalancesWorker").With("userId", job.Args.UserID, "jobId", job.ID)
	refreshedAccounts := []AccountBalance{}

	for _, plaidItemId := range job.Args.PlaidItemIds {
		itemLogger := logger.With("plaidItemId", plaidItemId)

		plaidItem, err := dao.PlaidItemDao{}.GetItemForUserByItemID(job.Args.UserID, plaidItemId)
		if err != nil {
			itemLogger.Error("db query failed for Plaid item", "error", err.Error())
			continue
		}
		if plaidItem == nil {
			itemLogger.Error("failed to find Plaid item", "error", "item not found")
			continue
		}

		// item errors indicate that the user will have to re-link the account before we can access it; skip it
		if plaidItem.ItemError != nil {
			itemLogger.Debug("skipping balance refresh due to item error", "item_error", plaidItem.ItemError)
			continue
		}

		accessToken, err := utils.DecryptPlaidAccessToken(plaidItem.EncryptedAccessToken, plaidItem.KmsEncryptedAccessToken)
		if err != nil {
			itemLogger.Error("failed to decrypt access token", "error", err.Error())
			continue
		}

		ps := plaid.PlaidService{Logger: itemLogger, Plaid: w.Plaid, DB: db.DB}
		err = ps.AccountsBalanceGetRequest(job.Args.UserID, plaidItemId, accessToken)
		if err != nil {
			itemLogger.Error("failed to refresh balances from Plaid", "error", err.Error())
			continue
		}

		var accounts []dao.PlaidAccountDao
		err = db.DB.Where("plaid_item_id = ? AND user_id = ?", plaidItemId, job.Args.UserID).Find(&accounts).Error
		if err != nil {
			itemLogger.Error("failed to fetch updated accounts", "error", err.Error())
			continue
		}

		for _, account := range accounts {
			refreshedAccounts = append(refreshedAccounts, AccountBalance{
				ID:                    account.ID,
				AvailableBalanceCents: account.AvailableBalanceCents,
			})
		}

		itemLogger.Debug("Successfully refreshed balances for Plaid item")
	}

	output := RefreshBalancesOutput{RefreshedAccounts: refreshedAccounts}

	outputMap := map[string]any{"refreshedAccounts": output.RefreshedAccounts}

	if err := river.RecordOutput(ctx, outputMap); err != nil {
		logger.Error("Error recording refresh balances job output", "err", err)
		return fmt.Errorf("failed to record job output: %w", err)
	}

	logger.Debug("Successfully completed balance refresh", "refreshedCount", len(refreshedAccounts))

	return nil
}

func RegisterRefreshBalancesWorker(workers *river.Workers, plaid *plaidSDK.APIClient) {
	river.AddWorker(workers, &RefreshBalancesWorker{Plaid: plaid})
}
