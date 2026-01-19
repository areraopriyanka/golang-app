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

type PlaidExchangePublicTokenRequest struct {
	plaid.PlaidLinkOnSuccess
	PublicToken string `json:"publicToken" validate:"required"`
}

// @Summary PlaidExchangePublicToken
// @Description Exchanges a Plaid public token for a secret access token
// @Tags plaid
// @Produce json
// @Param payload body PlaidExchangePublicTokenRequest true "PlaidExchangePublicTokenRequest"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 201 "Created"
// @header 201 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/plaid/public_token/exchange [post]
func (h *Handler) PlaidExchangePublicToken(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	requestData := new(PlaidExchangePublicTokenRequest)
	err := c.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	userId := cc.UserId
	logger := logging.GetEchoContextLogger(c).WithGroup("PlaidExchangePublicToken").With("linkSessionID", requestData.LinkSessionID)
	ps := plaid.PlaidService{Logger: logger, Plaid: h.Plaid, DB: db.DB, WebhookURL: h.Config.Plaid.WebhookURL}

	// NOTE: it might not always make sense to attempt to check for duplicates at this stage of linking.
	// Certain linking strategies (e.g., Same Day Micro-deposits) might not return an institution id.
	// If institution id is missing, it might make sense to implement a way to identify duplicate items
	// after link has completed: https://plaid.com/docs/link/duplicate-items/#identifying-existing-duplicate-items
	// Captured in DT-1032
	hasDuplicates, err := ps.CheckForDuplicateAccounts(userId, requestData.InstitutionID, &requestData.Accounts)
	if err != nil {
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      err.Error(),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}
	if hasDuplicates {
		msg := "Your account with this institution has already been linked."
		if len(requestData.Accounts) > 1 {
			msg = "Your accounts with this institution have already been linked."
		}
		return response.ErrorResponse{ErrorCode: "DUPLICATE_ACCOUNTS", Message: msg, StatusCode: http.StatusConflict}
	}

	logger.Info("sending publicToken", "publicToken", requestData.PublicToken)

	resp, err := ps.ItemPublicTokenExchangeRequest(requestData.PublicToken)
	if err != nil {
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      err.Error(),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	plaidItemId := resp.PlaidItemId
	accessToken := resp.AccessToken

	err = ps.InsertItem(userId, plaidItemId, accessToken)
	if err != nil {
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      err.Error(),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}
	err = ps.InitialAccountsGetRequest(userId, plaidItemId, accessToken)
	if err != nil {
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      err.Error(),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	return cc.NoContent(http.StatusCreated)
}
