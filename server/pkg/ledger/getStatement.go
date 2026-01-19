// This is API is not yet defined in the web docs
package ledger

import (
	"fmt"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type GetStatementRequest struct {
	Id string `json:"id" validate:"required"`
}

type GetStatementResult struct {
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
	} `json:"legalRepID"`
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

	// Deprecated: use PdfBase64()
	PdfFile *string `json:"pdfFile"`
	// Deprecated: use PdfBase64()
	PdfV2File *string `json:"pdfV2File"`
	// Deprecated: use PdfBase64()
	PdfV3File *string `json:"pdfV3File"`
	// Deprecated: use PdfBase64()
	PdfV4File *string `json:"pdfV4File"`
	// Deprecated: use PdfBase64()
	PdfV5File *string `json:"pdfV5File"`
	// Deprecated: use PdfBase64()
	PdfV6File *string `json:"pdfV6File"`
}

func (result GetStatementResult) PdfBase64() (string, error) {
	if result.PdfFile != nil {
		return *result.PdfFile, nil
	}
	if result.PdfV6File != nil {
		return *result.PdfV6File, nil
	}
	if result.PdfV5File != nil {
		return *result.PdfV5File, nil
	}
	if result.PdfV4File != nil {
		return *result.PdfV4File, nil
	}
	if result.PdfV3File != nil {
		return *result.PdfV3File, nil
	}
	if result.PdfV2File != nil {
		return *result.PdfV2File, nil
	}
	return "", fmt.Errorf("could not find a pdf field in response")
}

func (c *NetXDLedgerApiClient) GetStatement(req GetStatementRequest) (NetXDApiResponse[GetStatementResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[GetStatementResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[GetStatementResult]
	err := c.call("StatementService.GetStatement", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
