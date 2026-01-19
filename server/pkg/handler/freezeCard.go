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

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @Summary FreezeCard
// @Description Lock the card and update card status to TEMPRORY_BLOCKED_BY_CLIENT.
// @Tags cards
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} response.UpdateCardStatusResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 422 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/cards/freeze [post]
func FreezeCard(c echo.Context) error {
	return UpdateCardStatusMaybeInactiveReIssue(c, ledger.LOCK)
}

// @Summary UnfreezeCard
// @Description Unlock the card and update card status to ACTIVATED".
// @Tags cards
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} response.UpdateCardStatusResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 422 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/cards/unfreeze [post]
func UnfreezeCard(c echo.Context) error {
	return UpdateCardStatusMaybeInactiveReIssue(c, ledger.UNLOCK)
}

func UpdateCardStatusMaybeInactiveReIssue(c echo.Context, statusAction string) error {
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
	userAccountCard, errResponse := dao.RequireActiveCardHolderForUser(userId)
	if errResponse != nil {
		return errResponse
	}

	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, ledgerParamsBuilder)

	getCardRequest := ledgerClient.BuildGetCardDetailsRequest(user.LedgerCustomerNumber, userAccountCard.AccountNumber, userAccountCard.CardId)
	getCardResponse, err := ledgerClient.GetCardDetails(getCardRequest)
	if err != nil {
		logger.Error("Error from callLedgerGetCardDetails", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error from callLedgerGetCardDetails: %s", err.Error()), errtrace.Wrap(err))
	}

	if getCardResponse.Error != nil {
		logger.Error("The ledger responded with an error", "code", getCardResponse.Error.Code, "message", getCardResponse.Error.Message)
		return response.InternalServerError(fmt.Sprintf("The ledger responded with an error: %s", getCardResponse.Error.Message), errtrace.New(""))
	}

	if getCardResponse.Result == nil {
		logger.Error("The ledger responded with an empty result object", "response", getCardResponse)
		return response.InternalServerError("The ledger responded with an empty result object", errtrace.New(""))
	}

	var cardId string

	if MapLedgerCardStatus(getCardResponse.Result.Card.CardStatus, logger) == "inactive" && (getCardResponse.Result.Card.IsReIssue || userAccountCard.IsReissue) {
		cardId = *userAccountCard.PreviousCardId
	} else {
		cardId = userAccountCard.CardId
	}

	payload, err := ledgerClient.BuildUpdateStatusRequest(user.LedgerCustomerNumber, cardId, userAccountCard.AccountNumber, statusAction, "", false)
	if err != nil {
		logger.Error("Error while generating updateStatus request payload", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error while generating updateStatus request payload: %s", err.Error()), errtrace.Wrap(err))
	}

	var responseData ledger.NetXDApiResponse[ledger.UpdateStatusResult]

	responseData, err = ledgerClient.UpdateStatus(*payload)
	if err != nil {
		logger.Error("Error from updateCardStatus", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error from updateCardStatus: %s", err.Error()), errtrace.Wrap(err))
	}

	if responseData.Error != nil {
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

	if responseData.Result.Card.CardStatus != "" {
		logger.Info("Updated cardStatus", "status", responseData.Result.Card.CardStatus)
		return c.JSON(http.StatusOK, response.UpdateCardStatusResponse{
			UpdatedCardStatus: MapLedgerCardStatus(responseData.Result.Card.CardStatus, logger),
		})
	}

	logger.Error("ledger returned unexpected response")
	return response.InternalServerError("ledger returned unexpected response", errtrace.New(""))
}
