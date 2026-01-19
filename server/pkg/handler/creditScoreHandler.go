package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/debtwise"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"process-api/pkg/validators"
	"strconv"
	"time"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/riverqueue/river"
)

// @Summary Start Credit Score Job
// @Description Starts job sending request to debtwise for credit score
// @Tags credit score
// @Accept json
// @Produce json
// @param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} response.MaybeRiverJobResponse[CreditScoreResponse]
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/debtwise/equifax/complete [post]
func (h *Handler) CompleteDebtwiseOnboarding(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	ctx := context.Background()
	jobInsertResult, err := h.RiverClient.Insert(ctx, CreditScoreJobArgs{
		UserId: userId,
	}, nil)
	if err != nil {
		logger.Error("Failed to enqueue credit score job", "error", err)
		return response.InternalServerError(fmt.Sprintf("Failed to enqueue credit score job: %s", err.Error()), errtrace.Wrap(err))
	} else {
		logger.Debug("Enqueued credit score job successfully")
	}

	var payload *CreditScoreResponse = nil
	jobId := jobInsertResult.Job.ID
	var jobState string

	// Poll job status up to 3 times, every 500ms, include payload in response if job is complete
	for i := 0; i < 3; i++ {
		job, err := h.RiverClient.JobGet(ctx, jobId)
		if err != nil {
			logger.Error("Failed to fetch credit score job status", "jobId", jobId, "error", err.Error())
			break
		}

		jobState = string(job.State)

		if job.State == "completed" && job.Output() != nil {
			var output dao.UserCreditScoreDao
			if err := json.Unmarshal(job.Output(), &output); err == nil {
				creditScoreHistory, err := dao.UserCreditScoreDao{}.GetCreditScoreHistory(userId)
				if err != nil {
					return response.ErrorResponse{
						ErrorCode:       constant.INTERNAL_SERVER_ERROR,
						Message:         "Unable to retrieve credit score history",
						StatusCode:      http.StatusInternalServerError,
						LogMessage:      fmt.Sprintf("Error querying recent credit score history: %s", err.Error()),
						MaybeInnerError: errtrace.Wrap(err),
					}
				}

				creditScoreResponse := convertDaoToCreditScoreResponse(&output, creditScoreHistory)
				payload = &creditScoreResponse
			}
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	response := response.MaybeRiverJobResponse[CreditScoreResponse]{
		JobId:   int(jobInsertResult.Job.ID),
		State:   jobState,
		Payload: payload,
	}

	return cc.JSON(http.StatusOK, response)
}

// @Summary Get Credit Score Job Status
// @Description Gets status of credit score job
// @Tags credit score
// @Accept json
// @Produce json
// @Param jobId path int true "Job Id"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} response.MaybeRiverJobResponse[CreditScoreResponse]
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/debtwise/equifax/complete/{jobId} [get]
func (h *Handler) GetCreditScoreJobStatus(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	jobIdStr := c.Param("jobId")
	jobId, err := strconv.Atoi(jobIdStr)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	ctx := context.Background()
	job, err := h.RiverClient.JobGet(ctx, int64(jobId))
	if errors.Is(err, river.ErrNotFound) {
		logger.Error("Failed to get credit score job status", "jobId", jobId, "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, MaybeInnerError: errtrace.Wrap(err)}
	}
	if err != nil {
		logger.Error("Failed to get credit score job status", "jobId", jobId, "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Failed to get credit score job status. jobId: %d error: %s", jobId, err.Error()), errtrace.Wrap(err))
	}
	if job.Kind != (CreditScoreJobArgs{}).Kind() {
		logger.Error("Job has not CreditScoreJob", "jobId", jobId, "kind", job.Kind)
		return response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, MaybeInnerError: errtrace.New("")}
	}
	var args CreditScoreJobArgs
	if err := json.Unmarshal(job.EncodedArgs, &args); err != nil {
		logger.Error("Job args failed to Unmarshal", "jobId", jobId, "err", err)
		return response.InternalServerError(fmt.Sprintf("Job args failed to Unmarshal. jobId: %d error: %s", jobId, err.Error()), errtrace.Wrap(err))
	}
	if args.UserId != userId {
		logger.Error("Job userId does not match requesting user", "jobId", jobId, "userId", args.UserId)
		return response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, MaybeInnerError: errtrace.New("")}
	}

	var payload *CreditScoreResponse = nil
	if job.State == "completed" && job.Output() != nil {
		var output dao.UserCreditScoreDao
		if err := json.Unmarshal(job.Output(), &output); err != nil {
			logger.Error("Job output failed to Unmarshal", "jobId", jobId, "err", err)
			return response.InternalServerError(fmt.Sprintf("Job output failed to Unmarshal. jobId: %d error: %s", jobId, err.Error()), errtrace.Wrap(err))
		}
		creditScoreHistory, err := dao.UserCreditScoreDao{}.GetCreditScoreHistory(userId)
		if err != nil {
			return response.ErrorResponse{
				ErrorCode:       constant.INTERNAL_SERVER_ERROR,
				Message:         "Unable to retrieve credit score history",
				StatusCode:      http.StatusInternalServerError,
				LogMessage:      fmt.Sprintf("Error querying recent credit score history: %s", err.Error()),
				MaybeInnerError: errtrace.New(""),
			}
		}

		creditScoreResponse := convertDaoToCreditScoreResponse(&output, creditScoreHistory)
		payload = &creditScoreResponse
	}

	creditScoreJobResponse := response.MaybeRiverJobResponse[CreditScoreResponse]{
		JobId:   int(job.ID),
		State:   string(job.State),
		Payload: payload,
	}

	return c.JSON(http.StatusOK, creditScoreJobResponse)
}

