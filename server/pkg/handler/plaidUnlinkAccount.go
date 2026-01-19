package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/plaid"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @Summary PlaidAccountUnlink
// @Description Unlinks a Plaid item by calling Plaid's /item/remove endpoint and deleting associated database records
// @Tags plaid
// @Produce json
// @Param payload body PlaidAccountUnlinkRequest true "PlaidAccountUnlinkRequest"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 "OK"
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/plaid/account [delete]
func (h *Handler) PlaidAccountUnlink(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	requestData := new(PlaidAccountUnlinkRequest)
	err := c.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	userId := cc.UserId
	logger := logging.GetEchoContextLogger(c).WithGroup("PlaidUnlinkAccount").With("userId", userId, "plaidAccountId", requestData.PlaidExternalAccountID)
	ps := plaid.PlaidService{Logger: logger, Plaid: h.Plaid, DB: db.DB, WebhookURL: h.Config.Plaid.WebhookURL}
	account, err := dao.PlaidAccountDao{}.GetAccountForUserByID(userId, requestData.PlaidExternalAccountID)
	if err != nil {
		logger.Error("db error finding plaid account", "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("db error finding plaid account: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}
	if account == nil {
		logger.Error("could not find plaid account")
		return response.NotFoundError("could not find plaid account", errtrace.New(""))
	}

	plaidItemID := account.PlaidItemID
	item, err := dao.PlaidItemDao{}.GetItemForUserByItemID(userId, plaidItemID)
	if err != nil {
		logger.Error("db error finding plaid item", "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("db error finding plaid item: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}
	if item == nil {
		logger.Error("could not find plaid item")
		return response.NotFoundError("could not find plaid item", errtrace.New(""))
	}

	accessToken, err := utils.DecryptPlaidAccessToken(item.EncryptedAccessToken, item.KmsEncryptedAccessToken)
	if err != nil {
		logger.Error("Could not decrypt Plaid access token", "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Could not decrypt Plaid access token: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	err = ps.UnlinkItem(userId, plaidItemID, accessToken)
	if err != nil {
		logger.Error("error unlinking item", "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("error unlinking item: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	return cc.NoContent(http.StatusOK)
}

type PlaidAccountUnlinkRequest struct {
	PlaidExternalAccountID string `json:"plaidAccountId" validate:"required"`
}
