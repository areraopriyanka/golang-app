// Generated from https://apidocs.netxd.com/developers/docs/customer_apis/Update_Customer_-_Consumer/
package ledger

import (
	"process-api/pkg/db/dao"
	"process-api/pkg/validators"

	"braces.dev/errtrace"
)

type UpdateCustomerRequest struct {
	CustomerId string  `json:"ID" validate:"required"`
	Type       string  `json:"type"`
	DOB        string  `json:"DOB"`
	Title      string  `json:"title"`
	Address    Address `json:"Address"`
	Contact    Contact `json:"Contact"`
	FirstName  string  `json:"firstName" validate:"required"`
	LastName   string  `json:"lastName" validate:"required"`
}

type UpdateCustomerResult struct {
	ID             string `json:"ID"`
	Type           string `json:"type"`
	Identification []struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"identification"`
	Contact struct {
		Email       string `json:"email"`
		PhoneNumber string `json:"phoneNumber"`
	} `json:"contact"`
	Address struct {
		AddressLine1 string `json:"addressLine1"`
		City         string `json:"city"`
		State        string `json:"state"`
		Country      string `json:"country"`
		Zip          string `json:"zip"`
	} `json:"address"`
	DOB         string `json:"DOB"`
	Title       string `json:"title"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	CreatedDate string `json:"createdDate"`
	UpdatedDate string `json:"updatedDate"`
	Status      string `json:"status"`
}

func BuildUpdateCustomerRequest(user dao.MasterUserRecordDao) UpdateCustomerRequest {
	return UpdateCustomerRequest{
		CustomerId: user.LedgerCustomerNumber,
		Type:       "INDIVIDUAL",
		DOB:        user.DOB.Format("20060102"),
		Title:      "Ms",
		// According to https://apidocs.netxd.com/developers/docs/customer_apis/Update_Customer_-_Consumer/ address fields are optional
		// But, skipping them causes errors when not updating the address
		// Reusing the existing Address struct with its validations
		Address: Address{
			AddressLine1: user.StreetAddress,
			City:         user.City,
			State:        user.State,
			Country:      "US",
			ZIP:          user.ZipCode,
		},
		Contact: Contact{
			Email:       user.Email,
			PhoneNumber: user.MobileNo,
		},
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
}

func (c *NetXDLedgerApiClient) UpdateCustomer(req UpdateCustomerRequest) (NetXDApiResponse[UpdateCustomerResult], error) {
	if err := validators.ValidateStruct(req); err != nil {
		return NetXDApiResponse[UpdateCustomerResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[UpdateCustomerResult]
	err := c.call("CustomerService.UpdateCustomer", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