// @Summary LatestCreditScoreHandler
// @Description Retrieves latest debtwise credit score for user
// @Tags credit score
// @Accept json
// @Produce json
// @param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} CreditScoreResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/debtwise/latest_credit_score [get]
func LatestCreditScoreHandler(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId

	logger := logging.GetEchoContextLogger(cc)

	latestScore, err := dao.UserCreditScoreDao{}.FindLatestUserCreditScoreByUserId(userId)
	if err != nil {
		logger.Error("Error retireiving latest CreditScore record", "err", err)
		return response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, Message: constant.NO_DATA_FOUND_MSG, StatusCode: http.StatusNotFound, LogMessage: fmt.Sprintf("Credit score record not found: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	creditScoreHistory, err := dao.UserCreditScoreDao{}.GetCreditScoreHistory(userId)
	if err != nil {
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			Message:         "Unable to retrieve credit score history",
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Error querying recent credit score history: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	return cc.JSON(http.StatusOK, convertDaoToCreditScoreResponse(latestScore, creditScoreHistory))
}

func FetchAndStoreCreditScore(ctx context.Context, userId string) (*dao.UserCreditScoreDao, error) {
	debtwiseClient, err := debtwise.NewDebtwiseClient(config.Config.Debtwise, logging.Logger)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to create Debtwise client: %w", err))
	}

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to fetch user: %s", errResponse.Message))
	}

	if user.DebtwiseCustomerNumber == nil {
		return nil, errtrace.Wrap(fmt.Errorf("invalid Debtwise onboarding status: missing customer number for userId %s", userId))
	}
	debtwiseCustomerNumber := *user.DebtwiseCustomerNumber

	if user.DebtwiseOnboardingStatus == "uninitialized" {
		return nil, errtrace.Wrap(fmt.Errorf("invalid Debtwise onboarding status of uninitialized for userId %s", userId))
	}

	maybeTotalAccounts, err := retrieveTotalAccounts(debtwiseCustomerNumber, user, debtwiseClient)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to fetch credit report: %w", err))
	}
	if maybeTotalAccounts == nil {
		return nil, errtrace.Wrap(fmt.Errorf("error fetching credit report: totalAccounts is nil"))
	}
	totalAccounts := *maybeTotalAccounts

	debtwiseResponse, err := debtwiseClient.RetrieveCreditScoreWithResponse(
		ctx,
		debtwise.UserIdParam(debtwiseCustomerNumber),
		&debtwise.RetrieveCreditScoreParams{},
	)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to fetch credit score: %w", err))
	}
	if debtwiseResponse.JSON200 == nil {
		return nil, errtrace.Wrap(fmt.Errorf("unexpected response from Debtwise: JSON200 is nil"))
	}

	latestScore, err := dao.UserCreditScoreDao{}.FindLatestUserCreditScoreByUserId(userId)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to query latest credit score for userId %s, err: %w", userId, err))
	}

	apiScoreDate, err := time.Parse("2006-01-02", debtwiseResponse.JSON200.Date)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to parse API credit score date %s: %w", debtwiseResponse.JSON200.Date, err))
	}

	if latestScore == nil || !latestScore.Date.Equal(apiScoreDate) {
		latestScore, err = saveCreditScoreToDB(debtwiseResponse.JSON200, totalAccounts, debtwiseCustomerNumber, userId)
		if err != nil {
			return nil, errtrace.Wrap(fmt.Errorf("failed to save credit score to database for userId %s, err: %w", userId, err))
		}
		logging.Logger.Debug("Credit score record saved for user", "userId", userId)
	}

	return latestScore, nil
}

