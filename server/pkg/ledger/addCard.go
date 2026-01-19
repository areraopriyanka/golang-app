// Generated from https://apidocs.netxd.com/developers/docs/card_apis/Card%20Management/Add%20Physical%20Card
package ledger

import (
	"fmt"
	"process-api/pkg/clock"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type AddCardRequest struct {
	Reference       string             `json:"reference"`
	TransactionType string             `json:"transactionType" validate:"required"`
	CustomerId      string             `json:"customerId" validate:"required"`
	AccountNumber   string             `json:"accountNumber" validate:"required"`
	Product         string             `json:"product" validate:"required"`
	Channel         string             `json:"channel" validate:"required"`
	Program         string             `json:"program" validate:"required"`
	Card            AddCardRequestCard `json:"card"`
}

type AddCardRequestCard struct {
	CardHolderId string `json:"cardHolderId" validate:"required"`
	CardType     string `json:"cardType" validate:"required"`
}

type AddCardResult struct {
	Card struct {
		CardId                 string `json:"cardId"`
		CardType               string `json:"cardType"`
		PostedDate             string `json:"postedDate"`
		UpdatedDate            string `json:"updatedDate"`
		CardMaskNumber         string `json:"cardMaskNumber"`
		CardStatus             string `json:"cardStatus"`
		CardExpiryDate         string `json:"cardExpiryDate"`
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
		OrderStatus            string `json:"orderStatus"`
		OrderId                string `json:"orderId"`
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

func (c *NetXDCardApiConfig) BuildAddCardRequest(customerId string, accountNumber string, cardHolderId string) AddCardRequest {
	return AddCardRequest{
		Reference:       fmt.Sprintf("addcard_%s_%d", customerId, clock.Now().UnixNano()),
		TransactionType: "ADD_CARD",
		CustomerId:      customerId,
		AccountNumber:   accountNumber,
		Product:         c.cardProduct,
		Channel:         c.cardChannel,
		Program:         c.cardProgram,
		Card: AddCardRequestCard{
			CardHolderId: cardHolderId,
			CardType:     "PHYSICAL",
		},
	}
}

func (c *NetXDCardApiClient) AddCard(req AddCardRequest) (NetXDApiResponse[AddCardResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[AddCardResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[AddCardResult]
	err := c.call("ledger.CARD.request", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
