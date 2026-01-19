package ledger

import (
	"fmt"
	"process-api/pkg/clock"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type AchReturnRequest struct {
	Channel                   string `json:"channel" validate:"required"`
	TransactionType           string `json:"transactionType" validate:"required"`
	Reference                 string `json:"reference" validate:"required"`
	TransactionDateTime       string `json:"transactionDateTime" validate:"required"`
	OriginalTransactionNumber string `json:"originalTransactionNumber" validate:"required"`
	// NOTE: R23 is a known acceptable value, if there are unexpected failures the culprit
	// could be unsupported return codes
	ReturnCode string `json:"returnCode" validate:"required"`
	Reason     string `json:"reason" validate:"required"`
}

type AchReturnResult struct {
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
	TransactionStatus      string `json:"transactionStatus"`
	TransactionAmountCents int64  `json:"transactionAmountCents"`
	OriginalRequestBase64  string `json:"originalRequestBase64"`
	ProcessID              string `json:"processID"`
	TransactionNumber      string `json:"transactionNumber"`
	Status                 string `json:"status"`
}

func BuildAchReturnRequest(originalTransactionNumber string, returnCode string, reason string) AchReturnRequest {
	return AchReturnRequest{
		Reference:                 fmt.Sprintf("achreturns_%d", clock.Now().UnixNano()),
		TransactionType:           "ACH_IN_CREDIT_RETURN",
		TransactionDateTime:       clock.Now().Format("2006-01-02 15:04:05"),
		Channel:                   "ACH",
		OriginalTransactionNumber: originalTransactionNumber,
		ReturnCode:                returnCode,
		Reason:                    reason,
	}
}

func (c *NetXDPaymentApiClient) AchReturn(req AchReturnRequest) (NetXDApiResponse[AchReturnResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[AchReturnResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[AchReturnResult]
	err := c.call("ledger.ach.return", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
