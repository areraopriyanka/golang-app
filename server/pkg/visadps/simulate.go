package visadps

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func (c *VisaDPSClient) SimulateAuthorization(cardExternalId string, amountUSD float32) (*AuthorizationSimulatorResponse, error) {
	body := AuthorizationSimulatorRequest{
		BillingAmount: BillingAmount{
			Amount: amountUSD,
		},
		PaymentInstrument: PaymentInstrument{
			CardId: &cardExternalId,
		},
		TransactionType: TCTECOMMERCE,
	}
	authResponse, err := mleRequest[CommonResponseAuthorizationSimulatorResponse](c, body, func(bodyReader io.Reader) (*http.Response, error) {
		return c.AuthSimulatorWithBody(context.Background(), nil, "application/json", bodyReader)
	})
	if err != nil {
		return nil, fmt.Errorf("mle call to simulate authorization failed: %w", err)
	}
	if authResponse.Error != nil {
		return nil, fmt.Errorf("authorization API response error: %+v", authResponse.Error)
	}
	return authResponse.Resource, nil
}

func (c *VisaDPSClient) SimulateClearing(originalTransactionId string, amountUSD float32) (*ClearingSimulatorResponse, error) {
	body := ClearingSimulatorRequest{
		OriginalTransactionId: originalTransactionId,
		BillingAmount: BillingAmount{
			Amount: amountUSD,
		},
	}
	clearingResponse, err := mleRequest[CommonResponseClearingSimulatorResponse](c, body, func(bodyReader io.Reader) (*http.Response, error) {
		return c.ClearingSimulatorWithBody(context.Background(), nil, "application/json", bodyReader)
	})
	if err != nil {
		return nil, fmt.Errorf("mle call to simulate clearing failed: %w", err)
	}
	if clearingResponse.Error != nil {
		return nil, fmt.Errorf("clearing API response error: %+v", clearingResponse.Error)
	}
	return clearingResponse.Resource, nil
}
