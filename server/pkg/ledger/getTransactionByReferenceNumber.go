// Generated from https://apidocs.netxd.com/developers/docs/Transactions/Get%20Transactions%20By%20Reference%20Number/
package ledger

import (
	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type GetTransactionByReferenceNumberRequest struct {
	ReferenceId string `json:"ReferenceId" validate:"required"`
}

type GetTransactionByReferenceNumberResultAccount struct {
	AccountNumber string `json:"accountNumber"`
	Party         struct {
		Name    string `json:"name"`
		Address struct {
			Line1   string `json:"line1"`
			City    string `json:"city"`
			State   string `json:"state"`
			Country string `json:"country"`
			ZipCode string `json:"zipCode"`
		} `json:"address"`
	} `json:"party"`
	InstitutionId   string `json:"institutionId"`
	InstitutionName string `json:"institutionName"`
	CustomerName    string `json:"customerName"`
	CustomerID      string `json:"customerID"`
	Nickname        string `json:"nickName"`
}

type GetTransactionByReferenceNumberResult struct {
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
	DebtorAccount      *GetTransactionByReferenceNumberResultAccount `json:"debtorAccount,omitempty"`
	CreditorAccount    *GetTransactionByReferenceNumberResultAccount `json:"creditorAccount,omitempty"`
	ProcessID          string                                        `json:"processID"`
	Status             string                                        `json:"status"`
	CustomerID         string                                        `json:"customerID"`
	TransactionID      string                                        `json:"transactionID"`
	Credit             bool                                          `json:"credit"`
	AutoFileProcess    bool                                          `json:"autoFileProcess"`
	TokenAppFileUpload bool                                          `json:"tokenAppFileUpload"`
	TransactionNumber  string                                        `json:"transactionNumber"`
}

func BuildGetTransactionByReferenceNumberRequest(referenceId string) GetTransactionByReferenceNumberRequest {
	return GetTransactionByReferenceNumberRequest{
		ReferenceId: referenceId,
	}
}

func (c *NetXDLedgerApiClient) GetTransactionByReferenceNumber(req GetTransactionByReferenceNumberRequest) (NetXDApiResponse[GetTransactionByReferenceNumberResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[GetTransactionByReferenceNumberResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[GetTransactionByReferenceNumberResult]
	err := c.call("TransactionService.GetTransactionsByRef", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
