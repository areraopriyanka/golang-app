package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/plaid"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
	plaidSDK "github.com/plaid/plaid-go/v34/plaid"
	"github.com/riverqueue/river"
)

type CloseSuspendedAccountArgs struct {
	UserId string `json:"userId"`
}

func (CloseSuspendedAccountArgs) Kind() string { return "close_account" }

type CloseSuspendedAccountWorker struct {
	river.WorkerDefaults[CloseSuspendedAccountArgs]
	Plaid *plaidSDK.APIClient
}

func (w *CloseSuspendedAccountWorker) Work(ctx context.Context, job *river.Job[CloseSuspendedAccountArgs]) error {
	account, err := dao.UserAccountCardDao{}.FindOneByUserId(db.DB, job.Args.UserId)
	if err != nil {
		return err
	}
	reason := "Account closed due to completion of the 60-day suspension period"
	_, err = UpdateAccountStatus(ledger.CLOSED, account.AccountNumber, reason)
	if err != nil {
		logging.Logger.Error("Failed to close the suspended account for user", "userId", job.Args.UserId, "error", err)
		return err
	}
	// Unlink all external bank accounts if user had only one account which is CLOSED
	allClosed, err := checkIfAllAccountsAreClosed(job.Args.UserId)
	if err != nil {
		return err
	}
	if allClosed {
		return unlinkAllExternalBankAccounts(job.Args.UserId, w.Plaid)
	}

	logging.Logger.Debug("Account closed successfully", "userId", job.Args.UserId)
	return nil
}

func RegisterCloseSuspendedAccountWorker(workers *river.Workers, plaid *plaidSDK.APIClient) {
	river.AddWorker(workers, &CloseSuspendedAccountWorker{Plaid: plaid})
}

type EnqueueCloseAccountsArgs struct{}

func (EnqueueCloseAccountsArgs) Kind() string {
	return "close_accounts_periodic"
}

type EnqueueCloseAccountsWorker struct {
	river.WorkerDefaults[EnqueueCloseAccountsArgs]
	RiverClient *river.Client[*sql.Tx]
}

func (w *EnqueueCloseAccountsWorker) SetRiverClientForCloseAccountsEnqueueWorker(client *river.Client[*sql.Tx]) {
	w.RiverClient = client
}

func RegisterCloseAccountsEnqueueWorker(workers *river.Workers) *EnqueueCloseAccountsWorker {
	worker := &EnqueueCloseAccountsWorker{}
	river.AddWorker(workers, worker)
	return worker
}

func (w *EnqueueCloseAccountsWorker) Work(ctx context.Context, job *river.Job[EnqueueCloseAccountsArgs]) error {
	return EnqueueCloseAccountJob(ctx, w.RiverClient)
}

// @Summary SuspendAccount
// @Description Update the account status to 'SUSPENDED' to suspend the account for 60 days
// @Tags accounts
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @param closeAccountRequest body CloseAccountRequest true "payload"
// @Success 200 {object} response.UpdateAccountStatusResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/close [put]
func SuspendAccount(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	var requestData CloseAccountRequest

	err := c.Bind(&requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	userAccountCard, errResponse := dao.RequireActiveCardHolderForUser(userId)
	if errResponse != nil {
		return errResponse
	}

	if errResponse := closeCard(logger, user.LedgerCustomerNumber, userAccountCard.CardId, userAccountCard.AccountNumber); errResponse != nil {
		return errResponse
	}

	status, err := UpdateAccountStatus(ledger.SUSPENDED, userAccountCard.AccountNumber, requestData.AccountClosureReason)
	if err != nil {
		logger.Error("Error while updating account status", "error", err)
		return response.InternalServerError(fmt.Sprintf("Error while updating account status: %s", err), errtrace.Wrap(err))
	}

	return c.JSON(http.StatusOK, response.UpdateAccountStatusResponse{
		UpdatedAccountStatus: ledger.MapLedgerAccountStatus(status, logger),
	})
}

type CloseAccountRequest struct {
	AccountClosureReason string `json:"accountClosureReason" validate:"required,oneof='Not using it' 'Switched banks' 'Support issues' 'Lack of tools and benefits' 'App problems' 'Life changes' 'No progress' 'Other',max=255"`
}

func closeCard(logger *slog.Logger, ledgerCustomerNumber string, cardId string, accountNumber string) *response.ErrorResponse {
	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, ledgerParamsBuilder)

	getCardRequest := ledgerClient.BuildGetCardDetailsRequest(ledgerCustomerNumber, accountNumber, cardId)
	getCardResponse, err := ledgerClient.GetCardDetails(getCardRequest)
	if err != nil {
		return &response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Error calling ledger getCardDetails: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}
	if getCardResponse.Error != nil {
		logger.Error("Error receiving card details", "code", getCardResponse.Error.Code, "msg", getCardResponse.Error.Message)
		return &response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error receiving card details: %s", getCardResponse.Error.Message), MaybeInnerError: errtrace.New("")}
	}
	if getCardResponse.Result == nil {
		logger.Error("The ledger responded with an empty result object", "response", getCardResponse)
		return &response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: "The ledger responded with an empty result object", MaybeInnerError: errtrace.New("")}
	}

	switch getCardResponse.Result.Card.CardStatus {
	case "CLOSED", "EXPIRE_CARD", "CARD_REQUEST_NOT_PROCESSED":
		logger.Info("Card is already closed")
		return nil

	case "TEMPRORY_BLOCKED_BY_CLIENT":
		if err = updateCardStatus(*ledgerClient, ledgerCustomerNumber, accountNumber, cardId, ledger.UNLOCK); err != nil {
			logger.Error("Failed to unlock card", "error", err)
			return &response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Failed to unlock card: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
		}

	case "TEMPRORY_BLOCKED_BY_ADMIN":
		if err = updateCardStatus(*ledgerClient, ledgerCustomerNumber, accountNumber, cardId, ledger.UNBLOCK); err != nil {
			logger.Error("Failed to unblock card", "error", err)
			return &response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Failed to unblock card: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
		}
	}

	if err = updateCardStatus(*ledgerClient, ledgerCustomerNumber, accountNumber, cardId, ledger.CLOSE); err != nil {
		logger.Error("Failed to close card", "error", err)
		return &response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Failed to close card: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	return nil
}