const (
	RatingNeedsWork string = "Needs Work"
	RatingFair      string = "Fair"
	RatingGood      string = "Good"
	RatingExcellent string = "Excellent"
)

// NOTE: The credit factor ratings are derived from the design system in Figma
func MapPaymentHistory(amount float32) string {
	switch {
	case amount == 100:
		return RatingExcellent
	case amount >= 99:
		return RatingGood
	case amount >= 98:
		return RatingFair
	default:
		return RatingNeedsWork
	}
}

func MapCreditUtilization(amount float32) string {
	switch {
	case amount <= 20:
		return RatingExcellent
	case amount <= 40:
		return RatingGood
	case amount <= 60:
		return RatingFair
	default:
		return RatingNeedsWork
	}
}

func MapDerogatoryMarks(amount int) string {
	switch amount {
	case 0:
		return RatingExcellent
	case 1:
		return RatingGood
	case 2:
		return RatingFair
	default:
		return RatingNeedsWork
	}
}

func MapCreditAge(amount float32) string {
	switch {
	case amount < 5:
		return RatingNeedsWork
	case amount >= 5 && amount < 7:
		return RatingFair
	case amount >= 7 && amount < 9:
		return RatingGood
	default:
		return RatingExcellent
	}
}

func MapCreditAgeToString(amount float64) string {
	years := int(amount)
	months := int((amount - float64(years)) * 12)
	if years == 0 && months == 0 {
		return "No credit history"
	}
	return fmt.Sprintf("%d yrs %d mos", years, months)
}

func MapCreditMix(amount int) string {
	switch amount {
	case 0:
		return RatingNeedsWork
	case 1:
		return RatingFair
	case 2:
		return RatingGood
	default:
		return RatingExcellent
	}
}

func MapNewCredit(amount int) string {
	switch amount {
	case 0:
		return RatingExcellent
	case 1:
		return RatingGood
	case 2:
		return RatingFair
	default:
		return RatingNeedsWork
	}
}

func MapTotalAccounts(amount int) string {
	switch {
	case amount <= 5:
		return RatingNeedsWork
	case amount >= 6 && amount <= 12:
		return RatingFair
	case amount >= 13 && amount <= 21:
		return RatingGood
	default:
		return RatingExcellent
	}
}

