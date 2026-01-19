package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @Summary ReplaceCard
// @Description REPLACE - A replacement card is issued with a new Primary Account Number (PAN).
// @Tags cards
// @Produce json
// @Param payload body request.ReplaceCardRequest true "request.ReplaceCardRequest"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 "OK"
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 500 {object} response.ErrorResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 410 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Router /account/cards/replace [post]
func ReplaceCard(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	logger := logging.GetEchoContextLogger(c)

	requestData := new(request.ReplaceCardRequest)
	err := c.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	userId := cc.UserId

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	userPublicKey, errResponse := dao.RequireUserPublicKey(userId, cc.PublicKey)
	if errResponse != nil {
		return errResponse
	}

	payloadRecord, errResponse := dao.ConsumePayload(user.Id, requestData.PayloadId)
	if errResponse != nil {
		return errResponse
	}

	decryptedLedgerPassword, decryptedApiKey, err := utils.DecryptApiKeyAndLedgerPassword(user.LedgerPassword, user.KmsEncryptedLedgerPassword, userPublicKey.ApiKey, userPublicKey.KmsEncryptedApiKey, logger)
	if err != nil {
		logger.Error(err.Error())
		return response.InternalServerError(err.Error(), errtrace.Wrap(err))

	}

	userAccountCard, errResponse := dao.RequireActiveCardHolderForUser(userId)
	if errResponse != nil {
		logger.Error("Failed to get card holder", "error", errResponse.LogMessage)
		return response.InternalServerError(fmt.Sprintf("Failed to get card holder: error: %s", errResponse.LogMessage), errtrace.New(""))

	}

	if userAccountCard.IsReplaceLocked() {
		logger.Error("Card replacement is locked")
		return response.GenerateErrResponse(constant.INVALID_USER_STATE, "Card replacement is locked", "", http.StatusConflict, errtrace.New(""))
	}

	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, ledgerParamsBuilder)

	getCardRequest := ledgerClient.BuildGetCardDetailsRequest(user.LedgerCustomerNumber, userAccountCard.AccountNumber, userAccountCard.CardId)
	getCardResponse, err := ledgerClient.GetCardDetails(getCardRequest)
	if err != nil {
		logger.Error("Error from callLedgerGetCardDetails", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error from callLedgerGetCardDetails: error: %s", err.Error()), errtrace.Wrap(err))
	}

	if getCardResponse.Error != nil {
		logger.Error("The ledger responded with an error", "code", getCardResponse.Error.Code, "message", getCardResponse.Error.Message)
		return response.InternalServerError(fmt.Sprintf("The ledger responded with an error: %s", getCardResponse.Error.Message), errtrace.New(""))

	}

	if getCardResponse.Result == nil {
		logger.Error("The ledger responded with an empty result object", "response", getCardResponse)
		return response.InternalServerError("The ledger responded with an empty result object", errtrace.New(""))

	}

	if MapLedgerCardStatus(getCardResponse.Result.Card.CardStatus, logger) == "inactive" || MapLedgerCardStatus(getCardResponse.Result.Card.CardStatus, logger) == "cancelled" {
		logger.Error("Cannot replace card", "status", getCardResponse.Result.Card.CardStatus)
		return response.GenerateErrResponse(constant.INAPPROPRIATE_STATUS_ACTION, constant.INAPPROPRIATE_STATUS_ACTION_MSG, "", http.StatusConflict, errtrace.New(""))
	}

	userClient := ledger.CreateCardApiClient(userPublicKey.PublicKey, requestData.Signature, payloadRecord.Payload, user.Email, decryptedLedgerPassword, userPublicKey.KeyId, decryptedApiKey)

	var responseData ledger.NetXDApiResponse[ledger.ReplaceOrReissueCardResult]

	var request ledger.ReplaceOrReissueCardRequest
	err = json.Unmarshal([]byte(payloadRecord.Payload), &request)
	if err != nil {
		logger.Error("Error while unmarshaling payload", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error while unmarshaling payload: error: %s", err.Error()), errtrace.Wrap(err))
	}

	if request.StatusAction == ledger.REPLACE_CARD {
		userId := cc.UserId

		// Update card status to LOST_STOLEN before replacing the card
		_, err := UpdateStatus(logger, userId, ledger.REPORT_LOST_STOLEN)
		if err != nil {
			logger.Error("Error while updating card status to LOST_STOLEN", "error", err.Error())
			return err
		}
	}

	responseData, err = userClient.ReplaceOrReissueCard(request)
	if err != nil {
		logger.Error("Error from ReplaceOrReissueCard", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error from ReplaceOrReissueCard: error: %s", err.Error()), errtrace.Wrap(err))
	}

	if responseData.Error != nil {
		// ledger returns 1018 status code in case of Inappropriate status action
		if responseData.Error.Code == "1018" {
			logger.Error("Ledger returned error", "errorMessage", responseData.Error.Message)
			return response.GenerateErrResponse(constant.INAPPROPRIATE_STATUS_ACTION, constant.INAPPROPRIATE_STATUS_ACTION_MSG, "", http.StatusBadRequest, errtrace.New(""))
		}
		logger.Error("The ledger responded with an error", "code", responseData.Error.Code, "msg", responseData.Error.Message)
		return response.InternalServerError(fmt.Sprintf("The ledger responded with an error: code: %s, msg: %s", responseData.Error.Code, responseData.Error.Message), errtrace.New(""))
	}

	if responseData.Result.Api.Type == "REPLACE_CARD_ACK" {
		db.DB.Model(&userAccountCard).Updates(map[string]interface{}{
			"card_id":                   responseData.Result.NewCard.CardId,
			"is_replace":                responseData.Result.Card.IsReplace,
			"is_reissue":                responseData.Result.Card.IsReIssue,
			"previous_card_id":          userAccountCard.CardId,
			"previous_card_mask_number": userAccountCard.CardMaskNumber,
			// TODO: Remove this once the Ledger's GetCardDetails API issue is fixed.
			// Update the card expiration date when the card is reissued.
			"card_expiration_date": responseData.Result.NewCard.CardExpiryDate,
		})

		logger.Info("Card replaced or reissued successfully.", "cardId", responseData.Result.Card.CardId)
		return c.NoContent(http.StatusOK)
	}

	// ledger returned unexpected response
	logger.Error("ledger returned unexpected response")
	return response.InternalServerError("ledger returned unexpected response", errtrace.New(""))
}