func updateCardStatus(ledgerClient ledger.NetXDCardApiClient, ledgerCustomerNumber, accountNumber, cardId, action string) error {
	userClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, nil)
	updateRequest, err := userClient.BuildUpdateStatusRequest(ledgerCustomerNumber, cardId, accountNumber, action, "", false)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("error building updateStatus payload: %w", err))
	}

	updateResponse, err := ledgerClient.UpdateStatus(*updateRequest)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("error calling ledger updateStatus: %w", err))
	}

	if updateResponse.Error != nil {
		return errtrace.Wrap(fmt.Errorf("error updating card status: %+v", updateResponse.Error))
	}

	if updateResponse.Result == nil {
		return errtrace.Wrap(fmt.Errorf("ledger responded with an empty result object"))
	}

	return nil
}

func UpdateAccountStatus(status, accountNumber, reason string) (string, error) {
	ledgerClient := ledger.CreateLedgerApiClient(config.Config.Ledger)
	payload := ledger.BuildUpdateAccountStatusRequest(accountNumber, status)

	updateAccountStatusResponse, err := ledgerClient.UpdateAccountStatus(payload)
	if err != nil {
		return "", errtrace.Wrap(fmt.Errorf("error calling ledger updateAccountStatus API: %s", err.Error()))
	}

	if updateAccountStatusResponse.Error != nil {
		return "", errtrace.Wrap(fmt.Errorf("ledger responded with an error to updateAccountStatus API: %s", updateAccountStatusResponse.Error.Message))
	}
	userAccountCard := dao.UserAccountCardDao{
		AccountStatus:        status,
		AccountClosureReason: reason,
	}
	now := clock.Now()
	switch status {
	case ledger.SUSPENDED:
		userAccountCard.SuspendedAt = &now
	case ledger.CLOSED:
		userAccountCard.ClosedAt = &now
	}

	if updateAccountStatusResponse.Result != nil && updateAccountStatusResponse.Result.Status == status {
		updateAccountResult := db.DB.Model(dao.UserAccountCardDao{}).Where("account_number=?", accountNumber).Updates(userAccountCard)
		if updateAccountResult.Error != nil {
			return "", errtrace.Wrap(fmt.Errorf("error updating user's status: %s", updateAccountResult.Error.Error()))
		}

		if updateAccountResult.RowsAffected == 0 {
			return "", errtrace.Wrap(fmt.Errorf("error updating user's status"))
		}
		return updateAccountStatusResponse.Result.Status, nil
	}

	return "", errtrace.Wrap(fmt.Errorf("ledger returned unexpected response"))
}

func EnqueueCloseAccountJob(ctx context.Context, riverClient *river.Client[*sql.Tx]) error {
	records, err := dao.UserAccountCardDao{}.GetSuspendedAccounts60DaysOld(db.DB)
	if err != nil {
		logging.Logger.Error("Failed to fetch suspended accounts", "error", err)
		return errtrace.Wrap(err)
	}

	if len(records) == 0 {
		logging.Logger.Debug("No suspended accounts found")
		return nil
	}

	var batchParams []river.InsertManyParams
	for _, userAccountCard := range records {
		batchParams = append(batchParams, river.InsertManyParams{
			Args: CloseSuspendedAccountArgs{
				UserId: userAccountCard.UserId,
			},
		})
	}

	_, err = riverClient.InsertMany(ctx, batchParams)
	if err != nil {
		logging.Logger.Error("Failed to enqueue credit score jobs", "error", err.Error())
		return errtrace.Wrap(err)
	}

	logging.Logger.Debug("Enqueued close account job successfully", "userCount", len(records))
	return nil
}

func checkIfAllAccountsAreClosed(userId string) (bool, error) {
	var count int
	// This will work even if we support multiple accounts per user in the future
	result := db.DB.Model(&dao.UserAccountCardDao{}).
		Where("user_id = ? AND account_status <> ?", userId, "CLOSED").
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	if count == 0 {
		return true, nil
	}
	return false, nil
}

func unlinkAllExternalBankAccounts(userId string, plaidAPIClient *plaidSDK.APIClient) error {
	records, err := dao.PlaidItemDao{}.GetItemsByUserId(userId)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		logging.Logger.Info("No records found")
		return nil
	}
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: plaidAPIClient, DB: db.DB, WebhookURL: config.Config.Plaid.WebhookURL}
	for _, item := range records {
		accessToken, err := utils.DecryptPlaidAccessToken(item.EncryptedAccessToken, item.KmsEncryptedAccessToken)
		if err != nil {
			logging.Logger.Error("Could not decrypt Plaid access token", "error", err.Error())
			return err
		}

		err = ps.UnlinkItem(userId, item.PlaidItemID, accessToken)
		if err != nil {
			logging.Logger.Error("error unlinking item", "error", err.Error())
			return err
		}
	}
	logging.Logger.Debug("All external bank accounts are unlinked successfully", "userId", userId)
	return nil
}