func convertDaoToCreditScoreResponse(record *dao.UserCreditScoreDao, encryptedCreditScoreHistory []dao.EncryptedCreditScoreSummary) CreditScoreResponse {
	jsonString, err := utils.DecryptKmsBinary(record.EncryptedCreditData)
	if err != nil {
		logging.Logger.Error("Failed to decrypt encrypted credit data", "err", err)
	}

	var creditData dao.EncryptableCreditScoreData
	if err := json.Unmarshal([]byte(jsonString), &creditData); err != nil {
		logging.Logger.Error("Failed to unmarhsal encrypted credit data", "err", err)
	}

	if err := validators.ValidateStruct(creditData); err != nil {
		logging.Logger.Error("Failed to validate EncryptableCreditScoreData", "err", err)
	}

	var creditScoreHistory []dao.CreditScoreSummary
	for _, historyRecord := range encryptedCreditScoreHistory {
		jsonString, err := utils.DecryptKmsBinary(historyRecord.EncryptedCreditData)
		if err != nil {
			logging.Logger.Error("Failed to decrypt encrypted credit data", "err", err)
		}

		var creditData dao.EncryptableCreditScoreData
		if err := json.Unmarshal([]byte(jsonString), &creditData); err != nil {
			logging.Logger.Error("Failed to unmarhsal encrypted credit data", "err", err)
		}

		if err := validators.ValidateStruct(creditData); err != nil {
			logging.Logger.Error("Failed to validate EncryptableCreditScoreData", "err", err)
		}

		creditScoreHistory = append(creditScoreHistory, dao.CreditScoreSummary{
			Date:  historyRecord.Date,
			Score: creditData.Score,
		})
	}

	return CreditScoreResponse{
		Date:     record.Date.Format("2006-01-02"),
		Score:    creditData.Score,
		Increase: creditData.Increase,
		PaymentHistory: CreditFactorFloat{
			Amount: creditData.PaymentHistoryAmount,
			Factor: creditData.PaymentHistoryFactor,
		},
		CreditUtilization: CreditFactorFloat{
			Amount: creditData.CreditUtilizationAmount,
			Factor: creditData.CreditUtilizationFactor,
		},
		DerogatoryMarks: CreditFactorInt{
			Amount: creditData.DerogatoryMarksAmount,
			Factor: creditData.DerogatoryMarksFactor,
		},
		CreditAge: CreditAge{
			Amount: MapCreditAgeToString(creditData.CreditAgeAmount),
			Factor: creditData.CreditAgeFactor,
		},
		CreditMix: CreditFactorInt{
			Amount: creditData.CreditMixAmount,
			Factor: creditData.CreditMixFactor,
		},
		NewCredit: CreditFactorInt{
			Amount: creditData.NewCreditAmount,
			Factor: creditData.NewCreditFactor,
		},
		TotalAccounts: CreditFactorInt{
			Amount: creditData.TotalAccountsAmount,
			Factor: creditData.TotalAccountsFactor,
		},
		CreditScoreHistory: creditScoreHistory,
	}
}

func saveCreditScoreToDB(debtwiseResponse *debtwise.CreditScore, totalAccounts int, debtwiseCustomerNumber int, userId string) (*dao.UserCreditScoreDao, error) {
	var increaseValue int
	if debtwiseResponse.Increase != nil {
		increaseValue = *debtwiseResponse.Increase
	} else {
		increaseValue = 0
	}

	encryptableCreditData := dao.EncryptableCreditScoreData{
		Score:                   debtwiseResponse.Score,
		Increase:                increaseValue,
		DebtwiseCustomerNumber:  debtwiseCustomerNumber,
		PaymentHistoryAmount:    float64(debtwiseResponse.PaymentHistory.Amount),
		PaymentHistoryFactor:    MapPaymentHistory(debtwiseResponse.PaymentHistory.Amount),
		CreditUtilizationAmount: float64(debtwiseResponse.CreditUtilization.Amount),
		CreditUtilizationFactor: MapCreditUtilization(debtwiseResponse.CreditUtilization.Amount),
		DerogatoryMarksAmount:   debtwiseResponse.DerogatoryMarks.Amount,
		DerogatoryMarksFactor:   MapDerogatoryMarks(debtwiseResponse.DerogatoryMarks.Amount),
		CreditAgeAmount:         float64(debtwiseResponse.CreditAge.Amount),
		CreditAgeFactor:         MapCreditAge(debtwiseResponse.CreditAge.Amount),
		CreditMixAmount:         debtwiseResponse.CreditMix.Amount,
		CreditMixFactor:         MapCreditMix(debtwiseResponse.CreditMix.Amount),
		NewCreditAmount:         debtwiseResponse.NewCredit.Amount,
		NewCreditFactor:         MapNewCredit(debtwiseResponse.NewCredit.Amount),
		TotalAccountsAmount:     totalAccounts,
		TotalAccountsFactor:     MapTotalAccounts(totalAccounts),
	}

	if err := validators.ValidateStruct(encryptableCreditData); err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("validation failed for encryptable data: %w", err))
	}

	jsonData, err := json.Marshal(encryptableCreditData)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to marshal to JSON: %w", err))
	}

	encryptedData, err := utils.EncryptKmsBinary(string(jsonData))
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to encrypt data: %w", err))
	}

	scoreDate, err := time.Parse("2006-01-02", debtwiseResponse.Date)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to parse credit score date %s: %w", debtwiseResponse.Date, err))
	}

	newScoreRecord := dao.UserCreditScoreDao{
		Id:                  uuid.New().String(),
		Date:                scoreDate,
		EncryptedCreditData: encryptedData,
		UserId:              userId,
	}
	err = db.DB.Select("id", "encrypted_credit_data", "date", "increase", "score", "debtwise_customer_number", "payment_history_amount", "payment_history_factor",
		"credit_utilization_amount", "credit_utilization_factor", "derogatory_marks_amount", "derogatory_marks_factor", "credit_age_amount",
		"credit_age_factor", "credit_mix_amount", "credit_mix_factor", "new_credit_amount", "new_credit_factor", "total_accounts_amount", "total_accounts_factor",
		"user_id").Create(&newScoreRecord).Error
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	return &newScoreRecord, nil
}

