package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/debtwise"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @Summary EquifaxOtpResend
// @Description Calls equifax abort and invoke to resend otp
// @Tags credit score
// @Accept json
// @Produce json
// @Success 200 "OK"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/debtwise/equifax/otp/resend [post]
func EquifaxOtpResend(c echo.Context) error {
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

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	if user.DebtwiseOnboardingStatus != "inProgress" {
		logger.Error("Attempting equifax onboarding otp without debtwise onboarding status inProgress", "userId", userId, "onboardingStatus", user.DebtwiseOnboardingStatus)
		return response.GenerateErrResponse(constant.INVALID_STATUS_ACTION, "Invalid debtwise onboarding status", "", http.StatusConflict, errtrace.New(""))
	}

	debtwiseAbortResponse, err := debtwiseClient.EquifaxAuthAbortWithResponse(
		context.Background(),
		debtwise.UserIdParam(*user.DebtwiseCustomerNumber),
		&debtwise.EquifaxAuthAbortParams{},
	)
	if err != nil {
		logger.Error("Error occurred while calling equifax abort for resend otp", "debtwiseError", err.Error(), "userId", userId)
		return response.InternalServerError(fmt.Sprintf("Error occurred while calling equifax abort for resend otp: debtwiseError: %s, userId: %s", err.Error(), userId), errtrace.Wrap(err))
	}

	if debtwiseAbortResponse.HTTPResponse.StatusCode != http.StatusOK {
		logger.Error("Error occurred while calling equifax abort for resend otp", "userId", userId)
		return response.InternalServerError(fmt.Sprintf("Recieved non-success response from equifax abort for resend otp: userId: %s", userId), errtrace.New(""))
	}

	if debtwiseAbortResponse.HTTPResponse.StatusCode == http.StatusOK {
		updateResult := db.DB.Model(user).Update("debtwise_onboarding_status", "uninitialized")
		if updateResult.Error != nil {
			logger.Error("Failed to update debtwise onboarding status to uninitialized during resend otp", "dbError", updateResult.Error.Error(), "userId", userId)
			return response.InternalServerError(fmt.Sprintf("Failed to update debtwise onboarding status to uninitialized during resend otp: dbError: %s, userId: %s", updateResult.Error.Error(), userId), errtrace.Wrap(updateResult.Error))
		}
		if updateResult.RowsAffected == 0 {
			logger.Error("Failed to update debtwise onboarding status to uninitialized during resend otp", "userId", userId)
			return response.InternalServerError(fmt.Sprintf("Failed to update debtwise onboarding status to uninitialized during resend otp: userId: %s", userId), errtrace.New(""))
		}
	}

	requestData := debtwise.EquifaxAuthInvokeJSONBody{
		DeviceIp: cc.RealIP(),
	}

	jsonBytes, err := json.Marshal(requestData)
	if err != nil {
		logger.Error("Error occurred while marshalling request body: " + err.Error())
		return response.InternalServerError(fmt.Sprintf("Error occurred while marshalling request body: %s", err.Error()), errtrace.Wrap(err))
	}

	if user.DebtwiseOnboardingStatus == "uninitialized" {
		debtwiseResponse, err := debtwiseClient.EquifaxAuthInvokeWithBodyWithResponse(
			context.Background(),
			debtwise.UserIdParam(*user.DebtwiseCustomerNumber),
			&debtwise.EquifaxAuthInvokeParams{},
			"application/json",
			bytes.NewReader(jsonBytes),
		)
		if err != nil {
			logger.Error("Error occurred while invoking equifax onboarding during resend otp", "debtwiseError", err.Error(), "userId", userId)
			return response.InternalServerError(fmt.Sprintf("Error occurred while invoking equifax onboarding during resend otp: debtwiseError: %s, userId: %s", err.Error(), userId), errtrace.Wrap(err))
		}

		if debtwiseResponse.JSON200 != nil {
			updateResult := db.DB.Model(user).Update("debtwise_onboarding_status", "inProgress")
			if updateResult.Error != nil {
				logger.Error("Failed to update debtwise onboarding status to inProgress  during invoke", "dbError", updateResult.Error.Error(), "userId", userId)
				return response.InternalServerError(fmt.Sprintf("Failed to update debtwise onboarding status to inProgress during invoke: dbError: %s, userId: %s", updateResult.Error.Error(), userId), errtrace.Wrap(updateResult.Error))
			}

			if updateResult.RowsAffected == 0 {
				logger.Error("Failed to update debtwise onboarding status to inProgress during invoke", "userId", userId)
				return response.InternalServerError(fmt.Sprintf("Failed to update debtwise onboarding status to inProgress during invoke: userId: %s", userId), errtrace.New(""))
			}

			if string(debtwiseResponse.JSON200.Step) != "validate_otp" {
				logger.Error("Received unexpected next step while resending equifax otp", "userId", userId)
				return response.InternalServerError(fmt.Sprintf("Received unexpected next step while resending equifax otp: userId: %s", userId), errtrace.New(""))
			}

			return cc.NoContent(http.StatusOK)
		}

		if debtwiseResponse.JSON422 != nil {
			logger.Error("Received a 422 from debtwise during otp resend", "errorMessage", debtwiseResponse.JSON422.Error.Message, "userId", userId)
			return response.InternalServerError("Received unprocessable entity error from Debtwise", errtrace.New(""))
		}

		logger.Error("Unexpected Debtwise response for invoke during resend otp",
			"status", debtwiseResponse.HTTPResponse.StatusCode,
			"body", debtwiseResponse.HTTPResponse.Body,
			"userId", userId,
		)
		return response.InternalServerError("Unexpected response from Debtwise", errtrace.New(""))
	}

	return response.InternalServerError("Unexpected response from Debtwise", errtrace.New(""))
}
