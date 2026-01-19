// Generated from https://apidocs.netxd.com/developers/docs/card_apis/Card%20Management/Update%20Status
package ledger

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"process-api/pkg/clock"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

// Status Actions and Updated Statuses for card
const (
	// Status Actions
	ACTIVATE_CARD      = "ACTIVATE"
	REPORT_LOST_STOLEN = "REPORT_LOST_STOLEN"
	LOCK               = "LOCK"
	UNLOCK             = "UNLOCK"
	UNBLOCK            = "UNBLOCK"
	CLOSE              = "CLOSE"

	// Updated Statuses
	LOST_STOLEN                = "LOST_STOLEN"
	ACTIVATED                  = "ACTIVATED"
	TEMPRORY_BLOCKED_BY_CLIENT = "TEMPRORY_BLOCKED_BY_CLIENT"
)

type UpdateStatusRequest struct {
	Reference       string `json:"reference"`
	Product         string `json:"product" validate:"required"`
	Program         string `json:"program" validate:"required"`
	Channel         string `json:"channel" validate:"required"`
	TransactionType string `json:"transactionType" validate:"required"`
	CustomerId      string `json:"customerId" validate:"required"`
	CardId          string `json:"cardId" validate:"required"`
	AccountNumber   string `json:"accountNumber" validate:"required"`
	StatusAction    string `json:"statusAction" validate:"required"`
	// CardCvv is required only if statusAction is "ACTIVATE"; otherwise, it can be omitted
	CardCvv   string `json:"cardCvv" validate:"required_if=StatusAction ACTIVATE"`
	IsEncrypt bool   `json:"isEncrypt" validate:"required"`
}

type UpdateStatusResult struct {
	Card struct {
		CardId                 string `json:"cardId"`
		PostedDate             string `json:"postedDate"`
		UpdatedDate            string `json:"updatedDate"`
		CardStatus             string `json:"cardStatus"`
		AllowAtm               bool   `json:"allowAtm"`
		AllowEcommerce         bool   `json:"allowEcommerce"`
		AllowMoto              bool   `json:"allowMoto"`
		AllowPos               bool   `json:"allowPos"`
		AllowTips              bool   `json:"allowTips"`
		AllowPurchase          bool   `json:"allowPurchase"`
		AllowRefund            bool   `json:"allowRefund"`
		AllowCashback          bool   `json:"allowCashback"`
		AllowWithdraw          bool   `json:"allowWithdraw"`
		AllowAuthAndCompletion bool   `json:"allowAuthAndCompletion"`
		Smart                  bool   `json:"smart"`
		CheckAvsZip            bool   `json:"checkAvsZip"`
		CheckAvsAddr           bool   `json:"checkAvsAddr"`
		Cvv                    string `json:"cvv"`
		TransactionMade        bool   `json:"transactionMade"`
		IsReIssue              bool   `json:"isReIssue"`
		IsReplace              bool   `json:"isReplace"`
	} `json:"card"`
	Api struct {
		Type              string `json:"type"`
		Reference         string `json:"reference"`
		DateCreated       int64  `json:"dateCreated"`
		OriginalReference string `json:"originalReference"`
	} `json:"api"`
}

func (c *NetXDCardApiConfig) BuildUpdateStatusRequest(customerId string, cardId string, accountNumber string, statusAction string, cardCvv string, isUnEncrypted bool) (*UpdateStatusRequest, error) {
	var encryptedCvv string
	var err error
	// If cvv is empty then we can skip cvv encryption
	if cardCvv != "" {
		if isUnEncrypted {
			if encryptedCvv, err = encryptCardDataForLedger(cardCvv, c.publicKey); err != nil {
				return nil, errtrace.Wrap(fmt.Errorf("failed to encrypt cvv: %w", err))
			}
		} else {
			encryptedCvv = cardCvv
		}
	}
	return &UpdateStatusRequest{
		Reference:       fmt.Sprintf("updatestatus_%s_%d", customerId, clock.Now().UnixNano()),
		Product:         c.cardProduct,
		Channel:         c.cardChannel,
		Program:         c.cardProgram,
		TransactionType: "UPDATE_STATUS",
		CustomerId:      customerId,
		CardId:          cardId,
		AccountNumber:   accountNumber,
		StatusAction:    statusAction,
		CardCvv:         encryptedCvv,
		IsEncrypt:       true,
	}, nil
}

func (c *NetXDCardApiClient) UpdateStatus(req UpdateStatusRequest) (NetXDApiResponse[UpdateStatusResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[UpdateStatusResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[UpdateStatusResult]
	err := c.call("ledger.CARD.request", c.url, req, &response)
	return response, errtrace.Wrap(err)
}

func encryptCardDataForLedger(cardData string, publicKey string) (string, error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil || block.Type != "PUBLIC KEY" {
		return "", errtrace.Wrap(fmt.Errorf("failed to decode PEM block containing public key"))
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", errtrace.Wrap(fmt.Errorf("failed to parse public key: %v", err))
	}
	rsaPublicKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return "", errtrace.Wrap(fmt.Errorf("not an RSA public key"))
	}
	encryptedByt, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, []byte(cardData))
	if err != nil {
		return "", errtrace.Wrap(fmt.Errorf("failed to encrypt: %v", err))
	}
	encryptedvalue := base64.StdEncoding.EncodeToString(encryptedByt)
	return encryptedvalue, nil
}
