// Generated from https://apidocs.netxd.com/developers/docs/card_apis/Card%20Management/PIN%20Change
package ledger

import (
	"fmt"
	"process-api/pkg/clock"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

// TODO: This needs manual clean up since it:
//   - has all parameters, not just payload
//   - does not nest structs properly
type ChangePinRequest struct {
	Reference       string `json:"reference"`
	Product         string `json:"product" validate:"required"`
	Program         string `json:"program" validate:"required"`
	Channel         string `json:"channel" validate:"required"`
	TransactionType string `json:"transactionType" validate:"required"`
	CustomerId      string `json:"customerId" validate:"required"`
	AccountNumber   string `json:"accountNumber" validate:"required"`
	CardId          string `json:"cardId" validate:"required"`
	NewPIN          string `json:"newPIN" validate:"required"`
	IsEncrypt       bool   `json:"isEncrypt" validate:"required"`
}

type ChangePinResult struct {
	Api struct {
		Type              string `json:"type"`
		Reference         string `json:"reference"`
		DateCreated       int64  `json:"dateCreated"`
		OriginalReference string `json:"originalReference"`
	} `json:"api"`
}

func (c *NetXDCardApiConfig) BuildChangePinRequest(customerId string, accountNumber string, cardId string, newPIN string, isUnEncrypted bool) (*ChangePinRequest, error) {
	var encryptedPin string
	var err error

	if isUnEncrypted {
		encryptedPin, err = encryptCardDataForLedger(newPIN, c.publicKey)
		if err != nil {
			return nil, errtrace.Wrap(fmt.Errorf("failed to encrypt newPIN: %w", err))
		}
	} else {
		encryptedPin = newPIN
	}

	return &ChangePinRequest{
		Reference:       fmt.Sprintf("changepin_%s_%d", customerId, clock.Now().UnixNano()),
		Product:         c.cardProduct,
		Channel:         c.cardChannel,
		Program:         c.cardProgram,
		TransactionType: "PIN_CHANGE",
		CustomerId:      customerId,
		AccountNumber:   accountNumber,
		CardId:          cardId,
		NewPIN:          encryptedPin,
		IsEncrypt:       true,
	}, nil
}

func (c *NetXDCardApiClient) ChangePin(req ChangePinRequest) (NetXDApiResponse[ChangePinResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[ChangePinResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[ChangePinResult]
	err := c.call("ledger.CARD.request", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
