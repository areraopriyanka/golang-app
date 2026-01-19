package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// @Summary GetCardDetails
// @Description Returns card details.
// @Tags cards
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} response.GetCardResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Router /account/cards [get]
func GetCardDetails(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId
	logger := logging.GetEchoContextLogger(c)

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}
	userAccountCard, errResponse := dao.RequireCardHolderForUser(userId)
	if errResponse != nil {
		return errResponse
	}
	// Ledger returns a response for getCardDetails API even if the account is SUSPENDED
	// To avoid calling the Ledger API, directly returning card details from the DB
	if !userAccountCard.IsActive() {
		return c.JSON(http.StatusOK, response.GetCardResponse{
			Card: response.CardData{
				CardId:                 userAccountCard.CardId,
				CardStatus:             MapLedgerCardStatus("CLOSED", logger),
				CardStatusRaw:          "CLOSED",
				CardMaskNumber:         userAccountCard.CardMaskNumber,
				CardExpiryDate:         userAccountCard.CardExpirationDate,
				IsReIssue:              userAccountCard.IsReissue,
				IsReplace:              userAccountCard.IsReplace,
				IsReplaceLocked:        userAccountCard.IsReplaceLocked(),
				IsPreviousCardFrozen:   false,
				PreviousCardId:         userAccountCard.PreviousCardId,
				PreviousCardMaskNumber: userAccountCard.PreviousCardMaskNumber,
			},
		})
	}

	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerSignedClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, ledgerParamsBuilder)

	getCardRequest := ledgerSignedClient.BuildGetCardDetailsRequest(user.LedgerCustomerNumber, userAccountCard.AccountNumber, userAccountCard.CardId)
	getCardResponse, err := ledgerSignedClient.GetCardDetails(getCardRequest)
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

	var isPreviousCardFrozen bool

	if userAccountCard.IsReissue && userAccountCard.PreviousCardId != nil {
		currentCardStatus := getCardResponse.Result.Card.CardStatus
		isCurrentCardInactive := (currentCardStatus == "RETURNED_UNDELIVERED" || currentCardStatus == "CARD_IS_NOT_ACTIVATED")

		if isCurrentCardInactive {
			previousCardRequest := ledgerSignedClient.BuildGetCardDetailsRequest(user.LedgerCustomerNumber, userAccountCard.AccountNumber, *userAccountCard.PreviousCardId)
			previousCardResponse, err := ledgerSignedClient.GetCardDetails(previousCardRequest)
			if err != nil {
				logger.Error("Error getting previous card details", "error", err.Error(), "previousCardId", *userAccountCard.PreviousCardId)
				isPreviousCardFrozen = false
			} else if previousCardResponse.Error != nil {
				logger.Error("Ledger error getting previous card details", "code", previousCardResponse.Error.Code, "message", previousCardResponse.Error.Message)
				isPreviousCardFrozen = false
			} else {
				previousCardMappedStatus := MapLedgerCardStatus(previousCardResponse.Result.Card.CardStatus, logger)
				isPreviousCardFrozen = (previousCardMappedStatus == "frozen")
			}
		}
	}

	// TODO: Remove this conditional code once the Ledger's GetCardDetails API issue is fixed.
	expiryDate := userAccountCard.CardExpirationDate
	if expiryDate == "" {
		expiryDate = getCardResponse.Result.Card.CardExpiryDate
	}

	response := response.GetCardResponse{
		Card: response.CardData{
			CardId:                 getCardResponse.Result.Card.CardId,
			CardStatus:             MapLedgerCardStatus(getCardResponse.Result.Card.CardStatus, logger),
			CardStatusRaw:          getCardResponse.Result.Card.CardStatus,
			OrderStatus:            getCardResponse.Result.Card.OrderStatus,
			IsReIssue:              getCardResponse.Result.Card.IsReIssue || userAccountCard.IsReissue,
			IsReplace:              getCardResponse.Result.Card.IsReplace || userAccountCard.IsReplace,
			IsReplaceLocked:        userAccountCard.IsReplaceLocked(),
			CardMaskNumber:         getCardResponse.Result.Card.CardMaskNumber,
			CardExpiryDate:         expiryDate,
			ExternalCardId:         getCardResponse.Result.Card.ExternalCardId,
			PreviousCardId:         userAccountCard.PreviousCardId,
			PreviousCardMaskNumber: userAccountCard.PreviousCardMaskNumber,
			IsPreviousCardFrozen:   isPreviousCardFrozen,
		},
	}
	logger.Info("GetCardDetails was successful cardId", "cardId", getCardResponse.Result.Card.CardId)

	return c.JSON(http.StatusOK, response)
}

func MapLedgerCardStatus(s string, logger *slog.Logger) string {
	switch s {
	case "RETURNED_UNDELIVERED", "CARD_IS_NOT_ACTIVATED":
		return "inactive"
	case "ACTIVE", "ACTIVATED":
		return "active"
	case "TEMPRORY_BLOCKED_BY_CLIENT", "TEMPRORY_BLOCKED_BY_ADMIN":
		return "frozen"
	case "DEACTIVATED", "CLOSED", "LOST_STOLEN", "EXPIRED_CARD", "CARD_REQUEST_NOT_PROCESSED":
		return "cancelled"
	}
	logger.Info("Unknown cardStatus in card details", "status", s)
	return ""
}

func CreatePayloadResponse(payload interface{}) (*response.BuildPayloadResponse, error) {
	jsonPayloadBytes, marshallErr := json.Marshal(payload)
	if marshallErr != nil {
		return nil, errtrace.Wrap(fmt.Errorf("an error occurred while marshalling the payload: %s", marshallErr.Error()))
	}

	sessionId := uuid.New().String()

	payloadRecord := dao.SignablePayloadDao{
		Id:      sessionId,
		Payload: string(jsonPayloadBytes),
	}

	result := db.DB.Create(&payloadRecord)
	if result.Error != nil {
		return nil, errtrace.Wrap(fmt.Errorf("an error while creating payload record: %s", result.Error.Error()))
	}

	response := &response.BuildPayloadResponse{
		PayloadId: sessionId,
		Payload:   string(jsonPayloadBytes),
	}
	return response, nil
}
