// undocumented API endpoint. attachment of a curl and screen shot form a two year old email can be found here:
// https://dreamfi.atlassian.net/browse/DT-364
// https://dreamfi.atlassian.net/browse/DT-475
//
// Usage of PCI Check Flag:
//
//	The PCI check flag should be enabled at the customer level in XD Ledger.
//
//	It indicates the system accessing the APIs is eligible to access cardholder details (e.g., PAN, CVV, other card holder data).
//
//	If the flag is enabled, the system allows access to sensitive card data.
//
//	If the flag is disabled, the system restricts access to these details.
//
// DreamFi likely doesn't want to accept the additional compliance requirements of receiving data that most comply with PCI. We're
// likely going to remove this after the ledger API endpoints are updated to not provide the PCI data. Until then, we need it to
// unblock PRs that opperate on the debit card APIs
package ledger

import "braces.dev/errtrace"

type UpdateCustomerSettingsRequest struct {
	CustomerId string `json:"customerId"`
	PciCheck   bool   `json:"pciCheck"`
}

type UpdateCustomerSettingsResponse struct {
	Message string `json:"message"` // validate:"required" Docs say this isn't required... but that's the whole point of this endpoint
}

func (c *NetXDLedgerApiClient) UpdateCustomerSettings(req UpdateCustomerSettingsRequest) (NetXDApiResponse[UpdateCustomerSettingsResponse], error) {
	var response NetXDApiResponse[UpdateCustomerSettingsResponse]
	err := c.call("CustomerService.UpdateCustomerSettings", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
