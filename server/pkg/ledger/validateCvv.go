// Generated from https://apidocs.netxd.com/developers/docs/card_apis/Card%20Management/Validate%20CVV
package ledger

import (
	"fmt"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type ValidateCvvRequest struct {
	TransactionType string `json:"transactionType" validate:"required"`
	Product         string `json:"product" validate:"required"`
	Channel         string `json:"channel" validate:"required"`
	Program         string `json:"program" validate:"required"`
	CardId          string `json:"cardId" validate:"required"`
	CardCvv         string `json:"cardCvv" validate:"required"`
	IsEncrypt       bool   `json:"isEncrypt" validate:"required"`
}

type ValidateCvvResultApi struct {
	Type        string `json:"type"`
	Reference   string `json:"reference"`
	DateCreated int64  `json:"dateCreated"`
}

type ValidateCvvResult struct {
	CardId  string               `json:"cardId"`
	Api     ValidateCvvResultApi `json:"api"`
	Message string               `json:"message"`
}

func (c *NetXDCardApiConfig) BuildValidateCvvRequest(cardId string, cardCvv string, isUnEncrypted bool) (*ValidateCvvRequest, error) {
	var encryptedCvv string
	var err error
	if isUnEncrypted {
		encryptedCvv, err = encryptCardDataForLedger(cardCvv, c.publicKey)
		if err != nil {
			return nil, errtrace.Wrap(fmt.Errorf("failed to encrypted cvv: %w", err))
		}
	} else {
		encryptedCvv = cardCvv
	}

	return &ValidateCvvRequest{
		Product:         c.cardProduct,
		Channel:         c.cardChannel,
		Program:         c.cardProgram,
		TransactionType: "VALIDATE_CVV",
		CardId:          cardId,
		CardCvv:         encryptedCvv,
		IsEncrypt:       true,
	}, nil
}

const (
	ValidateCvvValidType   = "VALIDATE_CVV_ACK"
	ValidateCvvInvalidType = "VALIDATE_CVV_NACK"
)

func (c *NetXDCardApiClient) ValidateCvv(req ValidateCvvRequest) (NetXDApiResponse[ValidateCvvResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[ValidateCvvResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[ValidateCvvResult]
	err := c.call("ledger.CARD.request", c.url, req, &response)

	// ledger returns "INCORRECT CARD CVV" error message and "1019" error code in case of invalid cvv
	// A success with a different type seems clearer
	if response.Error != nil && response.Error.Code == "1019" {
		return NetXDApiResponse[ValidateCvvResult]{
			Result: &ValidateCvvResult{
				CardId:  req.CardId,
				Message: response.Error.Message,
				Api: ValidateCvvResultApi{
					Type:        ValidateCvvInvalidType,
					Reference:   "REF_internal",
					DateCreated: 0,
				},
			},
		}, nil
	}
	// HACK HACK HACK
	// The ledger is returning internal error 9999 for invalid cvv
	// this should be removed after NetXD addresses the issue:
	// https://dreamfi.atlassian.net/browse/DT-778
	if response.Error != nil && response.Error.Code == "9999" {
		return NetXDApiResponse[ValidateCvvResult]{
			Result: &ValidateCvvResult{
				CardId:  req.CardId,
				Message: response.Error.Message,
				Api: ValidateCvvResultApi{
					Type:        ValidateCvvInvalidType,
					Reference:   "REF_internal",
					DateCreated: 0,
				},
			},
		}, nil
	}

	return response, errtrace.Wrap(err)
}
