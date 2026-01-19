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

// @Summary GetAccountClosureStatus
// @Description Returns account and card details and indicates whether an account is closable
// @Tags cards
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} AccountClosureStatus
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Router /account/closure-status [get]
func GetAccountClosureStatus(c echo.Context) error {
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

	cardHolder, errResponse := dao.RequireCardHolderForUser(userId)
	if errResponse != nil {
		return errResponse
	}

	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDLedgerApiClient(config.Config.Ledger, ledgerParamsBuilder)

	getCustomerPayload := ledger.BuildGetCustomerByCustomerNoPayload(user.LedgerCustomerNumber)
	getCustomerResponse, err := ledgerClient.GetCustomer(*getCustomerPayload)
	if err != nil {
		logger.Error("Error while calling ListAccounts while checking account closure status", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error while calling ListAccounts while checking account closure status: error: %s", err.Error()), errtrace.Wrap(err))
	}
	if getCustomerResponse.Error != nil {
		logger.Error("Error from ledger ListAccounts while checking account closure status", "error", getCustomerResponse.Error)
		return response.InternalServerError("Error from ledger ListAccounts while checking account closure status: error", errtrace.New(""))
	}

	listTransactionsPayload := ledger.BuildListTransactionsByAccountPayload(cardHolder.AccountNumber)
	listTransactionsResponse, err := ledgerClient.ListTransactionsByAccount(listTransactionsPayload)
	if err != nil {
		logger.Error("Error while calling ListTransactionsByAccount while checking account pending transactions", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error while calling ListTransactionsByAccount while checking account pending transactions: error: %s", err.Error()), errtrace.Wrap(err))
	}
	if listTransactionsResponse.Error != nil {
		logger.Error("Error from ledger ListTransactionsByAccount while checking account pending transactions", "error", listTransactionsResponse.Error)
		return response.InternalServerError("Error from ledger ListTransactionsByAccount while checking account pending transactions: error", errtrace.New(""))
	}
	if listTransactionsResponse.Result == nil {
		logger.Error("The ledger responded with an empty result object", "responseData", listTransactionsResponse)
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: "The ledger responded with an empty result object", MaybeInnerError: errtrace.New("")}
	}

	finalTransactions, _ := MergeTransactions(listTransactionsResponse.Result.AccountTransactions)

	var pendingTransactions uint
	for _, transaction := range finalTransactions {
		if transaction.Type == "PRE_AUTH" {
			pendingTransactions++
		}
	}

	accountBalance := getCustomerResponse.Result.Accounts[0].Balance
	preAuthBalance := getCustomerResponse.Result.Accounts[0].PreAuthBalance
	holdBalance := getCustomerResponse.Result.Accounts[0].HoldBalance
	isAccountClosable := accountBalance == 0 && preAuthBalance == 0 && holdBalance == 0 && pendingTransactions == 0

	return cc.JSON(http.StatusOK, AccountClosureStatus{
		IsAccountClosable:   isAccountClosable,
		AccountBalance:      accountBalance,
		PreAuthBalance:      preAuthBalance,
		HoldBalance:         holdBalance,
		PendingTransactions: pendingTransactions,
	})
}

// NOTE: We are waiting for NetXD to implement the return of fields from the ledger that
// indicate Pending Credits and Pending Debits. In the future we can return those here
type AccountClosureStatus struct {
	IsAccountClosable   bool    `json:"isAccountClosable" validate:"required"`
	AccountBalance      float64 `json:"accountBalance" validate:"required" mask:"true"`
	PreAuthBalance      float64 `json:"preAuthBalance" validate:"required" mask:"true"`
	HoldBalance         float64 `json:"holdBalance" validate:"required" mask:"true"`
	PendingTransactions uint    `json:"pendingTransactions" validate:"required" mask:"true"`
}
