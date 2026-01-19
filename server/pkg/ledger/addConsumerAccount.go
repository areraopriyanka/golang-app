// Generated from https://apidocs.netxd.com/developers/docs/account_apis/AddAccount%20Consumer
package ledger

import (
	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type AddConsumerAccountRequest struct {
	CustomerID            string `json:"customerID" validate:"required"`
	Name                  string `json:"name" validate:"required"`
	AccountType           string `json:"accountType" validate:"required,oneof=SAVINGS CHECKING"`
	Currency              string `json:"currency" validate:"required"`
	ActivityAccountNumber string `json:"activityAccountNumber"`
	// undocumented field
	AccountCategory string `json:"AccountCategory"`
}

type AddConsumerAccountResult struct {
	ID            string `json:"ID"`
	Status        string `json:"status"`
	AccountNumber string `json:"accountNumber"`
	AccountType   string `json:"accountType"`
	InstitutionID string `json:"institutionID"`
	CustomerID    string `json:"customerID"`
}

func (c *NetXDLedgerApiClient) BuildAddConsumerAccountRequest(ledgerCustomerNumber string, accountType string, accountName string) AddConsumerAccountRequest {
	return AddConsumerAccountRequest{
		CustomerID:  ledgerCustomerNumber,
		Name:        accountName,
		AccountType: accountType,
		Currency:    "USD", // hardcoding this for now, as we assume USD
		// This field is undocumented by NetXD. Omitting it results in a 200 response with an Error:
		//     "code": "BAD_INPUT"
		//     "message": "ActivityAccountNumber is invalid or missing"
		AccountCategory: c.ledgerCategory,
	}
}

func (c *NetXDLedgerApiClient) AddConsumerAccount(req AddConsumerAccountRequest) (NetXDApiResponse[AddConsumerAccountResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[AddConsumerAccountResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[AddConsumerAccountResult]
	err := c.call("CustomerService.AddAccount", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
