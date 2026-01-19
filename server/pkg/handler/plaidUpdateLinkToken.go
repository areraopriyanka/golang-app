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

// @Summary PlaidUpdateLinkToken
// @Description Creates a Plaid link token for Plaid Link's update mode
// @Tags plaid
// @Produce json
// @Param payload body PlaidUpdateLinkTokenRequest true "PlaidUpdateLinkTokenRequest"
// @Success 200 {object} PlaidLinkTokenResponse
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/plaid/link/token/update [post]
func (h *Handler) PlaidUpdateLinkToken(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId
	logger := logging.GetEchoContextLogger(c).WithGroup("PlaidCreateLinkToken").With("userId", userId)

	requestData := new(PlaidUpdateLinkTokenRequest)
	err := c.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	accountID := *requestData.AccountID
	plaidAccount, err := dao.PlaidAccountDao{}.GetAccountForUserByID(userId, accountID)
	if err != nil {
		logger.Error("Failed to get Plaid account", "accountID", accountID, "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Failed to get Plaid account: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}
	if plaidAccount == nil {
		return response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, LogMessage: fmt.Sprintf("Plaid account id %s not found for user %s", accountID, userId), MaybeInnerError: errtrace.New("")}
	}
	plaidItemID := plaidAccount.PlaidItemID

	ps := plaid.PlaidService{Logger: logger, Plaid: h.Plaid, DB: db.DB, WebhookURL: h.Config.Plaid.WebhookURL}

	var linkToken string
	item, err := dao.PlaidItemDao{}.GetItemForUserIDByPlaidItemID(userId, plaidItemID)
	if err != nil {
		logger.Error("Failed to get Plaid item", "itemId", plaidItemID, "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Failed to get Plaid item: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}
	if item == nil {
		logger.Error("Plaid item not found", "itemId", plaidItemID)
		return response.NotFoundError(fmt.Sprintf("Plaid item not found. itemId: %s", plaidItemID), errtrace.New(""))
	}

	accessToken, decryptErr := utils.DecryptPlaidAccessToken(item.EncryptedAccessToken, item.KmsEncryptedAccessToken)
	if decryptErr != nil {
		logger.Error("Failed to decrypt access token", "itemId", plaidItemID, "error", decryptErr.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Failed to decrypt access token: %s", decryptErr.Error()),
			MaybeInnerError: errtrace.Wrap(decryptErr),
		}
	}

	linkToken, err = ps.LinkTokenCreateRequest(userId, requestData.Platform, h.Config.Plaid.LinkRedirectURI, h.Env, utils.Pointer(accessToken))
	if err != nil {
		logger.Error("Failed to create update mode link token", "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Failed to create update mode link token: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}
	logger.Debug("Plaid update mode link token created successfully", "itemId", plaidItemID)

	return c.JSON(http.StatusOK, PlaidLinkTokenResponse{LinkToken: linkToken})
}

type PlaidUpdateLinkTokenRequest struct {
	Platform  string  `json:"platform" validate:"required,oneof=ios android web"`
	AccountID *string `json:"accountId" validate:"required"`
}
