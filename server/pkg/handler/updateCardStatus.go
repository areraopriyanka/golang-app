package handler

import (
	"fmt"
	"log/slog"
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

func UpdateCardStatusHandler(c echo.Context, statusAction string) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId
	logger := logging.GetEchoContextLogger(c)

	responseData, err := UpdateStatus(logger, userId, statusAction)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.UpdateCardStatusResponse{
		UpdatedCardStatus: MapLedgerCardStatus(responseData.Result.Card.CardStatus, logger),
	})
}

func UpdateStatus(logger *slog.Logger, userId, statusAction string) (*ledger.NetXDApiResponse[ledger.UpdateStatusResult], error) {
	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return nil, errResponse
	}
	userAccountCard, errResponse := dao.RequireActiveCardHolderForUser(userId)
	if errResponse != nil {
		return nil, errResponse
	}

	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, ledgerParamsBuilder)

	payload, err := ledgerClient.BuildUpdateStatusRequest(user.LedgerCustomerNumber, userAccountCard.CardId, userAccountCard.AccountNumber, statusAction, "", false)
	if err != nil {
		logger.Error("Error while generating updateStatus request payload", "error", err.Error())
		return nil, response.InternalServerError(fmt.Sprintf("Error while generating updateStatus request payload: %s", err.Error()), errtrace.Wrap(err))
	}

	var responseData ledger.NetXDApiResponse[ledger.UpdateStatusResult]

	responseData, err = ledgerClient.UpdateStatus(*payload)
	if err != nil {
		logger.Error("Error from updateCardStatus", "error", err.Error())
		return nil, response.InternalServerError(fmt.Sprintf("Error from updateCardStatus: %s", err.Error()), errtrace.Wrap(err))
	}

	if responseData.Error != nil {
		if responseData.Error.Code == "1018" {
			return nil, response.BadRequestErrors{
				Errors: []response.BadRequestError{
					{
						FieldName: "status",
						Error:     "invalid",
					},
				},
			}
		}
		logger.Error("The ledger responded with an error", "code", responseData.Error.Code, "msg", responseData.Error.Message)
		return nil, response.InternalServerError(fmt.Sprintf("The ledger responded with an erro: %s", responseData.Error.Message), errtrace.New(""))
	}

	if responseData.Result.Card.CardStatus != "" {
		logger.Info("Updated cardStatus", "status", responseData.Result.Card.CardStatus)
		return &responseData, nil
	}

	logger.Error("ledger returned unexpected response")
	return nil, response.InternalServerError("ledger returned unexpected response", errtrace.New(""))
}
