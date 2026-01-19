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
	"strings"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @Summary EquifaxInvokeHandler
// @Description Invokes debtwise to begin equifax onboarding process for DW user
// @Tags credit score
// @Accept json
// @Produce json
// @Success 200 {object} EquifaxInvokeResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/debtwise/equifax/invoke [post]
func (h *Handler) EquifaxInvokeHandler(c echo.Context) error {
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

	deviceIp := cc.RealIP()
	if h.Env != constant.PROD && strings.Contains(user.FirstName, "DWFail") {
		deviceIp = "127.128.0.0"
		logger.Debug("Overriding device IP for equifax invoke call", "deviceIp", deviceIp)
	}

	requestData := debtwise.EquifaxAuthInvokeJSONBody{
		DeviceIp: deviceIp,
	}

	jsonBytes, err := json.Marshal(requestData)
	if err != nil {
		logger.Error("Error occurred while marshalling request body: " + err.Error())
		return response.InternalServerError(fmt.Sprintf("Error occurred while marshalling request body: %s", err.Error()), errtrace.Wrap(err))
	}

	if user.DebtwiseCustomerNumber == nil {
		logger.Error("Attempting to invoke equifax onboarding without debtwise customer number", "userId", userId)
		return response.GenerateErrResponse(constant.INVALID_STATUS_ACTION, "Invalid debtwise onboarding status", "", http.StatusConflict, errtrace.New(""))
	}

	if user.DebtwiseOnboardingStatus == "inProgress" {
		debtwiseAbortResponse, err := debtwiseClient.EquifaxAuthAbortWithResponse(
			context.Background(),
			debtwise.UserIdParam(*user.DebtwiseCustomerNumber),
			&debtwise.EquifaxAuthAbortParams{},
		)
		if err != nil {
			logger.Error("Error occurred while calling equifax abort", "debtwiseError", err.Error(), "userId", userId)
			return response.InternalServerError(fmt.Sprintf("Error occurred while calling equifax abort: debtwiseError: %s, userId: %s", err.Error(), userId), errtrace.Wrap(err))
		}

		if debtwiseAbortResponse.HTTPResponse.StatusCode != http.StatusOK {
			logger.Error("Error occurred while calling equifax abort", "userId", userId)
			return response.InternalServerError(fmt.Sprintf("Error occurred while calling equifax aborts. userId: %s", userId), errtrace.New(""))
		}

		if debtwiseAbortResponse.HTTPResponse.StatusCode == http.StatusOK {
			updateResult := db.DB.Model(user).Update("debtwise_onboarding_status", "uninitialized")
			if updateResult.Error != nil {
				logger.Error("Failed to update debtwise onboarding status to uninitialized during invoke", "dbError", updateResult.Error.Error(), "userId", userId)
				return response.InternalServerError(fmt.Sprintf("Failed to update debtwise onboarding status to uninitialized during invoke: dbError: %s, userId: %s", updateResult.Error.Error(), userId), errtrace.Wrap(updateResult.Error))
			}

			if updateResult.RowsAffected == 0 {
				logger.Error("Failed to update debtwise onboarding status to uninitialized during invoke", "userId", userId)
				return response.InternalServerError(fmt.Sprintf("Failed to update debtwise onboarding status to uninitialized during invoke: userId: %s", userId), errtrace.New(""))
			}

			user.DebtwiseOnboardingStatus = "uninitialized"
			logger.Info("Successfully aborted onboarding during Equifax invoke", "userId", userId)
		}
	}

	if user.DebtwiseOnboardingStatus == "complete" {
		logger.Error("Attempting to invoke equifax onboarding for already onboarded user", "userId", userId)
		return response.ErrorResponse{
			ErrorCode:       constant.INAPPROPRIATE_STATUS_ACTION,
			StatusCode:      http.StatusBadRequest,
			LogMessage:      fmt.Sprintf("Attempting to invoke equifax onboarding for already onboarded user: userId: %s", userId),
			MaybeInnerError: errtrace.New(""),
		}
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
			logger.Error("Error occurred while invoking equifax onboarding", "debtwiseError", err.Error(), "userId", userId)
			return response.InternalServerError(fmt.Sprintf("Error occurred while invoking equifax onboarding: debtwiseError: %s, userId: %s", err.Error(), userId), errtrace.Wrap(err))
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
			responsePayload := EquifaxInvokeResponse{
				NextStep: string(debtwiseResponse.JSON200.Step),
			}

			if responsePayload.NextStep == "validate_otp" && user.MobileNo != "" && len(user.MobileNo) >= 4 {
				masked := "XXX-XXX-" + (user.MobileNo)[len(user.MobileNo)-4:]
				responsePayload.MaskedMobileNo = &masked
			}

			logger.Info("Successfully invoked equifax onboarding for user", "userId", userId, "nextStep", string(debtwiseResponse.JSON200.Step))
			return cc.JSON(http.StatusOK, responsePayload)
		}

		if debtwiseResponse.JSON422 != nil {
			logger.Error("Received a 422 from debtwise", "errorMessage", debtwiseResponse.JSON422.Error.Message, "userId", userId)
			return response.InternalServerError("Received unprocessable entity error from Debtwise", errtrace.New(""))
		}

		logger.Error("Unexpected Debtwise response for invoke",
			"status", debtwiseResponse.HTTPResponse.StatusCode,
			"body", debtwiseResponse.HTTPResponse.Body,
			"userId", userId,
		)
		return response.InternalServerError("Unexpected response from Debtwise", errtrace.New(""))
	}

	return response.InternalServerError("Unexpected response from Debtwise", errtrace.New(""))
}

type EquifaxInvokeResponse struct {
	NextStep       string  `json:"nextStep" validate:"required" enums:"consent,validate_otp"`
	MaskedMobileNo *string `json:"maskedMobileNumber,omitempty" mask:"true"`
}
