// Generated from https://apidocs.netxd.com/developers/docs/Transactions/List%20Transaction%20by%20Account/
package ledger

import (
	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

/* import (
	"github.com/go-playground/validator/v10"
) */

type ListTransactionsByAccountRequest struct {
	AccountNumber string `json:"accountNumber" validate:"required"`
}

type ListTransactionsByAccountResultTransactionAccount struct {
	AccountNumber string `json:"accountNumber"`
	Party         struct {
		Name string `json:"name"`
	} `json:"party"`
	InstitutionId   string `json:"institutionId"`
	InstitutionName string `json:"institutionName"`
	CustomerName    string `json:"customerName"`
	CustomerID      string `json:"customerID"`
	Nickname        string `json:"nickName"`
}

type ListTransactionsByAccountResultTransaction struct {
	Type             string `json:"type"`
	ReferenceID      string `json:"ReferenceID"`
	TimeStamp        string `json:"timeStamp"`
	InstructedAmount struct {
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
	} `json:"instructedAmount"`
	AvailableBalance struct {
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
	} `json:"availableBalance"`
	HoldBalance struct {
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
	} `json:"holdBalance"`
	LedgerBalance struct {
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
	} `json:"ledgerBalance"`
	Mcc                    string                                            `json:"mcc"`
	DebtorAccount          ListTransactionsByAccountResultTransactionAccount `json:"debtorAccount"`
	CreditorAccount        ListTransactionsByAccountResultTransactionAccount `json:"creditorAccount"`
	ProcessID              string                                            `json:"processID"`
	Status                 string                                            `json:"status"`
	CustomerID             string                                            `json:"customerID"`
	TransactionID          string                                            `json:"transactionID"`
	Credit                 bool                                              `json:"credit"`
	AutoFileProcess        bool                                              `json:"autoFileProcess"`
	TokenAppFileUpload     bool                                              `json:"tokenAppFileUpload"`
	TransactionNumber      string                                            `json:"transactionNumber"`
	Reason                 *string                                           `json:"reason,omitempty"`
	CardAcceptor           string                                            `json:"cardAcceptor"`
	TransactionTypeDetails string                                            `json:"transactionTypeDetails"`
}

type ListTransactionsByAccountResult struct {
	TotalDocs           int64                                        `json:"totalDocs"`
	AccountTransactions []ListTransactionsByAccountResultTransaction `json:"accountTransactions"`
}

func BuildListTransactionsByAccountPayload(accountNumber string) ListTransactionsByAccountRequest {
	payload := ListTransactionsByAccountRequest{
		AccountNumber: accountNumber,
	}
	return payload
}

const ListTransactionsEmptyError = "NOT_FOUND_TRANSACTION_ENTRIES"

func (c *NetXDLedgerApiClient) ListTransactionsByAccount(req ListTransactionsByAccountRequest) (NetXDApiResponse[ListTransactionsByAccountResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[ListTransactionsByAccountResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[ListTransactionsByAccountResult]
	err := c.call("TransactionService.ListTransactions", c.url, req, &response)

	if response.Error != nil && response.Error.Code == ListTransactionsEmptyError {
		response.Error = nil
		response.Result = &ListTransactionsByAccountResult{
			TotalDocs: 0,
		}
	}

	return response, errtrace.Wrap(err)
}
