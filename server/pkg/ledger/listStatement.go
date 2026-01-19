// This is API is not yet defined in the web docs
package ledger

import (
	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

// As per ledger doc mentioned in https://dreamfi.atlassian.net/browse/DT-868 json fields are pageNumber and pageSize
type ListStatementRequest struct {
	PageNumber int `json:"pageNumber" validate:"required"`
	PageSize   int `json:"pageSize" validate:"required"`
}

type ListStatementResult struct {
	Accounts []struct {
		Id            string `json:"id"`
		CreatedDate   string `json:"createdDate"`
		UpdatedDate   string `json:"updatedDate"`
		CustomerID    string `json:"customerId"`
		AccountId     string `json:"accountId"`
		Md5           string `json:"md5"`
		AccountNumber string `json:"accountNumber"`
		AccountName   string `json:"accountName"`
		CustomerName  string `json:"customerName"`
		LegalRepId    []struct {
			Id   string `json:"ID"`
			Name string `json:"name"`
		} `json:"legalRepId"`
		ClosingBalanceCents int64  `json:"closingBalanceCents"`
		DebitVolume         int64  `json:"debitVolume"`
		DebitCount          int32  `json:"debitCount"`
		CreditVolume        int64  `json:"creditVolume"`
		CreditCount         int32  `json:"creditCount"`
		TotalRecords        int32  `json:"totalRecords"`
		TotalAmmount        int64  `json:"totalAmmount"`
		Currency            string `json:"currency"`
		Month               string `json:"month"`
		Year                int32  `json:"year"`
		LastDate            string `json:"lastDate"`
		AverageVolume       int64  `json:"AverageVolume"`
		FileType            string `json:"fileType"`
	} `json:"statements"`
	TotalCounts int32 `json:"totalCounts"`
}

func (c *NetXDLedgerApiClient) ListStatement(req ListStatementRequest) (NetXDApiResponse[ListStatementResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[ListStatementResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[ListStatementResult]
	err := c.call("StatementService.ListStatement", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