func retrieveTotalAccounts(debtwiseCustomerNumber int, user *dao.MasterUserRecordDao, debtwiseClient *debtwise.ClientWithResponses) (*int, error) {
	debtwiseReportResponse, err := debtwiseClient.EquifaxCreditReportsWithResponse(
		context.Background(),
		debtwise.UserIdParam(debtwiseCustomerNumber),
		&debtwise.EquifaxCreditReportsParams{},
	)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("error occurred while fetching credit report: %w", err))
	}

	var totalAccounts int
	if debtwiseReportResponse.JSON200 != nil {
		totalAccounts = len(debtwiseReportResponse.JSON200.Accounts)

		// NOTE: Retrieving a credit report is the final step of Equifax onboarding.
		// if this is the first time a user is retreiving the report we will mark their
		// status as complete otherwise we just collect the total number of accounts
		if user.DebtwiseOnboardingStatus != "complete" && user.DebtwiseOnboardingStatus == "inProgress" {
			updateResult := db.DB.Model(user).Update("debtwise_onboarding_status", "complete")
			if updateResult.Error != nil {
				return nil, errtrace.Wrap(fmt.Errorf("failed to update debtwise onboarding status to complete after fetching credit report for userId %s, err: %w", user.Id, updateResult.Error))
			}

			if updateResult.RowsAffected == 0 {
				return nil, errtrace.Wrap(fmt.Errorf("no rows updated, failed to update debtwise onboarding status to complete after fetching credit report for userId: %s", user.Id))
			}

			logging.Logger.Debug("Successfully updated user debtwise onboarding status as complete", "userId", user.Id)
		}
	}

	if debtwiseReportResponse.JSON422 != nil {
		return nil, errtrace.Wrap(fmt.Errorf("unprocessable entity error occurred while fetching credit report: %s", debtwiseReportResponse.JSON422.Error.Message))
	}

	return &totalAccounts, nil
}

type CreditScoreJobArgs struct {
	UserId string `json:"user_id"`
}

func (CreditScoreJobArgs) Kind() string {
	return "credit_score_job"
}

type CreditScoreWorker struct {
	river.WorkerDefaults[CreditScoreJobArgs]
}

func (CreditScoreJobArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: "debtwise",
	}
}

func RegisterCreditScoreWorker(workers *river.Workers) {
	river.AddWorker(workers, &CreditScoreWorker{})
}

func (w *CreditScoreWorker) Work(ctx context.Context, job *river.Job[CreditScoreJobArgs]) error {
	latestScore, err := FetchAndStoreCreditScore(ctx, job.Args.UserId)
	if err != nil {
		logging.Logger.Error("Failed to fetch and store credit score for user", "userId", job.Args.UserId, "error", err.Error())
	}

	if err := river.RecordOutput(ctx, latestScore); err != nil {
		logging.Logger.Error("Error recording credit score job output", "err", err)
		return fmt.Errorf("failed to record credit score job output: %w", err)
	}

	return nil
}

type CreditScoreEnqueueBatchJobArgs struct{}

