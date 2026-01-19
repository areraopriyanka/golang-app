// Generated from https://apidocs.netxd.com/developers/docs/card_apis/Card%20Management/Replace%20Reissue%20Card
package ledger

import (
	"fmt"
	"process-api/pkg/clock"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

const (
	REPLACE_CARD = "REPLACE_CARD"
	REISSUE      = "REISSUE"
)

type ReplaceOrReissueCardRequest struct {
	Reference       string `json:"reference"`
	Product         string `json:"product" validate:"required"`
	Program         string `json:"program" validate:"required"`
	Channel         string `json:"channel" validate:"required"`
	TransactionType string `json:"transactionType" validate:"required"`
	CustomerId      string `json:"customerId" validate:"required"`
	CardId          string `json:"cardId" validate:"required"`
	AccountNumber   string `json:"accountNumber" validate:"required"`
	StatusAction    string `json:"statusAction" validate:"required"`
}

type ReplaceOrReissueCard struct {
	CardId                 string   `json:"cardId"`
	CardHolderId           string   `json:"cardHolderId"`
	CardHolderName         string   `json:"cardHolderName"`
	CardProduct            string   `json:"cardProduct"`
	CustomerId             string   `json:"customerId"`
	AccountId              string   `json:"accountId"`
	Product                string   `json:"product"`
	Program                string   `json:"program"`
	CardType               string   `json:"cardType"`
	PostedDate             string   `json:"postedDate"`
	UpdatedDate            string   `json:"updatedDate"`
	CardMaskNumber         string   `json:"cardMaskNumber"`
	CardNumber             string   `json:"cardNumber"`
	CardStatus             string   `json:"cardStatus"`
	CardExpiryDate         string   `json:"cardExpiryDate"`
	AllowAtm               bool     `json:"allowAtm"`
	AllowEcommerce         bool     `json:"allowEcommerce"`
	AllowMoto              bool     `json:"allowMoto"`
	AllowPos               bool     `json:"allowPos"`
	AllowTips              bool     `json:"allowTips"`
	AllowPurchase          bool     `json:"allowPurchase"`
	AllowRefund            bool     `json:"allowRefund"`
	AllowCashback          bool     `json:"allowCashback"`
	AllowWithdraw          bool     `json:"allowWithdraw"`
	AllowAuthAndCompletion bool     `json:"allowAuthAndCompletion"`
	Smart                  bool     `json:"smart"`
	CheckAvsZip            bool     `json:"checkAvsZip"`
	CheckAvsAddr           bool     `json:"checkAvsAddr"`
	Cvv                    string   `json:"cvv"`
	AccountNumber          string   `json:"accountNumber"`
	CardName               string   `json:"cardName"`
	Patterns               []string `json:"patterns"`
	TransactionMade        bool     `json:"transactionMade"`
	OrderStatus            string   `json:"orderStatus"`
	OrderId                string   `json:"orderId"`
	Network                string   `json:"network"`
	IsReIssue              bool     `json:"isReIssue"`
	IsReplace              bool     `json:"isReplace"`
	ExternalCardId         string   `json:"externalCardId"`
	CardCreatedYear        string   `json:"cardCreatedYear"`
	OrderSubStatus         string   `json:"orderSubStatus"`
	AccountName            string   `json:"accountName"`
	CustomerName           string   `json:"customerName"`
}

type ReplaceOrReissueCardResult struct {
	Card    ReplaceOrReissueCard `json:"card" validate:"required"`
	NewCard ReplaceOrReissueCard `json:"newCard" validate:"required"`
	Api     struct {
		Type              string `json:"type"`
		Reference         string `json:"reference"`
		DateCreated       int64  `json:"dateCreated"`
		OriginalReference string `json:"originalReference"`
	} `json:"api"`
}

func (c *NetXDCardApiConfig) BuildReplaceOrReissueCardRequest(customerId string, cardId string, accountNumber string, statusAction string) ReplaceOrReissueCardRequest {
	return ReplaceOrReissueCardRequest{
		Reference:       fmt.Sprintf("replaceOrReissueCard_%s_%d", customerId, clock.Now().UnixNano()),
		Product:         c.cardProduct,
		Channel:         c.cardChannel,
		Program:         c.cardProgram,
		TransactionType: REPLACE_CARD,
		CustomerId:      customerId,
		CardId:          cardId,
		AccountNumber:   accountNumber,
		// StatusAction can be either REPLACE_CARD or REISSUE
		StatusAction: statusAction,
	}
}

func (c *NetXDCardApiClient) ReplaceOrReissueCard(req ReplaceOrReissueCardRequest) (NetXDApiResponse[ReplaceOrReissueCardResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[ReplaceOrReissueCardResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[ReplaceOrReissueCardResult]
	err := c.call("ledger.CARD.request", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
