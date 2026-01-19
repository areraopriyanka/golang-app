package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/debtwise"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @Summary EquifaxValidateOtp
// @Description Validates user provided code for equifax issued OTP
// @Tags credit score
// @Accept json
// @Produce json
// @Param payload body EquifaxValidateOtpRequest true "EquifaxValidateOtpRequest payload"
// @Success 200 {object} EquifaxValidateOtpResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 422 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/debtwise/equifax/otp [post]
func EquifaxValidateOtp(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId

	logger := logging.GetEchoContextLogger(cc)

	debtwiseClient, err := debtwise.NewDebtwiseClient(config.Config.Debtwise, logger)
	if err != nil {
		logger.Error("Error occurred while creating debtwise client", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error occurred while creating debtwise client: error: %s", err.Error()), errtrace.Wrap(err))
	}

	requestData := new(EquifaxValidateOtpRequest)
	err = cc.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := cc.Validate(requestData); err != nil {
		return err
	}

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	if user.DebtwiseOnboardingStatus != "inProgress" {
		logger.Error("Attempting equifax onboarding otp without debtwise onboarding status inProgress", "userId", userId, "onboardingStatus", user.DebtwiseOnboardingStatus)
		return response.GenerateErrResponse(constant.INVALID_STATUS_ACTION, "Invalid debtwise onboarding status", "", http.StatusConflict, errtrace.New(""))
	}

	jsonBytes, err := json.Marshal(requestData)
	if err != nil {
		logger.Error("Error occurred while marshalling request body", "error", err.Error(), "userId", userId)
		return response.InternalServerError(fmt.Sprintf("Error occurred while marshalling request body: error: %s, userId: %s", err.Error(), userId), errtrace.Wrap(err))
	}

	debtwiseResponse, err := debtwiseClient.EquifaxAuthValidateOtpWithBodyWithResponse(
		context.Background(),
		debtwise.UserIdParam(*user.DebtwiseCustomerNumber),
		&debtwise.EquifaxAuthValidateOtpParams{},
		"application/json",
		bytes.NewReader(jsonBytes),
	)
	if err != nil {
		logger.Error("Error occurred while invoking equifax validate otp", "debtwiseError", err.Error(), "userId", userId)
		return response.InternalServerError(fmt.Sprintf("Error occurred while invoking equifax validate otp: debtwiseError: %s, userId: %s", err.Error(), userId), errtrace.Wrap(err))
	}

	if debtwiseResponse.JSON200 != nil {
		logger.Info("Successfully validated Equifax OTP for user", "userId", userId)
		return cc.JSON(http.StatusOK, EquifaxValidateOtpResponse{
			NextStep: string(debtwiseResponse.JSON200.Step),
		})
	} else if debtwiseResponse.JSON422 != nil {
		logger.Error("Incorrect input provided for OTP", "errorMessage", debtwiseResponse.JSON422.Error.Message, "userId", userId)
		return cc.JSON(http.StatusBadRequest, response.ErrorResponse{
			ErrorCode:       constant.INVALID_OTP,
			Message:         constant.INVALID_OTP_MSG,
			StatusCode:      http.StatusBadRequest,
			LogMessage:      "Incorrect OTP code",
			MaybeInnerError: errtrace.New(""),
		})
	} else {
		logger.Error("Unexpected Debtwise response for otp validation",
			"status", debtwiseResponse.HTTPResponse.StatusCode,
			"body", debtwiseResponse.HTTPResponse.Body,
			"userId", userId,
		)
		return response.InternalServerError("Unexpected Debtwise response for otp validation", errtrace.New(""))
	}
}

type EquifaxValidateOtpRequest struct {
	Code string `json:"code" validate:"required,min=4,max=8,numeric"`
}

type EquifaxValidateOtpResponse struct {
	NextStep string `json:"nextStep" validate:"required" enums:"consent"`
}
