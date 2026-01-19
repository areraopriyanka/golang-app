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
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

type PayloadRequest struct {
	Signature string `json:"signature" validate:"required"`
	PayloadId string `json:"payloadId" validate:"required"`
}

type ActivateCardRequest struct {
	ValidateCvv  PayloadRequest `json:"validateCvv" validate:"required"`
	SetCardPin   PayloadRequest `json:"setCardPin" validate:"required"`
	UpdateStatus PayloadRequest `json:"updateStatus" validate:"required"`
}

// @Summary ActivateCard
// @Description Activate card.
// @Tags cards
// @Produce json
// @Param payload body ActivateCardRequest true "ActivateCardRequest"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} response.GetCardResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 410 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/cards/activate [post]
func ActivateCard(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	requestData := new(ActivateCardRequest)
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
	userPublicKey, errResponse := dao.RequireUserPublicKey(userId, cc.PublicKey)
	if errResponse != nil {
		return errResponse
	}

	decryptedLedgerPassword, decryptedApiKey, err := utils.DecryptApiKeyAndLedgerPassword(user.LedgerPassword, user.KmsEncryptedLedgerPassword, userPublicKey.ApiKey, userPublicKey.KmsEncryptedApiKey, logger)
	if err != nil {
		logger.Error(err.Error())
		return response.InternalServerError(fmt.Sprintf("Error in decryption: %s", err.Error()), errtrace.Wrap(err))
	}

	validateCvvPayloadRecord, errResponse := dao.ConsumePayload(user.Id, requestData.ValidateCvv.PayloadId)
	if errResponse != nil {
		return errResponse
	}
	userClient := ledger.CreateCardApiClient(userPublicKey.PublicKey, requestData.ValidateCvv.Signature, validateCvvPayloadRecord.Payload, user.Email, decryptedLedgerPassword, userPublicKey.KeyId, decryptedApiKey)
	var validateCvvRequest ledger.ValidateCvvRequest
	err = json.Unmarshal([]byte(validateCvvPayloadRecord.Payload), &validateCvvRequest)
	if err != nil {
		return response.InternalServerError(fmt.Sprintf("Error while unmarshaling validateCvv payload: %s", err.Error()), errtrace.Wrap(err))
	}
	validateCvvResponse, err := userClient.ValidateCvv(validateCvvRequest)
	if err != nil {
		return response.InternalServerError(fmt.Sprintf("Error calling ledger validateCvv: %s", err.Error()), errtrace.Wrap(err))
	}
	if validateCvvResponse.Error != nil {
		logger.Error("ledger responded with an error to validateCvv", "code", validateCvvResponse.Error.Code, "msg", validateCvvResponse.Error.Message)
		return response.InternalServerError(fmt.Sprintf("ledger responded with an error to validateCvv: %s", validateCvvResponse.Error.Message), errtrace.New(""))
	}
	if validateCvvResponse.Result == nil || validateCvvResponse.Result.Api.Type != ledger.ValidateCvvValidType {
		return response.BadRequestErrors{
			Errors: []response.BadRequestError{
				{
					FieldName: "cvv",
					Error:     "invalid",
				},
			},
		}
	}

	changeCardPinPayloadRecord, errResponse := dao.ConsumePayload(user.Id, requestData.SetCardPin.PayloadId)
	if errResponse != nil {
		return errResponse
	}
	userClient = ledger.CreateCardApiClient(userPublicKey.PublicKey, requestData.SetCardPin.Signature, changeCardPinPayloadRecord.Payload, user.Email, decryptedLedgerPassword, userPublicKey.KeyId, decryptedApiKey)
	var changeCardPinRequest ledger.ChangePinRequest
	err = json.Unmarshal([]byte(changeCardPinPayloadRecord.Payload), &changeCardPinRequest)
	if err != nil {
		return response.InternalServerError(fmt.Sprintf("Error while unmarshaling setCardPingRequest payload: %s", err.Error()), errtrace.Wrap(err))
	}
	changeCardPinResponse, err := userClient.ChangePin(changeCardPinRequest)
	if err != nil {
		return response.InternalServerError(fmt.Sprintf("Error calling ledger changeCardPin: %s", err.Error()), errtrace.Wrap(err))
	}
	if changeCardPinResponse.Error != nil {
		logger.Error("ledger responded with an error to changeCardPin", "code", changeCardPinResponse.Error.Code, "msg", changeCardPinResponse.Error.Message)
		return response.InternalServerError(fmt.Sprintf("ledger responded with an error to changeCardPin: %s", changeCardPinResponse.Error.Message), errtrace.New(""))
	}
	if changeCardPinResponse.Result == nil {
		logger.Error("The ledger responded with an empty result object", "response", changeCardPinResponse)
		return response.InternalServerError("The ledger responded with an empty result object", errtrace.New(""))
	}
	if changeCardPinResponse.Result.Api.Type != "PIN_CHANGE_ACK" {
		logger.Error("Unexpected response type from Ledger API", "Api.Type", changeCardPinResponse.Result.Api.Type)
		return response.InternalServerError("Unexpected response type from Ledger API", errtrace.New(""))
	}

	updateStatusPayloadRecord, errResponse := dao.ConsumePayload(user.Id, requestData.UpdateStatus.PayloadId)
	if errResponse != nil {
		return errResponse
	}
	userClient = ledger.CreateCardApiClient(userPublicKey.PublicKey, requestData.UpdateStatus.Signature, updateStatusPayloadRecord.Payload, user.Email, decryptedLedgerPassword, userPublicKey.KeyId, decryptedApiKey)
	var updateStatusRequest ledger.UpdateStatusRequest
	err = json.Unmarshal([]byte(updateStatusPayloadRecord.Payload), &updateStatusRequest)
	if err != nil {
		logger.Error("Error while unmarshaling payload", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error while unmarshaling payload: %s", err.Error()), errtrace.Wrap(err))
	}

	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerSignedClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, ledgerParamsBuilder)
	getCardRequest := ledgerSignedClient.BuildGetCardDetailsRequest(user.LedgerCustomerNumber, userAccountCard.AccountNumber, userAccountCard.CardId)
	getCardResponse, err := ledgerSignedClient.GetCardDetails(getCardRequest)
	if err != nil {
		return response.InternalServerError(fmt.Sprintf("Error calling ledger getCardDetails: %s", err.Error()), errtrace.Wrap(err))
	}
	if getCardResponse.Error != nil {
		logger.Error("ledger responded with an error to getCardDetails", "code", getCardResponse.Error.Code, "msg", getCardResponse.Error.Message)
		return response.InternalServerError(fmt.Sprintf("ledger responded with an error to getCardDetails: %s", getCardResponse.Error.Message), errtrace.New(""))
	}
	if getCardResponse.Result == nil {
		logger.Error("The ledger responded with an empty result object", "response", getCardResponse)
		return response.InternalServerError("The ledger responded with an empty result object", errtrace.New(""))
	}
	err = updateCardMaskNumberInDB(user.Id, getCardResponse.Result.Card.CardMaskNumber)
	if err != nil {
		logger.Error("Error while inserting maskedCardNumber", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error while inserting maskedCardNumber: %s", err.Error()), errtrace.Wrap(err))
	}
	if MapLedgerCardStatus(getCardResponse.Result.Card.CardStatus, logger) == "active" {
		logger.Warn("Card already activated", "carId", userAccountCard.CardId)
		resp := response.GetCardResponse{
			Card: response.CardData{
				CardId:         getCardResponse.Result.Card.CardId,
				CardStatus:     MapLedgerCardStatus(getCardResponse.Result.Card.CardStatus, logger),
				CardStatusRaw:  getCardResponse.Result.Card.CardStatus,
				OrderStatus:    getCardResponse.Result.Card.OrderStatus,
				IsReIssue:      getCardResponse.Result.Card.IsReIssue,
				IsReplace:      getCardResponse.Result.Card.IsReplace,
				CardMaskNumber: getCardResponse.Result.Card.CardMaskNumber,
				CardExpiryDate: getCardResponse.Result.Card.CardExpiryDate,
				ExternalCardId: getCardResponse.Result.Card.ExternalCardId,
			},
		}
		return c.JSON(http.StatusOK, resp)
	}

	responseData, err := userClient.UpdateStatus(updateStatusRequest)
	if err != nil {
		logger.Error("Error from callLedgerGetCard", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error from callLedgerGetCard: %s", err.Error()), errtrace.Wrap(err))
	}
	if responseData.Error != nil {
		// ledger returns 1018 status code in case of Inappropriate status action
		if responseData.Error.Code == "1018" {
			return response.BadRequestErrors{
				Errors: []response.BadRequestError{
					{
						FieldName: "status",
						Error:     "invalid",
					},
				},
			}
		}
		logger.Error("The ledger responded with an error", "code", responseData.Error.Code, "msg", responseData.Error.Message)
		return response.InternalServerError(fmt.Sprintf("The ledger responded with an error: %s", responseData.Error.Message), errtrace.New(""))
	}

	if MapLedgerCardStatus(responseData.Result.Card.CardStatus, logger) == "active" {
		logger.Info("Card activated successfully")
		response := response.GetCardResponse{
			Card: response.CardData{
				CardId:        responseData.Result.Card.CardId,
				CardStatus:    MapLedgerCardStatus(responseData.Result.Card.CardStatus, logger),
				CardStatusRaw: responseData.Result.Card.CardStatus,
				IsReIssue:     responseData.Result.Card.IsReIssue,
				IsReplace:     responseData.Result.Card.IsReplace,
				// NetXD does not return the entire GetCardDetails response for UpdateStatus
				// https://dreamfi.atlassian.net/browse/DT-765
				OrderStatus:            getCardResponse.Result.Card.OrderStatus,
				CardMaskNumber:         getCardResponse.Result.Card.CardMaskNumber,
				CardExpiryDate:         getCardResponse.Result.Card.CardExpiryDate,
				ExternalCardId:         getCardResponse.Result.Card.ExternalCardId,
				PreviousCardId:         userAccountCard.PreviousCardId,
				PreviousCardMaskNumber: userAccountCard.PreviousCardMaskNumber,
			},
		}
		return c.JSON(http.StatusOK, response)
	}

	// ledger returned unexpected response
	logger.Error("ledger returned unexpected response")
	return response.InternalServerError("ledger returned unexpected response", errtrace.New(""))
}

func updateCardMaskNumberInDB(userId, maskedCardNumber string) error {
	result := db.DB.Model(dao.UserAccountCardDao{}).Where("user_id=? AND account_status=?", userId, "ACTIVE").Update("card_mask_number", maskedCardNumber)
	if result.Error != nil {
		return errtrace.Wrap(result.Error)
	}
	return nil
}
