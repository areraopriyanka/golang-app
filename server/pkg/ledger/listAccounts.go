// Generated from https://apidocs.netxd.com/developers/docs/account_apis/Get%20All%20Accounts/
// Note: AccountService.AccountList is currently not functioning.
// Using "CustomerService.GetCustomer" to fetch the account list instead
package ledger

import (
	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type GetAllAccountsRequest struct {
	Method     string `json:"method" validate:"required"`
	Id         string `json:"id" validate:"required"`
	Params     string `json:"params" validate:"required"`
	Api        string `json:"api" validate:"required"`
	Signature  string `json:"signature" validate:"required"`
	KeyId      string `json:"keyId" validate:"required"`
	Credential string `json:"credential" validate:"required"`
	Payload    string `json:"payload" validate:"required"`
	PageNumber int64  `json:"PageNumber"`
	PageSize   int64  `json:"PageSize"`
	Filter     string `json:"filter"`
}

type GetAllAccountsResult struct {
	Accounts []struct {
		Id              string `json:"id"`
		Name            string `json:"name"`
		Number          string `json:"number"`
		CreatedDate     string `json:"createdDate"`
		UpdatedDate     string `json:"updatedDate"`
		Balance         int64  `json:"balance"`
		HoldBalance     int64  `json:"holdBalance"`
		CustomerID      string `json:"customerID"`
		CustomerName    string `json:"customerName"`
		AccountCategory string `json:"accountCategory"`
		AccountType     string `json:"accountType"`
		Currency        string `json:"currency"`
		CurrencyCode    string `json:"currencyCode"`
		Status          string `json:"status"`
		InstitutionID   string `json:"institutionID"`
		GlAccount       string `json:"glAccount"`
		IsVerify        bool   `json:"isVerify"`
		LedgerBalance   int64  `json:"ledgerBalance"`
		PreAuthBalance  int64  `json:"preAuthBalance"`
	} `json:"accounts"`
	RiskScore int64 `json:"riskScore"`
}

func (c *NetXDLedgerApiClient) GetAllAccounts(req GetAllAccountsRequest) (NetXDApiResponse[GetAllAccountsResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[GetAllAccountsResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[GetAllAccountsResult]
	err := c.call("AccountService.ListAccounts", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
