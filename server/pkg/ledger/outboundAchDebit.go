// Generated from https://apidocs.netxd.com/developers/docs/payment_apis/US%20Domestic%20Payments/ACH/Outbound%20ACH%20Debit
package ledger

import (
	"fmt"
	"process-api/pkg/clock"
	"process-api/pkg/db/dao"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type OutboundAchDebitRequestTransactionAmount struct {
	Amount   string `json:"amount" validate:"required"`
	Currency string `json:"currency" validate:"required"`
}

type OutboundAchDebitRequestInstitution struct {
	Identification     string `json:"identification" validate:"required"`
	IdentificationType string `json:"identificationType" validate:"required"`
}

type OutboundAchDebitRequestDebtor struct {
	FirstName string `json:"firstName" validate:"required"`
}

type OutboundAchDebitRequestCreditor struct {
	UserType  string `json:"userType" validate:"required"`
	FirstName string `json:"firstName" validate:"required"`
}

type OutboundAchDebitRequestDebtorAccount struct {
	Identification      string                             `json:"identification" validate:"required"`
	IdentificationType  string                             `json:"identificationType" validate:"required"`
	IdentificationType2 string                             `json:"identificationType2" validate:"required"` // Default: "CHECKING"
	Institution         OutboundAchDebitRequestInstitution `json:"institution" validate:"required"`
}

type OutboundAchDebitRequestCreditorAccount struct {
	Identification     string `json:"identification" validate:"required"`
	IdentificationType string `json:"identificationType" validate:"required"`
}

type OutboundAchDebitRequest struct {
	Channel                string                                   `json:"channel" validate:"required"`
	TransactionType        string                                   `json:"transactionType" validate:"required"`
	Reference              string                                   `json:"reference" validate:"required"`
	TransactionDateTime    string                                   `json:"transactionDateTime" validate:"required"`
	TransactionAmount      OutboundAchDebitRequestTransactionAmount `json:"transactionAmount" validate:"required"`
	Debtor                 OutboundAchDebitRequestDebtor            `json:"debtor" validate:"required"`
	DebtorAccount          OutboundAchDebitRequestDebtorAccount     `json:"debtorAccount" validate:"required"`
	Creditor               OutboundAchDebitRequestCreditor          `json:"creditor" validate:"required"`
	CreditorAccount        OutboundAchDebitRequestCreditorAccount   `json:"creditorAccount" validate:"required"`
	Reason                 *string                                  `json:"reason,omitempty"`
	StandardEntryClassCode *string                                  `json:"standardEntryClassCode,omitempty"`
}

type OutboundAchDebitResult struct {
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

func BuildOutboundAchDebitRequest(
	user *dao.MasterUserRecordDao,
	accountNumber string,
	amountCents string,
	externalUserFirstName string,
	externalAccountNumber string,
	externalInstitutionNumber string,
	externalAccountType IdentificationType2,
	reason *string,
	secCode *string,
) OutboundAchDebitRequest {
	// NetXD insists that reason is required by NACHA
	if reason == nil || len(*reason) == 0 {
		temp := "ACH Debit"
		reason = &temp
	}

	// Our current user is the Creditor. We are pulling in from the Debtor.
	return OutboundAchDebitRequest{
		Channel:                "ACH",
		TransactionType:        "ACH_PULL",
		Reference:              fmt.Sprintf("ledger.ach.transfer_ach_pull_%d", clock.Now().UnixNano()),
		TransactionDateTime:    clock.Now().Format("2006-01-02 15:04:05"),
		Reason:                 reason,
		StandardEntryClassCode: secCode,
		TransactionAmount: OutboundAchDebitRequestTransactionAmount{
			Amount:   amountCents,
			Currency: "USD",
		},
		Debtor: OutboundAchDebitRequestDebtor{
			FirstName: externalUserFirstName,
		},
		DebtorAccount: OutboundAchDebitRequestDebtorAccount{
			Identification:      externalAccountNumber,
			IdentificationType:  "ACCOUNT_NUMBER",
			IdentificationType2: externalAccountType,
			Institution: OutboundAchDebitRequestInstitution{
				Identification:     externalInstitutionNumber,
				IdentificationType: "ABA",
			},
		},
		Creditor: OutboundAchDebitRequestCreditor{
			UserType:  "INDIVIDUAL",
			FirstName: user.FirstName,
		},
		CreditorAccount: OutboundAchDebitRequestCreditorAccount{
			Identification:     accountNumber,
			IdentificationType: "ACCOUNT_NUMBER",
		},
	}
}

func (c *NetXDPaymentApiClient) OutboundAchDebit(req OutboundAchDebitRequest) (NetXDApiResponse[OutboundAchDebitResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[OutboundAchDebitResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[OutboundAchDebitResult]
	err := c.call("ledger.ach.transfer", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
