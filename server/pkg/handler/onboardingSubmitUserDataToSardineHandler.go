package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/sardine"
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"strconv"
	"time"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// @summary SubmitUserDataToSardine
// @description Starts a job for posting customer's data to sardine
// @tags onboarding
// @accept json
// @produce json
// @param requestSardineKycRequest body request.SardineKycRequest true "payload"
// @Param userId path string true "User ID"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} response.MaybeRiverJobResponse[response.SardinePayload]
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /onboarding/customer/kyc [put]
// @router /onboarding/customer/{userId}/kyc [put]
func (h *Handler) SubmitUserDataToSardine(c echo.Context) error {
	userId := c.Param("userId")

	if userId == "" {
		cc, ok := c.(*security.OnboardingUserContext)
		if !ok {
			return response.ErrorResponse{ErrorCode: constant.UNAUTHORIZED_ACCESS_ERROR, Message: constant.UNAUTHORIZED_ACCESS_ERROR_MSG, StatusCode: http.StatusUnauthorized, LogMessage: "Invalid type of custom context", MaybeInnerError: errtrace.New("")}
		}

		userId = cc.UserId
	}

	logger := logging.GetEchoContextLogger(c)

	request := new(request.SardineKycRequest)

	if err := c.Bind(request); err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(request); err != nil {
		return err
	}

	user, errResponse := dao.RequireUserWithState(
		userId, constant.ADDRESS_CONFIRMED,
	)

	if errResponse != nil {
		return errResponse
	}

	// Check in the Ledger for an existing SSN before beginning KYC
	SSNExist, err := CheckDuplicateSSN(request.SSN)
	if err != nil {
		logger.Error("Error while checking duplicate ssn", "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Error while checking duplicate ssn: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	// If SSN already exists
	if SSNExist {
		logger.Error("SSN already exists in ledger")
		return response.BadRequestErrors{
			Errors: []response.BadRequestError{
				{
					FieldName: "ssn",
					Error:     constant.DUPLICATE_SSN_MSG,
				},
			},
		}
	}

	// Store SSN in encrypted form in River job args
	encryptedSSN, err := utils.EncryptKms(request.SSN)
	if err != nil {
		logging.Logger.Error("Failed to encrypt ssn", "err", err)
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Failed to encrypt ssn: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	ctx := context.Background()
	jobInsertResult, err := h.RiverClient.Insert(ctx, SardineJobArgs{
		UserId:     user.Id,
		SessionKey: request.SardineSessionKey,
		SSN:        encryptedSSN,
	}, nil)
	if err != nil {
		logger.Error("Failed to enqueue sardine job", "error", err)
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Failed to enqueue sardine job: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	} else {
		logger.Debug("Enqueued sardine job successfully")
	}

	jobId := jobInsertResult.Job.ID
	var jobState string

	// Poll job status up to 3 times, every 500ms, include payload in response if job is complete
	for i := 0; i < 3; i++ {
		job, err := h.RiverClient.JobGet(ctx, jobId)
		if err != nil {
			logger.Error("Failed to fetch sardine job status", "jobId", jobId, "error", err.Error())
			break
		}

		jobState = string(job.State)

		if jobState == "completed" && job.Output() != nil {

			var output SardineJobResult
			if err := json.Unmarshal(job.Output(), &output); err == nil {
				response := handleSardineResponse(job, output, logger)
				if response != nil {
					return c.JSON(http.StatusOK, response)
				}
			}
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	response := response.MaybeRiverJobResponse[response.SardinePayload]{
		JobId:   int(jobInsertResult.Job.ID),
		State:   jobState,
		Payload: nil,
	}

	return c.JSON(http.StatusOK, response)
}

// @summary GetSardineJobStatus
// @description Get sardine Job Status
// @tags onboarding
// @produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Param jobId path int true "Jod id to get sardine job status"
// @Success 200 {object} response.MaybeRiverJobResponse[response.SardinePayload]
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /onboarding/customer/kyc/{jobId} [get]
func (h *Handler) GetSardineJobStatus(c echo.Context) error {
	cc, ok := c.(*security.OnboardingUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	logger := logging.GetEchoContextLogger(c)
	userId := cc.UserId

	jobIdStr := c.Param("jobId")
	jobId, err := strconv.Atoi(jobIdStr)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	ctx := context.Background()
	job, err := h.RiverClient.JobGet(ctx, int64(jobId))
	if errors.Is(err, river.ErrNotFound) {
		logger.Error("Failed to get sardine job status", "jobId", jobId, "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, MaybeInnerError: errtrace.Wrap(err)}
	}

	if err != nil {
		logger.Error("Failed to get sardine job status", "jobId", jobId, "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Failed to get sardine job status: jobId: %d, error: %s", jobId, err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	if job.Kind != (SardineJobArgs{}).Kind() {
		logger.Error("Job is not sardineJob", "jobId", jobId, "kind", job.Kind)
		return response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, MaybeInnerError: errtrace.New("")}
	}
	var args SardineJobArgs
	if err := json.Unmarshal(job.EncodedArgs, &args); err != nil {
		logger.Error("Job args failed to Unmarshal", "jobId", jobId, "err", err)
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Job args failed to Unmarshal: jobId: %d, error: %s", jobId, err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	if args.UserId != userId {
		logger.Error("Job userId does not match requesting user", "jobId", jobId, "userId", args.UserId)
		return response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, MaybeInnerError: errtrace.New("")}
	}

	jobState := string(job.State)

	if jobState == "completed" && job.Output() != nil {
		var output SardineJobResult
		if err := json.Unmarshal(job.Output(), &output); err == nil {
			response := handleSardineResponse(job, output, logger)
			if response != nil {
				return c.JSON(http.StatusOK, response)
			}
		}
	}

	sardineJobResponse := response.MaybeRiverJobResponse[response.SardinePayload]{
		JobId:   int(job.ID),
		State:   string(job.State),
		Payload: nil,
	}

	return cc.JSON(http.StatusOK, sardineJobResponse)
}

func AddCustomer(user dao.MasterUserRecordDao, ssn string) (string, error) {
	var ledgerPassword string
	encrypedDBLedgerPassword := user.KmsEncryptedLedgerPassword
	if len(encrypedDBLedgerPassword) != 0 {
		var passwordErr error
		ledgerPassword, passwordErr = utils.DecryptKmsBinary(encrypedDBLedgerPassword)
		if passwordErr != nil {
			return "", errtrace.Wrap(fmt.Errorf("error decrypting stored ledgerPassword: %v", passwordErr.Error()))
		}
	} else {
		var passwordErr error
		ledgerPassword, passwordErr = security.GenerateLedgerPassword()
		if passwordErr == nil {
			passwordErr = setLedgerPasswordForUser(ledgerPassword, user)
		}
		if passwordErr != nil {
			return "", errtrace.Wrap(fmt.Errorf("error generating ledger password %s", passwordErr.Error()))
		}
	}

	// Remove country code from mobileNumber as ledger does not supports it.
	mobileNo, err := utils.RemoveCountryCodeFromMobileNumber(user.MobileNo)
	if err != nil {
		return "", errtrace.Wrap(fmt.Errorf("error while validating and converting mobile number: %s", err.Error()))
	}
	user.MobileNo = mobileNo
	customerNumber, addCustomerErr := ledger.AddCustomer(user.Id, buildAddCustomerPayload(user, ledgerPassword, ssn))
	if addCustomerErr != nil {
		return "", errtrace.Wrap(fmt.Errorf("error adding customer to ledger %s", addCustomerErr.Error()))
	}

	updateCustomerNumberResult := db.DB.Model(&user).Update("ledger_customer_number", customerNumber)
	if updateCustomerNumberResult.Error != nil {
		return "", errtrace.Wrap(fmt.Errorf("error saving user's ledger customerNumber %s", updateCustomerNumberResult.Error.Error()))
	}
	if updateCustomerNumberResult.RowsAffected == 0 {
		return customerNumber, errtrace.Wrap(fmt.Errorf("the user record not updated with the ledgerCustomerNumber"))
	}
	return customerNumber, nil
}

func createSardineRequestData(user dao.MasterUserRecordDao, ssn string, sardineSessionKey string) sardine.PostCustomerInformationJSONRequestBody {
	dob := user.DOB.Format("2006-01-02")
	countryCode := "US"
	isPhoneVerified := true
	isEmailVerified := false
	createdAtMillis := user.CreatedAt.UnixMilli()
	sessionKey := sardineSessionKey
	flowName := "onboarding"
	flowType := sardine.FlowType("onboarding")
	checkpoints := []sardine.PostCustomerInformationJSONBodyCheckpoints{
		"customer",
	}

	// var requestBody sardine.PostCustomerInformationJSONRequestBody
	requestBody := sardine.PostCustomerInformationJSONRequestBody{
		Flow: sardine.Flow{
			Name: &flowName,
			Type: &flowType,
		},

		SessionKey: sessionKey,
		Customer: sardine.Customer{
			Id:              user.Id,
			CreatedAtMillis: &createdAtMillis,
			FirstName:       &user.FirstName,
			LastName:        &user.LastName,
			DateOfBirth:     &dob,
			TaxId:           &ssn,
			Address: &sardine.Address{
				Street1:     &user.StreetAddress,
				Street2:     &user.ApartmentNo,
				CountryCode: &countryCode,
				City:        &user.City,
				RegionCode:  &user.State,
				PostalCode:  &user.ZipCode,
			},
			Phone:           &user.MobileNo,
			EmailAddress:    &user.Email,
			IsPhoneVerified: &isPhoneVerified,
			IsEmailVerified: &isEmailVerified,
		},
		Checkpoints: &checkpoints,
	}
	return requestBody
}

func buildAddCustomerPayload(user dao.MasterUserRecordDao, password string, ssn string) ledger.AddCustomerData {
	return ledger.AddCustomerData{
		DOB:          user.DOB,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		PhoneNumber:  user.MobileNo,
		Email:        user.Email,
		AddressLine1: user.StreetAddress,
		AddressLine2: user.ApartmentNo,
		City:         user.City,
		State:        user.State,
		ZIP:          user.ZipCode,
		UserName:     user.Email,
		Password:     password,
		SSN:          ssn,
	}
}

func insertSessionKeyInDB(sessionKey, userId string) error {
	sardineData := dao.SardineDataDao{
		SessionKey:         sessionKey,
		SessionKeyExpiryDT: clock.Now().Add(30 * time.Minute),
		UserId:             userId,
	}
	result := db.DB.Create(&sardineData)
	// Handle other DB error
	if result.Error != nil {
		return errtrace.Wrap(fmt.Errorf("error occurred while creating record in sardine_data table: %w", result.Error))
	}
	return nil
}

func updateSardineResponseInDB(response, userId string) error {
	updateData := dao.SardineDataDao{
		SardineResponse: response,
	}
	result := db.DB.Model(&dao.SardineDataDao{}).Where("user_id=?", userId).Updates(updateData)
	// Handle DB errors
	if result.Error != nil {
		return errtrace.Wrap(fmt.Errorf("error while updating kyc response: %w", result.Error))
	} else if result.RowsAffected <= 0 {
		return errtrace.Wrap(errors.New("no record found"))
	}

	return nil
}

func updateKycStatusInDB(kycStatus, userId string) error {
	result := db.DB.Model(&dao.MasterUserRecordDao{}).Where("id=?", userId).Update("user_status", kycStatus)
	// Handle DB errors
	if result.Error != nil {
		return errtrace.Wrap(fmt.Errorf("error while updating user's status: %w", result.Error))
	} else if result.RowsAffected <= 0 {
		return errtrace.Wrap(errors.New("no record found"))
	}
	return nil
}

func CheckDuplicateSSN(ssn string) (bool, error) {
	payload := ledger.BuildGetCustomerBySSNPayload(ssn)

	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDLedgerApiClient(config.Config.Ledger, ledgerParamsBuilder)

	responseData, err := ledgerClient.GetCustomer(*payload)
	if err != nil {
		return false, errtrace.Wrap(fmt.Errorf("error while calling GetCustomer: %s", err.Error()))
	}

	// In the response error object contains a `Code` field; e.g., `NOT_FOUND_CUSTOMER` in case of user not found.
	if responseData.Error != nil && responseData.Error.Code == "NOT_FOUND_CUSTOMER" {
		return false, nil
	}

	if responseData.Result != nil && responseData.Result.Identification != nil {
		return true, nil
	}

	// Ledger returns statusCode 200(OK) but received unexpected response
	return false, errtrace.Wrap(errors.New("ledger returned unexpected response"))
}

func handleSardineResponse(job *rivertype.JobRow, output SardineJobResult, logger *slog.Logger) *response.MaybeRiverJobResponse[response.SardinePayload] {
	payload := &response.SardinePayload{}
	switch {
	case output.JSON200 != nil:
		payload.Result = &response.KycResult{
			UserId:    output.JSON200.UserId,
			KycStatus: output.JSON200.KycStatus,
		}

	case output.JSON400 != nil:
		logger.Error("Sardine returned error", "error", output.JSON400.ErrorMessage)
		payload.Error = &response.KycError{
			Message: constant.SARDINE_RETRY_ERROR_MSG,
		}

	case output.JSON401 != nil:
		logger.Error("Middleware is misconfigured for sardine", "error", output.JSON401.ErrorMessage)
		payload.Error = &response.KycError{
			Message: constant.SARDINE_RETRY_ERROR_MSG,
		}

	case output.JSON422 != nil:
		logger.Error("Sardine returned error", "error", output.JSON422.ErrorMessage)
		payload.Error = &response.KycError{
			Message: constant.SARDINE_RETRY_ERROR_MSG,
		}

	default:
		logger.Error("Unexpected Sardine API response", "jobId", int(job.ID))
		payload.Error = &response.KycError{
			Message: constant.SARDINE_RETRY_ERROR_MSG,
		}
	}
	return &response.MaybeRiverJobResponse[response.SardinePayload]{
		JobId:   int(job.ID),
		State:   string(job.State),
		Payload: payload,
	}
}

type SardineJobResult struct {
	JSON200 *Sardine200Response   `json:"json200,omitempty"`
	JSON400 *SardineErrorResponse `json:"json400,omitempty"`
	JSON401 *SardineErrorResponse `json:"json401,omitempty"`
	JSON422 *SardineErrorResponse `json:"json422,omitempty"`
}

type Sardine200Response struct {
	UserId    string `json:"userId"`
	KycStatus string `json:"kycStatus"`
}

type SardineErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`
}

type SardineJobArgs struct {
	UserId     string `json:"userId"`
	SessionKey string `json:"sessionKey"`
	SSN        string `json:"ssn"`
}

func (SardineJobArgs) Kind() string {
	return "sardine_job"
}

type SardineWorker struct {
	river.WorkerDefaults[SardineJobArgs]
}

func RegisterSardineWorker(workers *river.Workers) {
	river.AddWorker(workers, &SardineWorker{})
}

// This job(Sardine call + ledger call) may exceed 1.5s (usually 2–2.5s). We need to call GetSardineJobStatus API to fetch the result.
func (w *SardineWorker) Work(ctx context.Context, job *river.Job[SardineJobArgs]) error {
	var result SardineJobResult

	user, err := dao.MasterUserRecordDao{}.FindOneByUserId(job.Args.UserId)
	if err != nil {
		return fmt.Errorf("failed to get user record: %w", err)
	}

	// Decrypt SSN
	decryptedSSN, err := utils.DecryptKms(job.Args.SSN)
	if err != nil {
		logging.Logger.Error("Failed to decrypt ssn", "error", err.Error())
		return err
	}

	switch user.UserStatus {
	case constant.ADDRESS_CONFIRMED:
		return handleSardineApiCall(ctx, job, user, decryptedSSN)
	case constant.KYC_PASS:
		if user.LedgerCustomerNumber == "" {
			logging.Logger.Info("KYC already passed, skipping Sardine call", "userId", user.Id)
			customerNumber, err := AddCustomer(*user, decryptedSSN)
			if err != nil {
				logging.Logger.Error("Error in addCustomer", "error", err.Error())
				return err
			}
			logging.Logger.Info("User added to the ledger", "customerNo", customerNumber)
		}

		result = SardineJobResult{
			JSON200: &Sardine200Response{
				KycStatus: user.UserStatus,
				UserId:    job.Args.UserId,
			},
		}
		if err := river.RecordOutput(ctx, result); err != nil {
			logging.Logger.Error("Error recording sardine job output", "err", err)
			return fmt.Errorf("failed to record job output: %w", err)
		}
		return nil
		// KYC API was successful and status is KYC_FAIL, resolving the job without retry
	case constant.KYC_FAIL:
		result := SardineJobResult{
			JSON200: &Sardine200Response{
				KycStatus: user.UserStatus,
				UserId:    job.Args.UserId,
			},
		}
		if err := river.RecordOutput(ctx, result); err != nil {
			logging.Logger.Error("Error recording sardine job output", "err", err)
			return fmt.Errorf("failed to record job output: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("invalid user status: %s", user.UserStatus)
	}
}

func handleSardineApiCall(ctx context.Context, job *river.Job[SardineJobArgs], user *dao.MasterUserRecordDao, decryptedSSN string) error {
	logging.Logger.Info("Calling Sardine API", "userId", user.Id)

	sardineResponse, err := CallSardineAPI(*user, decryptedSSN, job.Args.SessionKey)
	if err != nil {
		logging.Logger.Error("Failed to call sardine API", "userId", job.Args.UserId, "error", err.Error())
		return fmt.Errorf("failed to call sardine API: %w", err)
	}

	var result SardineJobResult

	switch {
	case sardineResponse.JSON200 != nil:

		jsonData, err := json.Marshal(sardineResponse.JSON200)
		if err != nil {
			logging.Logger.Error("Error occurred while marshalling JSON200", "error", err.Error())
			return err

		}

		err = updateSardineResponseInDB(string(jsonData), job.Args.UserId)
		if err != nil {
			logging.Logger.Error(err.Error())
			return err
		}

		var kycStatus string
		if *sardineResponse.JSON200.Level == "low" {
			kycStatus = constant.KYC_PASS
		} else {
			kycStatus = constant.KYC_FAIL
		}

		// Update KYC status
		if err := updateKycStatusInDB(kycStatus, job.Args.UserId); err != nil {
			logging.Logger.Error(err.Error())
			return err
		}

		// If KYC is passed, then add the customer to the ledger
		if kycStatus == constant.KYC_PASS && user.LedgerCustomerNumber == "" {
			customerNumber, err := AddCustomer(*user, decryptedSSN)
			if err != nil {
				logging.Logger.Error("Error in addCustomer", "error", err.Error())
				return err
			}
			logging.Logger.Info("User added to the ledger", "customerNo", customerNumber)

		}
		if kycStatus == constant.KYC_FAIL {
			logging.Logger.Info("Reporting feedback to sardine for KYC fail")
			if err := reportFeedbackToSardine(*user, logging.Logger, job.Args.SessionKey, sardine.FeedbackStatusDeclined); err != nil {
				logging.Logger.Warn("Failed to report feedback to Sardine", "error", err.Error())
			}
		}
		result = SardineJobResult{
			JSON200: &Sardine200Response{
				KycStatus: kycStatus,
				UserId:    job.Args.UserId,
			},
		}

	case sardineResponse.JSON400 != nil:
		result = SardineJobResult{
			JSON400: &SardineErrorResponse{
				ErrorMessage: *sardineResponse.JSON400.Message,
			},
		}

	case sardineResponse.JSON401 != nil:
		result = SardineJobResult{
			JSON401: &SardineErrorResponse{
				ErrorMessage: *sardineResponse.JSON401.Reason,
			},
		}

	case sardineResponse.JSON422 != nil:
		result = SardineJobResult{
			JSON422: &SardineErrorResponse{
				ErrorMessage: *sardineResponse.JSON422.Message,
			},
		}

	default:
		logging.Logger.Error("Received unexpected error response from sardine API")
		return response.ErrorResponse{
			ErrorCode:  constant.INTERNAL_SERVER_ERROR,
			StatusCode: http.StatusInternalServerError,
			LogMessage: "Received unexpected error response from sardine API",
		}
	}

	if err := river.RecordOutput(ctx, result); err != nil {
		logging.Logger.Error("Error recording sardine job output", "err", err)
		return errtrace.Wrap(fmt.Errorf("failed to record job output: %w", err))
	}
	// We are not retrying the job in case of 4xx from sardine
	return nil
}

func CallSardineAPI(user dao.MasterUserRecordDao, ssn string, sessionKey string) (*sardine.PostCustomerInformationResponse, error) {
	// Create sardine client
	client, err := utils.NewSardineClient(config.Config.Sardine)
	if err != nil {
		logging.Logger.Error("Error occurred while creating sardine client", "error", err.Error())
		return nil, errtrace.Wrap(err)
	}

	requestBody := createSardineRequestData(user, ssn, sessionKey)

	// Create record in sardine_data table
	err = insertSessionKeyInDB(requestBody.SessionKey, user.Id)
	if err != nil {
		logging.Logger.Error("Failed to insert session key in DB", "error", err.Error())
		return nil, errtrace.Wrap(err)
	}

	// Sardine request body
	jsonRequestBody, err := json.MarshalIndent(requestBody, "", "  ")
	if err != nil {
		logging.Logger.Error("Error occurred while marshaling requestBody", "error", err.Error())
		return nil, errtrace.Wrap(err)
	}

	logging.Logger.Debug("Outgoing sardine request", "sardineRequestBody", string(jsonRequestBody))

	sardineResponse, err := client.PostCustomerInformationWithResponse(context.Background(), requestBody)
	if err != nil {
		logging.Logger.Error("Error occurred while calling sardine API", "error", err.Error())
		return nil, errtrace.Wrap(err)
	}

	// Sardine response body
	var res bytes.Buffer
	if err := json.Indent(&res, sardineResponse.Body, "", "  "); err != nil {
		logging.Logger.Error("Error occurred while formatting sardine response body", "error", err.Error())
		return nil, errtrace.Wrap(err)
	}
	logging.Logger.Info("Received from Sardine", "statusCode", sardineResponse.StatusCode(), "response", res.String())
	return sardineResponse, nil
}
