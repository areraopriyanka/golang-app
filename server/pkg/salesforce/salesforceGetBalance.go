package salesforce

import (
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"

	"braces.dev/errtrace"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/labstack/echo/v4"
)

type BalanceResponse struct {
	// The account's available balance as expressed in cents.
	AvailableBalanceCents int64 `json:"availableBalanceCents" example:"10000" validate:"required"`
}

// @Summary SalesforceGetBalance
// @Description Gets the available balance for a given ledger account id
// @Tags salesforce
// @Produce json
// @Param ledgerAccountID path int true "ledger account id"
// @Success 200 {object} BalanceResponse
// @Failure 401 {object} response.ErrorResponse "Missing or invalid authentication"
// @Failure 403 {object} response.ErrorResponse "Insufficient permissions (missing required scope)"
// @Failure 404 {object} response.ErrorResponse "Could not find account"
// @Failure 500 {object} response.ErrorResponse
// @Router /api/salesforce/accounts/{ledgerAccountID}/balance [get]
func SalesforceGetBalance(c echo.Context) error {
	ledgerAccountID := c.Param("ledgerAccountID")
	logger := logging.GetEchoContextLogger(c).With("ledgerAccountID", ledgerAccountID)
	claims, ok := c.Get("salesforce_claims").(*validator.ValidatedClaims)
	if !ok {
		logger.Error("Could not get custom salesforce claims")
		return response.UnauthorizedError("Could not get custom salesforce claims")
	}

	customClaims, ok := claims.CustomClaims.(*SalesforceClaims)
	if !ok {
		logger.Error("Could not cast to CustomClaims")
		return response.UnauthorizedError("Could not cast to CustomClaims")
	}

	if !customClaims.HasScope("read:balance") {
		logger.Error("Custom claims missing read:balance scope")
		return response.ForbiddenError("Custom claims missing read:balance scope", errtrace.New(""))
	}

	logger.Debug("Fetching balance for Salesforce")

	account, err := dao.UserAccountCardDao{}.FindOneByAccountID(db.DB, ledgerAccountID)
	if err != nil {
		logger.Error("Failed to fetch account from database", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Failed to fetch account from database: %s", err.Error()), errtrace.Wrap(err))
	}
	if account == nil {
		logger.Error("Account not found")
		return response.NotFoundError("Account not found", errtrace.New(""))
	}

	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDLedgerApiClient(config.Config.Ledger, ledgerParamsBuilder)
	ledgerAccountResp, err := ledgerClient.GetAccount(account.AccountId)
	if err != nil {
		logger.Error("error while calling ledger's GetAccount", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("error while calling ledger's GetAccount: %s", err.Error()), errtrace.Wrap(err))
	}
	if ledgerAccountResp.Error != nil {
		logger.Error("error from ledger's GetAccount", "error", ledgerAccountResp.Error)
		return response.InternalServerError(fmt.Sprintf("error from ledger's GetAccount: %s", ledgerAccountResp.Error), errtrace.New(""))
	}
	if ledgerAccountResp.Result == nil {
		logger.Error("no error was reported from ledger GetAccount, but Result is missing from response")
		return response.InternalServerError("no error was reported from ledger GetAccount, but Result is missing from response", errtrace.New(""))
	}

	ledgerAccount := ledgerAccountResp.Result.Account
	// While ledgerAccount has many different "balance" fields, it looks like `Balance` is the correct
	// one to use, provided the docs from GetAccount apply here as well:
	// https://apidocs.netxd.com/developers/docs/account_apis/Get%20Account
	availableBalanceCents := int64(ledgerAccount.Balance)
	// Build response
	balance := BalanceResponse{
		AvailableBalanceCents: availableBalanceCents,
	}

	logger.Debug("Returning balance", "availableBalanceCents", availableBalanceCents)

	return c.JSON(http.StatusOK, balance)
}
