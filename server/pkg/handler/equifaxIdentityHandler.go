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

// @Summary EquifaxIdentityHandler
// @Description Retrieves identity report from equifax for confirmation from user
// @Tags credit score
// @Accept json
// @Produce json
// @Success 200 {object} EquifaxIdentityResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/debtwise/equifax/identity [post]
func (h *Handler) EquifaxIdentityHandler(c echo.Context) error {
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

	var postalCode string
	if h.Env != constant.PROD {
		postalCode = "30030"
		logger.Info("Overriding postal code for equifax identity call", "postalCode", postalCode)
	} else {
		postalCode = user.ZipCode
	}

	debtwiseRequest := debtwise.EquifaxAuthIdentityJSONBody{
		PostalCode: postalCode,
	}

	jsonBytes, err := json.Marshal(debtwiseRequest)
	if err != nil {
		logger.Error("Error occurred while marshalling request body", "error", err.Error(), "userId", userId)
		return response.InternalServerError(fmt.Sprintf("Error occurred while marshalling request body: error: %s, userId: %s", err.Error(), userId), errtrace.Wrap(err))
	}

	debtwiseResponse, err := debtwiseClient.EquifaxAuthIdentityWithBodyWithResponse(
		context.Background(),
		debtwise.UserIdParam(*user.DebtwiseCustomerNumber),
		&debtwise.EquifaxAuthIdentityParams{},
		"application/json",
		bytes.NewReader(jsonBytes),
	)
	if err != nil {
		logger.Error("Error occurred while getting equifax identity", "debtwiseError", err.Error(), "userId", userId)
		return response.InternalServerError(fmt.Sprintf("Error occurred while getting equifax identity: debtwiseError: %s, userId: %s", err.Error(), userId), errtrace.Wrap(err))
	}

	if debtwiseResponse.JSON200 != nil {
		logger.Info("Successfully retrieved Equifax identity for user", "userId", userId)

		return cc.JSON(http.StatusOK, EquifaxIdentityResponse{
			Addresses:   debtwiseResponse.JSON200.Addresses,
			DateOfBirth: debtwiseResponse.JSON200.DateOfBirth,
			FirstName:   debtwiseResponse.JSON200.FirstName,
			LastName:    debtwiseResponse.JSON200.LastName,
			MaskedSsn:   debtwiseResponse.JSON200.MaskedSsn,
		})
	}

	if debtwiseResponse.JSON422 != nil {
		logger.Error("Received a 422 response from debtwise", "errorMessage", debtwiseResponse.JSON422.Error.Message, "userId", userId)
		return response.InternalServerError("Received unprocessable entity error from Debtwise", errtrace.New(""))
	}

	logger.Error("Unexpected Debtwise response for identity",
		"status", debtwiseResponse.HTTPResponse.StatusCode,
		"body", debtwiseResponse.HTTPResponse.Body,
		"userId", userId,
	)
	return response.InternalServerError("Unexpected response from Debtwise", errtrace.New(""))
}

type EquifaxIdentityResponse struct {
	Addresses   *[]debtwise.EquifaxAddress `json:"addresses,omitempty" validate:"required" mask:"true"`
	DateOfBirth string                     `json:"dateOfBirth" validate:"required" mask:"true"`
	FirstName   string                     `json:"firstName" validate:"required"`
	LastName    string                     `json:"lastName" validate:"required" mask:"true"`
	MaskedSsn   *string                    `json:"maskedSsn,omitempty" mask:"true"`
}