func (CreditScoreEnqueueBatchJobArgs) Kind() string {
	return "credit_score_enqueue_batch_job"
}

func (CreditScoreEnqueueBatchJobArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: "debtwise",
	}
}

type CreditScoreEnqueueBatchWorker struct {
	river.WorkerDefaults[CreditScoreEnqueueBatchJobArgs]
	RiverClient *river.Client[*sql.Tx]
}

func (w *CreditScoreEnqueueBatchWorker) SetRiverClientForBatchWorker(client *river.Client[*sql.Tx]) {
	w.RiverClient = client
}

func RegisterCreditScoreEnqueueBatchWorker(workers *river.Workers, riverClient *river.Client[*sql.Tx]) *CreditScoreEnqueueBatchWorker {
	worker := &CreditScoreEnqueueBatchWorker{
		RiverClient: riverClient,
	}
	river.AddWorker(workers, worker)
	return worker
}

func (w *CreditScoreEnqueueBatchWorker) Work(ctx context.Context, job *river.Job[CreditScoreEnqueueBatchJobArgs]) error {
	return QueryOldRecordsAndEnqueue(w.RiverClient, ctx)
}

func QueryOldRecordsAndEnqueue(riverClient *river.Client[*sql.Tx], ctx context.Context) error {
	logger := logging.Logger

	cutoffDate := clock.Now().AddDate(0, 0, -30)

	logger.Debug("Executing query with cutoff date", "cutoffDate", cutoffDate)

	oldRecords, err := dao.UserCreditScoreDao{}.FindSufficientlyOldCreditScores(cutoffDate)
	if err != nil {
		logger.Error("Failed to query old credit score records", "error", err.Error())
		return errtrace.Wrap(err)
	}

	var batchParams []river.InsertManyParams
	for _, record := range oldRecords {
		batchParams = append(batchParams, river.InsertManyParams{
			Args: CreditScoreJobArgs{
				UserId: record.UserId,
			},
		})
	}

	if len(batchParams) == 0 {
		logger.Debug("No sufficiently old credit score records found to process")
		return nil
	}

	_, err = riverClient.InsertMany(ctx, batchParams)
	if err != nil {
		logger.Error("Failed to enqueue credit score jobs", "error", err.Error())
		return errtrace.Wrap(err)
	}

	logger.Info("Enqueued credit score job successfully", "userCount", len(oldRecords))
	return nil
}

type CreditScoreResponse struct {
	Date               string                   `json:"date" validate:"required"`
	Score              int                      `json:"score" validate:"required" mask:"true"`
	Increase           int                      `json:"increase" validate:"required" mask:"true"`
	PaymentHistory     CreditFactorFloat        `json:"paymentHistory" validate:"required" mask:"true"`
	CreditUtilization  CreditFactorFloat        `json:"creditUtilization" validate:"required" mask:"true"`
	DerogatoryMarks    CreditFactorInt          `json:"derogatoryMarks" validate:"required" mask:"true"`
	CreditAge          CreditAge                `json:"creditAge" validate:"required" mask:"true"`
	CreditMix          CreditFactorInt          `json:"creditMix" validate:"required" mask:"true"`
	NewCredit          CreditFactorInt          `json:"newCredit" validate:"required" mask:"true"`
	TotalAccounts      CreditFactorInt          `json:"totalAccounts" validate:"required" mask:"true"`
	CreditScoreHistory []dao.CreditScoreSummary `json:"creditScoreHistory" validate:"required" mask:"true"`
}

type CreditFactorFloat struct {
	Amount float64 `json:"amount" validate:"required" mask:"true"`
	Factor string  `json:"factor" enums:"Needs Work,Fair,Good,Excellent" validate:"required" mask:"true"`
}

type CreditFactorInt struct {
	Amount int    `json:"amount" validate:"required" mask:"true"`
	Factor string `json:"factor" enums:"Needs Work,Fair,Good,Excellent" validate:"required" mask:"true"`
}

type CreditAge struct {
	Amount string `json:"amount" validate:"required" mask:"true"`
	Factor string `json:"factor" enums:"Needs Work,Fair,Good,Excellent" validate:"required" mask:"true"`
}
