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

type BuildActivateCardRequest struct {
	CVV string `json:"cvv" validate:"required" mask:"true"`
	Pin string `json:"pin" validate:"required" mask:"true"`
}

type BuildActivateCardResponse struct {
	ValidateCvv  response.BuildPayloadResponse `json:"validateCvv" validate:"required"`
	ChangePin    response.BuildPayloadResponse `json:"changePin" validate:"required"`
	UpdateStatus response.BuildPayloadResponse `json:"updateStatus" validate:"required"`
}

// @Summary BuildActivateCardPayload
// @Description Generate ActivateCard payload.
// @Tags cards
// @Accept json
// @Produce json
// @Param payload body BuildActivateCardRequest true "payload with cvv"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} BuildActivateCardResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Router /account/cards/activate/build [post]
func BuildActivateCardPayload(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	requestData := new(BuildActivateCardRequest)
	err := c.Bind(requestData)
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

	validateCvv, errResponse := getValidateCvvPayload(userAccountCard.CardId, requestData.CVV, userId)
	if errResponse != nil {
		return errResponse
	}

	changePin, errResponse := getChangePinPayload(user.LedgerCustomerNumber, userAccountCard.AccountNumber, userAccountCard.CardId, requestData.Pin, userId)
	if errResponse != nil {
		return errResponse
	}

	updateStatus, errResponse := getUpdateCardStatusPayload(user, userAccountCard.AccountNumber, userAccountCard.CardId, ledger.ACTIVATE_CARD, requestData.CVV)
	if errResponse != nil {
		return errResponse
	}

	payloadResponse := BuildActivateCardResponse{
		ValidateCvv:  *validateCvv,
		ChangePin:    *changePin,
		UpdateStatus: *updateStatus,
	}

	logger.Info("ActivateCard payload created successfully for user", "userId", userId)
	return c.JSON(http.StatusOK, payloadResponse)
}

func getUpdateCardStatusPayload(user *dao.MasterUserRecordDao, accountNumber, cardId, action, cvv string) (*response.BuildPayloadResponse, *response.ErrorResponse) {
	userClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, nil)
	payload, err := userClient.BuildUpdateStatusRequest(user.LedgerCustomerNumber, cardId, accountNumber, action, cvv, utils.IsValidUnencrypted3DigitCVV(cvv))
	if err != nil {
		return nil, &response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("error occurred while creating updateStatus payload: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	payloadResponse, errResponse := dao.CreateSignablePayloadForUser(user.Id, payload)
	if errResponse != nil {
		return nil, errResponse
	}
	return payloadResponse, nil
}
