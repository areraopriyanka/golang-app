package handler

import (
	"context"
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

// @Summary EquifaxConsentHandler
// @Description Indicates user's consent to Equifax T&C
// @Tags credit score
// @Accept json
// @Produce json
// @Success 200 {object} EquifaxConsentResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/debtwise/equifax/consent [post]
func EquifaxConsentHandler(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId

	logger := logging.GetEchoContextLogger(cc)

	debtwiseClient, err := debtwise.NewDebtwiseClient(config.Config.Debtwise, logger)
	if err != nil {
		logger.Error("Error occurred while creating debtwise client", "error", err.Error(), "userId", userId)
		return response.InternalServerError(fmt.Sprintf("Error occurred while creating debtwise client: error: %s userId: %s", err.Error(), userId), errtrace.Wrap(err))
	}

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	if user.DebtwiseOnboardingStatus != "inProgress" {
		logger.Error("Attempting equifax onboarding otp without debtwise onboarding status inProgress", "userId", userId, "onboardingStatus", user.DebtwiseOnboardingStatus)
		return response.GenerateErrResponse(constant.INVALID_STATUS_ACTION, "Invalid debtwise onboarding status", "", http.StatusConflict, errtrace.New(""))
	}

	debtwiseResponse, err := debtwiseClient.EquifaxAuthConsentWithResponse(
		context.Background(),
		debtwise.UserIdParam(*user.DebtwiseCustomerNumber),
		&debtwise.EquifaxAuthConsentParams{},
	)
	if err != nil {
		logger.Error("Error occurred while invoking equifax consent", "debtwiseError", err.Error(), "userId", userId)
		return response.InternalServerError(fmt.Sprintf("Error occurred while invoking equifax consent: debtwiseError: %s userId: %s", err.Error(), userId), errtrace.Wrap(err))
	}

	if debtwiseResponse.JSON200 != nil {
		logger.Info("Successfully completed equifax consent for user", "userId", userId)
		return cc.JSON(http.StatusOK, EquifaxValidateOtpResponse{
			NextStep: string(debtwiseResponse.JSON200.Step),
		})
	} else {
		logger.Error("Unexpected Debtwise response for consent",
			"status", debtwiseResponse.HTTPResponse.StatusCode,
			"body", debtwiseResponse.HTTPResponse.Body,
			"userId", userId,
		)
		return response.InternalServerError("Unexpected Debtwise response for consent", errtrace.New(""))
	}
}

type EquifaxConsentResponse struct {
	NextStep string `json:"nextStep" validate:"required" enums:"identity"`
}
