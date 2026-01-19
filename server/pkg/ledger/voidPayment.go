// Generated from https://apidocs.netxd.com/developers/docs/payment_apis/Closed%20Loop%20Payments/Void%20Payment
package ledger

import (
	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type VoidPaymentRequest struct {
	Type        string `json:"type"`
	Product     string `json:"product"`
	Program     string `json:"program"`
	ReferenceID string `json:"referenceId"`
	CustomerID  string `json:"customerId"`
	Notes       string `json:"notes"`
}

type VoidPaymentResult struct {
	Status        string `json:"status"`
	TransactionID string `json:"TransactionID"`
	ReferenceID   string `json:"referenceID"`
	IsPartial     bool   `json:"isPartial"`
}

func BuildVoidPaymentRequest(originalTransactionReferenceID string, customerID string, notes string) VoidPaymentRequest {
	return VoidPaymentRequest{
		Type:        "VOID",
		Product:     "DEFAULT",
		Program:     "DEFAULT",
		ReferenceID: originalTransactionReferenceID,
		CustomerID:  customerID,
		Notes:       notes,
	}
}

func (c *NetXDLedgerApiClient) VoidPayment(req VoidPaymentRequest) (NetXDApiResponse[VoidPaymentResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[VoidPaymentResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[VoidPaymentResult]
	err := c.call("TransactionService.Payment", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
