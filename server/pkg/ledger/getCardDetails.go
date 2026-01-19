package ledger

import (
	"fmt"
	"process-api/pkg/clock"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type GetCardDetailsRequest struct {
	Reference       string `json:"reference"`
	TransactionType string `json:"transactionType" validate:"required"`
	CustomerId      string `json:"customerId" validate:"required"`
	AccountNumber   string `json:"accountNumber" validate:"required"`
	Product         string `json:"product" validate:"required"`
	Channel         string `json:"channel" validate:"required"`
	Program         string `json:"program" validate:"required"`
	CardId          string `json:"cardId" validate:"required"`
}

type GetCardDetailsResult struct {
	Card struct {
		CardId         string `json:"cardId"`
		CardProduct    string `json:"cardProduct"`
		CreatedDate    string `json:"createdDate"`
		UpdatedDate    string `json:"updatedDate"`
		CardMaskNumber string `json:"cardMaskNumber"`
		CardStatus     string `json:"cardStatus"`
		CardExpiryDate string `json:"cardExpiryDate"`
		OrderStatus    string `json:"orderStatus"`
		OrderId        string `json:"orderId"`
		IsReIssue      bool   `json:"isReIssue"`
		IsReplace      bool   `json:"isReplace"`
		OrderSubStatus string `json:"orderSubStatus"`
		ExternalCardId string `json:"externalCardId"`
	} `json:"cardDetails"`
	Api struct {
		Type              string `json:"type"`
		Reference         string `json:"reference"`
		OriginalReference string `json:"originalReference"`
		DateCreated       int64  `json:"dateCreated"`
	} `json:"api"`
}

func (c *NetXDCardApiConfig) BuildGetCardDetailsRequest(customerId string, accountNumber string, cardId string) GetCardDetailsRequest {
	return GetCardDetailsRequest{
		Reference:       fmt.Sprintf("getcarddetails_%s_%d", customerId, clock.Now().UnixNano()),
		TransactionType: "GET_CARD_DETAILS",
		CustomerId:      customerId,
		AccountNumber:   accountNumber,
		Product:         c.cardProduct,
		Channel:         c.cardChannel,
		Program:         c.cardProgram,
		CardId:          cardId,
	}
}

func (c *NetXDCardApiClient) GetCardDetails(req GetCardDetailsRequest) (NetXDApiResponse[GetCardDetailsResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[GetCardDetailsResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[GetCardDetailsResult]
	err := c.call("ledger.CARD.request", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
