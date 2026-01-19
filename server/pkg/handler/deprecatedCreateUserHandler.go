package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// @summary CreateOnboardingUser
// @description Creates a user in middleware DB and returns a session id
// @tags onboarding
// @accept json
// @produce json
// @param createOnboardingUserRequest body request.CreateOrUpdateOnboardingUserRequest true "CreateOnboardingUser payload"
// @Success 201 {object} response.CreateOnboardingUserResponse
// @failure 500 {object} response.ErrorResponse
// @failure 400 {object} response.ErrorResponse
// @router /onboarding/customer [put]
func CreateUser(c echo.Context) error {
	var requestData request.CreateOrUpdateOnboardingUserRequest

	logger := logging.GetEchoContextLogger(c)
	err := json.NewDecoder(c.Request().Body).Decode(&requestData)
	if err != nil {
		logger.Error(constant.INTERNAL_SERVER_ERROR, "error", err.Error())
		return response.InternalServerError(err.Error(), errtrace.Wrap(err))
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	// Ensure the email is not already registered, either in middleware or in the ledger
	isEmailExists, err := IsUserEmailRegistered(c, requestData.Email)
	if err != nil {
		logger.Error("isUserEmailRegistered failed", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("isUserEmailRegistered failed: error: %s", err.Error()), errtrace.Wrap(err))
	}
	if isEmailExists {
		logger.Error(constant.EMAIL_ALREADY_REGISTERED_MSG)
		return response.ErrorResponse{ErrorCode: constant.EMAIL_ALREADY_REGISTERED, StatusCode: http.StatusInternalServerError, LogMessage: constant.EMAIL_ALREADY_REGISTERED_MSG, MaybeInnerError: errtrace.New("")}
	}
	// Generate long lived non-guessable id
	sessionId := uuid.New().String()

	userRecord := dao.MasterUserRecordDao{
		Id:         sessionId,
		FirstName:  requestData.FirstName,
		LastName:   requestData.LastName,
		Suffix:     requestData.Suffix,
		Email:      requestData.Email,
		UserStatus: constant.USER_CREATED,
	}

	result := db.DB.Select("id", "first_name", "last_name", "email", "suffix", "user_status", "created_at").Create(&userRecord)
	if result.Error != nil {
		logger.Error(constant.INTERNAL_SERVER_ERROR, "error", result.Error)
		return response.InternalServerError(result.Error.Error(), errtrace.Wrap(result.Error))
	}

	response := response.CreateOnboardingUserResponse{
		Id: sessionId,
	}
	return c.JSON(http.StatusCreated, response)
}

func IsUserEmailRegistered(c echo.Context, email string) (bool, error) {
	// check if user already exists at middleware DB
	user, err := dao.MasterUserRecordDao{}.FindUserByEmail(email)
	if err != nil {
		return false, errtrace.Wrap(err)
	}
	// If user record present in middleware DB
	if user != nil {
		return true, nil
	}

	// check if user already exists at ledger
	var emailPayload request.UserAlreadyRegisteredRequest
	emailPayload.Contact.Email = email

	emailPayloadData, err := json.Marshal(emailPayload)
	if err != nil {
		return false, errtrace.Wrap(fmt.Errorf("an error occurred while marshaling payload data: %w", err))
	}
	statusCode, respBody, err := callIsUserAlreadyRegisteredLedgerAPI(c, emailPayloadData)
	// if any error while calling ledger API
	if err != nil {
		return false, errtrace.Wrap(fmt.Errorf("callIsUserAlreadyRegisteredLedgerAPI failed: %w", err))
	}

	isEmailExists, error := checkIsUserAlreadyRegisteredResult(statusCode, respBody)
	if error != nil {
		return false, errtrace.Wrap(fmt.Errorf("checkIsUserAlreadyRegisteredResult failed: %v", error))
	}

	return isEmailExists, nil
}

func callIsUserAlreadyRegisteredLedgerAPI(c echo.Context, payload []byte) (int, json.RawMessage, error) {
	logging.Logger.Info("Inside callIsUserAlreadyRegisteredLedgerAPI")
	requestBody := &ledger.Request{
		Method: constant.GET_CUSTOMER_METHOD,
		Id:     constant.LEDGER_REQUEST_ID,
	}
	requestBody.Params.Payload = payload
	err := requestBody.SignRequestWithMiddlewareKey()
	if err != nil {
		return 0, nil, errtrace.Wrap(fmt.Errorf("an error occurred while signing unsigned mobile request payload: %w", err))
	}

	statusCode, respBody, err := ledger.CallLedgerAPIAndGetRawResponse(c, requestBody)
	if err != nil {
		return 0, nil, errtrace.Wrap(fmt.Errorf("an error occurred while calling Ledger API: %w", err))
	}
	return statusCode, respBody, nil
}

func checkIsUserAlreadyRegisteredResult(statusCode int, respBody []byte) (bool, error) {
	var errResp ledger.LedgerErrorResponse
	var resultResp ledger.UserAlreadyRegisteredResponse
	if statusCode == http.StatusOK {
		// Unmarshal Raw json to ResultResponse Struct
		err := json.Unmarshal(respBody, &resultResp)
		if err != nil {
			return false, errtrace.Wrap(fmt.Errorf("an error occurred while Unmarshaling response: %w", err))
		}
		// if we got result value then email or mobile number already registered
		if resultResp.Result.ID != "" {
			return true, nil
		}
		// Unmarshal Raw json to ErrResponse Struct
		err = json.Unmarshal(respBody, &errResp)
		if err != nil {
			return false, errtrace.Wrap(fmt.Errorf("an error occurred while Unmarshaling error response: %w", err))
		}
		// if email address or mobile number not registered
		if errResp.Error.Code == "NOT_FOUND_CUSTOMER" {
			return false, nil
		}
		// If email address or mobile number is already registered
		if errResp.Error.Code == "MULTI_CUSTOMER" {
			return true, nil
		}

		// If statuscode is 200 but received unexpected error then return error. Response already logged in CallLedgerAPIWithUrlAndGetRawResponse
		return false, errtrace.Wrap(fmt.Errorf("statuscode is 200 but received unexpected response from ledger"))

	}
	// if status code is not 200 then return error. Response already logged in CallLedgerAPIWithUrlAndGetRawResponse
	return false, errtrace.Wrap(fmt.Errorf("received error response from ledger"))
}
