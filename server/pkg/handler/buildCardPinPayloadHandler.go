package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

type BuildCardPinPayloadRequest struct {
	NewPin string `json:"newPin" validate:"required"`
}

// @summary BuildCardPinPayload
// @description Builds and stores signable payload for setting pin
// @tags cards
// @accept json
// @produce json
// @param buildCardPinPayloadRequest body BuildCardPinPayloadRequest true "Request body to build card PIN payload"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} response.BuildPayloadResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/cards/pin/build [post]
func BuildCardPinPayload(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	logger := logging.GetEchoContextLogger(c)
	userId := cc.UserId

	var requestData BuildCardPinPayloadRequest

	if err := c.Bind(&requestData); err != nil {
		logging.Logger.Error("Invalid request", "error", err.Error())
		return response.BadRequestErrors{
			Errors: []response.BadRequestError{
				{Error: err.Error()},
			},
		}
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

	payloadResponse, errResponse := getChangePinPayload(user.LedgerCustomerNumber, userAccountCard.AccountNumber, userAccountCard.CardId, requestData.NewPin, userId)
	if errResponse != nil {
		return errResponse
	}
	logger.Debug("ChangePin payload created successfully for user", "userId", userId)
	return c.JSON(http.StatusOK, payloadResponse)
}

func getChangePinPayload(ledgerCustomerNumber string, ledgerInitialCheckingAccountNumber string, cardId string, pin string, userId string) (*response.BuildPayloadResponse, *response.ErrorResponse) {
	apiConfig := ledger.NewNetXDCardApiConfig(config.Config.Ledger)
	payload, err := apiConfig.BuildChangePinRequest(ledgerCustomerNumber, ledgerInitialCheckingAccountNumber, cardId, pin, utils.IsValidUnencryptedPIN(pin))
	if err != nil {
		return nil, &response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("error occurred while creating setPin payload: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	payloadResponse, errResponse := dao.CreateSignablePayloadForUser(userId, payload)
	if errResponse != nil {
		return nil, errResponse
	}
	return payloadResponse, nil
}
