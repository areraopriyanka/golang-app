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

type Transaction struct {
	// The reference id for the transaction
	Id string `json:"id" example:"ledger.ach.transfer_ach_pull_1756829941484592000" validate:"required"`
	// The ledger's account id associated with the transaction
	AccountId string `json:"accountId" example:"11522031" validate:"required"`
	// The amount of funds (expressed in cents)
	AmountCents int64 `json:"amountCents" example:"10000" validate:"required"`
	// The name associated with the debtor's account (if a credit); the name associated with the creditor's account otherwise
	Merchant string `json:"merchant" example:"Alberta Bobbeth Charleson" validate:"required"`
	// The type of transaction as defined by the ledger: "PRE_AUTH", "PURCHASE_WITH_CASH", "PURCHASE", etc
	Type string `json:"type" example:"ACH_OUT" validate:"required"`
	// The date of the transaction
	Date string `json:"date" example:"2025-09-28T12:21:53Z" validate:"required"`
	// The status of the transaction: ACTIVE, CLOSED, DORMANT, SUSPENDED
	Status string `json:"status" example:"ACTIVE" validate:"required"`
}

// @Summary SalesforceGetTransactions
// @Description Gets a list of transactions given a ledger account id
// @Tags salesforce
// @Produce json
// @Param ledgerAccountID path string true "ledger account id"
// @Success 200 {object} []Transaction
// @Failure 401 {object} response.ErrorResponse "Missing or invalid authentication"
// @Failure 403 {object} response.ErrorResponse "Insufficient permissions (missing required scope)"
// @Failure 404 {object} response.ErrorResponse "Could not find account"
// @Failure 500 {object} response.ErrorResponse
// @Router /api/salesforce/accounts/{ledgerAccountID}/transactions [get]
func SalesforceGetTransactions(c echo.Context) error {
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

	if !customClaims.HasScope("read:transactions") {
		logger.Error("Custom claims missing read:transactions scope")
		return response.ForbiddenError("Custom claims missing read:transactions scope", errtrace.New(""))
	}

	logger.Debug("Fetching transactions for Salesforce")

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

	ledgerRequest := ledger.BuildListTransactionsByAccountPayload(account.AccountNumber)
	ledgerResponse, err := ledgerClient.ListTransactionsByAccount(ledgerRequest)
	if err != nil {
		logger.Error("Error from listTransactionsByAccount", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error from listTransactionsByAccount: %s", err.Error()), errtrace.Wrap(err))
	}

	if ledgerResponse.Error != nil {
		logger.Error("The ledger responded with an error", "code", ledgerResponse.Error.Code, "msg", ledgerResponse.Error.Message)
		return response.InternalServerError(fmt.Sprintf("The ledger responded with an error: %s", ledgerResponse.Error.Message), errtrace.New(""))
	}

	if ledgerResponse.Result == nil {
		logger.Error("The ledger responded with an empty result object")
		return response.InternalServerError("The ledger responded with an empty result object", errtrace.New(""))
	}

	transactions := make([]Transaction, 0, len(ledgerResponse.Result.AccountTransactions))

	for _, data := range ledgerResponse.Result.AccountTransactions {
		var merchantAccount ledger.ListTransactionsByAccountResultTransactionAccount
		if data.Credit {
			merchantAccount = data.DebtorAccount
		} else {
			merchantAccount = data.CreditorAccount
		}
		merchantName := ledger.GetTransactionAccountMerchantName(merchantAccount)
		transactions = append(transactions, Transaction{
			Id:          data.ReferenceID,
			AccountId:   account.AccountId,
			AmountCents: data.InstructedAmount.Amount,
			Merchant:    merchantName,
			Type:        data.Type,
			Date:        data.TimeStamp,
			Status:      data.Status,
		})
	}

	logger.Debug("Returning transactions", "count", len(transactions))

	return c.JSON(http.StatusOK, transactions)
}
