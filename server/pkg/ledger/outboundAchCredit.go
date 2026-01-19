// Generated from https://apidocs.netxd.com/developers/docs/payment_apis/US%20Domestic%20Payments/ACH/Outbound%20ACH%20Credit/
package ledger

import (
	"fmt"
	"process-api/pkg/clock"
	"process-api/pkg/db/dao"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type OutboundAchCreditRequestTransactionAmount struct {
	Amount   string `json:"amount" validate:"required"`
	Currency string `json:"currency" validate:"required"`
}

type OutboundAchCreditRequestInstitution struct {
	Identification     string `json:"identification" validate:"required"`
	IdentificationType string `json:"identificationType" validate:"required"`
}

type OutboundAchCreditRequestDebtor struct {
	FirstName string `json:"firstName" validate:"required"`
}

type OutboundAchCreditRequestCreditor struct {
	UserType  string `json:"userType" validate:"required"`
	FirstName string `json:"firstName" validate:"required"`
}

type OutboundAchCreditRequestDebtorAccount struct {
	Identification     string                              `json:"identification" validate:"required"`
	IdentificationType string                              `json:"identificationType" validate:"required"`
	Institution        OutboundAchCreditRequestInstitution `json:"institution" validate:"required"`
}

type OutboundAchCreditRequestCreditorAccount struct {
	Identification      string                              `json:"identification" validate:"required"`
	IdentificationType  string                              `json:"identificationType" validate:"required"`
	IdentificationType2 string                              `json:"identificationType2" validate:"required"` // Default: "CHECKING"
	Institution         OutboundAchCreditRequestInstitution `json:"institution" validate:"required"`
}

type OutboundAchCreditRequest struct {
	Channel                string                                    `json:"channel" validate:"required"`
	TransactionType        string                                    `json:"transactionType" validate:"required"`
	Reference              string                                    `json:"reference" validate:"required"`
	TransactionDateTime    string                                    `json:"transactionDateTime" validate:"required"`
	TransactionAmount      OutboundAchCreditRequestTransactionAmount `json:"transactionAmount" validate:"required"`
	Debtor                 OutboundAchCreditRequestDebtor            `json:"debtor" validate:"required"`
	DebtorAccount          OutboundAchCreditRequestDebtorAccount     `json:"debtorAccount" validate:"required"`
	Creditor               OutboundAchCreditRequestCreditor          `json:"creditor" validate:"required"`
	CreditorAccount        OutboundAchCreditRequestCreditorAccount   `json:"creditorAccount" validate:"required"`
	Reason                 *string                                   `json:"reason,omitempty"`
	StandardEntryClassCode *string                                   `json:"standardEntryClassCode,omitempty"`
}

type OutboundAchCreditResult struct {
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
	ProcessID              string `json:"processID"`
}

func BuildOutboundAchCreditRequest(
	user *dao.MasterUserRecordDao,
	accountNumber string,
	routingNumber string,
	amountCents string,
	externalUserFirstName string,
	externalAccountNumber string,
	externalInstitutionNumber string,
	externalAccountType IdentificationType2,
	reason *string,
	secCode *string,
) OutboundAchCreditRequest {
	// NetXD insists that reason is required by NACHA
	if reason == nil || len(*reason) == 0 {
		temp := "ACH Credit"
		reason = &temp
	}

	// Our current user is the Debtor. We are sending out to the Creditor.
	return OutboundAchCreditRequest{
		Channel:                "ACH",
		TransactionType:        "ACH_OUT",
		Reference:              fmt.Sprintf("ledger.ach.transfer_ach_out_%d", clock.Now().UnixNano()),
		TransactionDateTime:    clock.Now().Format("2006-01-02 15:04:05"),
		Reason:                 reason,
		StandardEntryClassCode: secCode,
		TransactionAmount: OutboundAchCreditRequestTransactionAmount{
			Amount:   amountCents,
			Currency: "USD",
		},
		Debtor: OutboundAchCreditRequestDebtor{
			FirstName: user.FirstName,
		},
		DebtorAccount: OutboundAchCreditRequestDebtorAccount{
			Identification:     accountNumber,
			IdentificationType: "ACCOUNT_NUMBER",
			Institution: OutboundAchCreditRequestInstitution{
				Identification:     routingNumber,
				IdentificationType: "ABA",
			},
		},
		Creditor: OutboundAchCreditRequestCreditor{
			UserType:  "INDIVIDUAL",
			FirstName: externalUserFirstName,
		},
		CreditorAccount: OutboundAchCreditRequestCreditorAccount{
			Identification:      externalAccountNumber,
			IdentificationType:  "ACCOUNT_NUMBER",
			IdentificationType2: externalAccountType,
			Institution: OutboundAchCreditRequestInstitution{
				Identification:     externalInstitutionNumber,
				IdentificationType: "ABA", //? Not required by endpoint
			},
		},
	}
}

func (c *NetXDPaymentApiClient) OutboundAchCredit(req OutboundAchCreditRequest) (NetXDApiResponse[OutboundAchCreditResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[OutboundAchCreditResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[OutboundAchCreditResult]
	err := c.call("ledger.ach.transfer", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
