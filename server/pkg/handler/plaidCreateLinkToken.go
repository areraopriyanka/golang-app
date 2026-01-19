package handler

import (
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/plaid"
	"process-api/pkg/security"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @Summary PlaidCreateLinkToken
// @Description Creates a Plaid link token
// @Tags plaid
// @Produce json
// @Param payload body PlaidLinkTokenRequest true "PlaidLinkTokenRequest"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} PlaidLinkTokenResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/plaid/link/token [post]
func (h *Handler) PlaidCreateLinkToken(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId
	logger := logging.GetEchoContextLogger(c).WithGroup("PlaidCreateLinkToken").With("userId", userId)

	requestData := new(PlaidLinkTokenRequest)
	err := c.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	ps := plaid.PlaidService{Logger: logger, Plaid: h.Plaid, DB: db.DB, WebhookURL: h.Config.Plaid.WebhookURL}
	linkToken, err := ps.LinkTokenCreateRequest(userId, requestData.Platform, h.Config.Plaid.LinkRedirectURI, h.Env, nil)
	if err != nil {
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      err.Error(),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}
	logger.Info("Plaid link token created successfully")
	return c.JSON(http.StatusOK, PlaidLinkTokenResponse{LinkToken: linkToken})
}

type PlaidLinkTokenRequest struct {
	Platform string `json:"platform" validate:"required,oneof=ios android web"`
}

type PlaidLinkTokenResponse struct {
	LinkToken string `json:"linkToken" validate:"required" mask:"true"`
}
