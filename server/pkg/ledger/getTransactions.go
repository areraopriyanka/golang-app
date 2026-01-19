// Generated from https://apidocs.netxd.com/developers/docs/Transactions/List%20Transaction%20by%20Account/
package ledger

import (
	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

/* import (
	"github.com/go-playground/validator/v10"
) */

type GetTransactionsRequest struct {
	Page          uint   `json:"page"`
	Size          uint   `json:"size"`
	FromDate      string `json:"fromDate"`      // Formatted as YYYY-MM-DD
	ToDate        string `json:"toDate"`        // Formatted as YYYY-MM-DD
	AccountNumber string `json:"accountNumber"` // validate:"required"`
}

type GetTransactionsResultAccount struct {
	AccountLevel  string `json:"accountLevel"`
	AccountNumber string `json:"accountNumber"`
	Address       struct {
		City    string `json:"city"`
		Country string `json:"country"`
		Line1   string `json:"line1"`
		Line2   string `json:"line2"`
		State   string `json:"state"`
		ZipCode string `json:"zipCode"`
	} `json:"address"`
	CustomerAccountName   string `json:"customerAccountName"`
	CustomerAccountNumber string `json:"customerAccountNumber"`
	CustomerAccountType   string `json:"customerAccountType"`
	CustomerNumber        string `json:"customerNumber"`
	HolderId              string `json:"holderId"`
	HolderIdType          string `json:"holderIdType"`
	HolderName            string `json:"holderName"`
	InstitutionId         string `json:"institutionId"`
	InstitutionName       string `json:"institutionName"`
	Name                  string `json:"name"`
	NickName              string `json:"nickName"`
	Party                 struct {
		Contact struct {
			Email       string `json:"email"`
			PhoneNumber string `json:"phoneNumber"`
		} `json:"contact"`
	} `json:"party"`
	Reference string `json:"reference"`
}

type GetTransactionsResultTransaction struct {
	AibCreated            bool                         `json:"aibCreated"`
	AibRequired           bool                         `json:"aibRequired"`
	AvalBalance           int64                        `json:"avalBalance"`
	CeReject              bool                         `json:"CeReject"`
	Channel               string                       `json:"channel"`
	CreatedDate           string                       `json:"createdDate"`
	EftBatchID            string                       `json:"eftBatchID"`
	EftBatchStatus        string                       `json:"eftBatchStatus"`
	EftReported           bool                         `json:"eftReported"`
	ErpPosted             bool                         `json:"erpPosted"`
	EscrowReported        bool                         `json:"escrowReported"`
	Id                    string                       `json:"id"`
	InstructedAmount      int64                        `json:"instructedAmount"`
	InstructedCurrency    string                       `json:"instructedCurrency"`
	IsFedExported         bool                         `json:"isFedExported"`
	IsLegalRepApproved    bool                         `json:"isLegalRepApproved"`
	IsPartial             bool                         `json:"isPartial"`
	IsReturned            bool                         `json:"isReturned"`
	IsWireReturned        bool                         `json:"isWireReturned"`
	NachaForwarding       bool                         `json:"nachaForwarding"`
	PostedDate            string                       `json:"postedDate"`
	PreAuthExpiry         string                       `json:"preAuthExpiry"`
	PreAuthHoldRelease    bool                         `json:"preAuthHoldRelease"`
	ProcessId             string                       `json:"processId"`
	Product               string                       `json:"product"`
	Reason                string                       `json:"reason"`
	ReferenceID           string                       `json:"referenceID"`
	Source                string                       `json:"source"`
	Status                string                       `json:"status"`
	SubTransactionType    string                       `json:"subTransactionType"`
	ToCustomerAvalBalance int64                        `json:"toCustomerAvalBalance"`
	TransactionNumber     string                       `json:"transactionNumber"`
	TransactionType       string                       `json:"transactionType"`
	UpdatedDate           string                       `json:"updatedDate"`
	WpsReported           bool                         `json:"wpsReported"`
	CreditorAccount       GetTransactionsResultAccount `json:"creditorAccount"`
	DebtorAccount         GetTransactionsResultAccount `json:"debtorAccount"`
	NameScreen            struct {
		WatchListStatus string `json:"watchListStatus"`
	} `json:"nameScreen"`
}

type GetTransactionsResult struct {
	TotalDocs    int64                              `json:"totalDocs"`
	Transactions []GetTransactionsResultTransaction `json:"transactions"`
}

func (c *NetXDLedgerApiClient) GetTransactions(req GetTransactionsRequest) (NetXDApiResponse[GetTransactionsResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[GetTransactionsResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[GetTransactionsResult]
	err := c.call("TransactionService.GetTransactions", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
