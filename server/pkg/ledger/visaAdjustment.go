package ledger

import (
	"fmt"
	"process-api/pkg/clock"
	"process-api/pkg/db/dao"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type VisaAdjustmentTransactionAmount struct {
	Amount   string `json:"amount" validate:"required"`
	Currency string `json:"currency" validate:"required"`
}

type VisaAdjustmentDebtor struct {
	FirstName  string `json:"firstName" validate:"required"`
	MiddleName string `json:"_middleName,omitempty"`
	LastName   string `json:"_lastName,omitempty"`
}

type VisaAdjustmentDebtorAccount struct {
	Identification      string `json:"identification" validate:"required"`
	IdentificationType  string `json:"identificationType" validate:"required"`
	IdentificationType2 string `json:"identificationType2" validate:"required"`
}

type VisaAdjustmentRequest struct {
	Channel             string                          `json:"channel" validate:"required"`
	TransactionType     string                          `json:"transactionType" validate:"required"`
	TransactionDateTime string                          `json:"transactionDateTime" validate:"required"`
	Reference           string                          `json:"reference" validate:"required"`
	Reason              string                          `json:"reason" validate:"required"`
	TransactionAmount   VisaAdjustmentTransactionAmount `json:"transactionAmount" validate:"required"`
	Debtor              VisaAdjustmentDebtor            `json:"debtor" validate:"required"`
	DebtorAccount       VisaAdjustmentDebtorAccount     `json:"debtorAccount" validate:"required"`
}

type VisaAdjustmentResult struct {
	Api struct {
		Type      string `json:"type"`
		Reference string `json:"reference"`
		DateTime  string `json:"dateTime"`
	} `json:"api"`
	Account struct {
		AccountId    string `json:"accountId"`
		BalanceCents int64  `json:"balanceCents"`
		Status       string `json:"status"`
	} `json:"account"`
	TransactionNumber      string `json:"transactionNumber"`
	TransactionStatus      string `json:"transactionStatus"`
	TransactionAmountCents int64  `json:"transactionAmountCents"`
	OriginalRequestBase64  string `json:"originalRequestBase64"`
	ProcessId              string `json:"processId"`
}

func BuildVisaAdjustmentRequest(
	user *dao.MasterUserRecordDao,
	accountNumber string,
	amountCents string,
	reference string,
	reason string,
) VisaAdjustmentRequest {
	if reference == "" {
		reference = fmt.Sprintf("visa_adjustment_%d", clock.Now().UnixNano())
	}
	return VisaAdjustmentRequest{
		Channel:             "API",
		TransactionType:     "VISA_ADJUSTMENT",
		TransactionDateTime: clock.Now().Format("2006-01-02 15:04:05"),
		Reference:           reference,
		Reason:              reason,
		TransactionAmount: VisaAdjustmentTransactionAmount{
			Amount:   amountCents,
			Currency: "USD",
		},
		Debtor: VisaAdjustmentDebtor{
			FirstName:  user.FirstName,
			MiddleName: "",
			LastName:   user.LastName,
		},
		DebtorAccount: VisaAdjustmentDebtorAccount{
			Identification:      accountNumber,
			IdentificationType:  "ACCOUNT_NUMBER",
			IdentificationType2: "CHECKING",
		},
	}
}

func (c *NetXDPaymentApiClient) VisaAdjustment(req VisaAdjustmentRequest) (NetXDApiResponse[VisaAdjustmentResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[VisaAdjustmentResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[VisaAdjustmentResult]
	err := c.call("ledger.ach.transfer", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
