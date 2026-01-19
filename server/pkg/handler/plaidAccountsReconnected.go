package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/plaid"
	"process-api/pkg/security"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

type PlaidAccountsReconnectedRequest struct {
	Accounts      []plaid.PlaidLinkAccount `json:"accounts" validate:"required"`
	LinkSessionID string                   `json:"linkSessionId" validate:"required"`
}

// @Summary PlaidAccountsReconnected
// @Description Registers that accounts were reconnected after Plaid Link was run in Update mode
// @Tags plaid
// @Produce json
// @Param payload body PlaidAccountsReconnectedRequest true "PlaidAccountsReconnectedRequest"
// @Success 200 "Ok"
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/plaid/accounts/reconnected [post]
func (h *Handler) PlaidAccountsReconnected(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	requestData := new(PlaidAccountsReconnectedRequest)
	err := c.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	userId := cc.UserId
	logger := logging.GetEchoContextLogger(c).WithGroup("PlaidAccountsReconnected").With("userId", userId, "linkSessionID", requestData.LinkSessionID)

	if len(requestData.Accounts) == 0 {
		logger.Error("No plaid accounts were sent to update")
		return response.NotFoundError("No plaid accounts were sent to update", errtrace.New(""))
	}

	plaidAccountID := requestData.Accounts[0].ID
	item, err := dao.PlaidItemDao{}.FindFirstItemForUserIDByPlaidAccountID(userId, plaidAccountID)
	if err != nil {
		logger.Error("DB error attempting to find item associated with plaid account id", "plaidAccountID", plaidAccountID, "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("DB error attempting to find item associated with plaid account id: %s, error: %s", plaidAccountID, err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}
	if item == nil {
		logger.Error("Could not find item associated with plaid account id", "plaidAccountID", plaidAccountID)
		return response.NotFoundError(fmt.Sprintf("Could not find item associated with plaid account id. plaidAccountID: %s", plaidAccountID), errtrace.New(""))
	}

	err = dao.PlaidItemDao{}.ClearItemError(item.PlaidItemID)
	if err != nil {
		logger.Error("DB error attempting to clear Plaid item's error", "plaidItemID", item.PlaidItemID, "plaidAccountID", plaidAccountID, "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("DB error attempting to clear Plaid item's error. plaidItemID: %s, error: %s", item.PlaidItemID, err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	err = dao.PlaidItemDao{}.SetIsPendingDisconnect(item.PlaidItemID, false)
	if err != nil {
		logger.Error("DB error attempting to reset isPendingDisconnect for Plaid item", "plaidItemID", item.PlaidItemID, "plaidAccountID", plaidAccountID, "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("DB error attempting to reset isPendingDisconnect for Plaid item. plaidItemID: %s, error: %s", item.PlaidItemID, err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	return cc.NoContent(http.StatusOK)
}
