// Generated from https://apidocs.netxd.com/developers/docs/payment_apis/Closed%20Loop%20Payments/Provisional%20Credit
package ledger

import (
	"fmt"
	"process-api/pkg/clock"
	"process-api/pkg/db/dao"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type ProvisionalCreditTransactionAmount struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type ProvisionalCreditCreditor struct {
	UserType  string `json:"userType"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type ProvisionalCreditCreditorPostalAddress struct {
	AddressType  string `json:"addressType,omitempty"`
	AddressLine1 string `json:"addressLine1,omitempty"`
	AddressLine2 string `json:"addressLine2,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	ZipCode      string `json:"zipCode,omitempty"`
	CountryCode  string `json:"countryCode,omitempty"`
}

type ProvisionalCreditCreditorContact struct {
	PrimaryEmail string `json:"primaryEmail,omitempty"`
	PrimaryPhone string `json:"primaryPhone,omitempty"`
}

type ProvisionalCreditCreditorAccount struct {
	Identification      string                        `json:"identification"`
	IdentificationType  string                        `json:"identificationType"`
	IdentificationType2 string                        `json:"identificationType2"`
	Institution         *ProvisionalCreditInstitution `json:"institution,omitempty"`
}

type ProvisionalCreditInstitution struct {
	Name               string `json:"name"`
	Identification     string `json:"identification"`
	IdentificationType string `json:"identificationType"`
}

type ProvisionalCreditRequest struct {
	Channel               string                                  `json:"channel"`
	TransactionType       string                                  `json:"transactionType"`
	TransactionDateTime   string                                  `json:"transactionDateTime"`
	Reference             string                                  `json:"reference"`
	Reason                *string                                 `json:"reason"`
	TransactionAmount     ProvisionalCreditTransactionAmount      `json:"transactionAmount"`
	Creditor              *ProvisionalCreditCreditor              `json:"creditor,omitempty"`
	CreditorPostalAddress *ProvisionalCreditCreditorPostalAddress `json:"creditorPostalAddress,omitempty"`
	CreditorContact       *ProvisionalCreditCreditorContact       `json:"creditorContact,omitempty"`
	CreditorAccount       ProvisionalCreditCreditorAccount        `json:"creditorAccount"`
}

type ProvisionalCreditResult struct {
	Api struct {
		Type      string `json:"type"`
		Reference string `json:"reference"`
		DateTime  string `json:"dateTime"`
	} `json:"api"`
	Account struct {
		AccountId        string `json:"accountId"`
		BalanceCents     int64  `json:"balanceCents"`
		HoldBalanceCents int64  `json:"holdBalanceCents"`
		Status           string `json:"status"`
	} `json:"account"`
	TransactionNumber      string `json:"transactionNumber"`
	TransactionStatus      string `json:"transactionStatus"`
	TransactionAmountCents int64  `json:"transactionAmountCents"`
	OriginalRequestBase64  string `json:"originalRequestBase64"`
	ProcessId              string `json:"processId"`
}

func BuildProvisionalCreditRequest(
	user *dao.MasterUserRecordDao,
	amountCents string,
	externalAccountNumber string,
	externalAccountType string,
	reason *string,
) (*ProvisionalCreditRequest, error) {
	return &ProvisionalCreditRequest{
		Channel:             "API",
		TransactionType:     "PROVISIONAL_CREDIT",
		TransactionDateTime: clock.Now().Format("2006-01-02 15:04:05"),
		Reference:           fmt.Sprintf("ledger.paymentv2.provisional_credit_%d", clock.Now().UnixNano()),
		Reason:              reason,
		TransactionAmount: ProvisionalCreditTransactionAmount{
			Amount:   amountCents,
			Currency: "USD",
		},
		// Missing documentation from ledger API. These structs can be skipped, but only if omitted entirely
		// Left commented here for future reference if this API endpoint changes.
		//
		// Creditor: ProvisionalCreditCreditor{
		// 	UserType:  "INDIVIDUAL",
		// 	FirstName: user.FirstName,
		// },
		// CreditorPostalAddress: ProvisionalCreditCreditorPostalAddress{
		// 	AddressLine1: user.StreetAddress,
		// 	AddressLine2: user.ApartmentNo,
		// 	City:         user.City,
		// 	State:        user.State,
		// 	ZipCode:      user.ZipCode,
		// 	CountryCode:  "US",
		// },
		// CreditorContact: ProvisionalCreditCreditorContact{
		// 	PrimaryEmail: user.Email,
		// 	// PrimaryPhone needs the +1 prefix dropped like other ledger phone fields
		// 	PrimaryPhone: mobileNo,
		// },
		CreditorAccount: ProvisionalCreditCreditorAccount{
			Identification:      externalAccountNumber,
			IdentificationType:  "ACCOUNT_NUMBER",
			IdentificationType2: externalAccountType,
		},
	}, nil
}

func (c *NetXDPaymentApiClient) ProvisionalCredit(req ProvisionalCreditRequest) (NetXDApiResponse[ProvisionalCreditResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[ProvisionalCreditResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[ProvisionalCreditResult]
	err := c.call("ledger.transfer", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
