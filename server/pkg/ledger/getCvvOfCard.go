// Generated from https://apidocs.netxd.com/developers/docs/card_apis/Card%20Management/Get%20CVV
package ledger

import (
	"fmt"
	"process-api/pkg/clock"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type GetCvvOfCardRequest struct {
	Reference       string `json:"reference"`
	Product         string `json:"product" validate:"required"`
	Program         string `json:"program" validate:"required"`
	Channel         string `json:"channel" validate:"required"`
	TransactionType string `json:"transactionType" validate:"required"`
	CustomerId      string `json:"customerId" validate:"required"`
	AccountNumber   string `json:"accountNumber" validate:"required"`
	CardId          string `json:"cardId" validate:"required"`
}

type GetCvvOfCardResult struct {
	Card struct {
		CardId                 string `json:"cardId"`
		PostedDate             string `json:"postedDate"`
		UpdatedDate            string `json:"updatedDate"`
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

func (c *NetXDCardApiConfig) BuildGetCvvOfCardRequest(customerId string, accountNumber string, cardId string) GetCvvOfCardRequest {
	return GetCvvOfCardRequest{
		Reference:       fmt.Sprintf("getcvvofcard_%s_%d", customerId, clock.Now().UnixNano()),
		Product:         c.cardProduct,
		Channel:         c.cardChannel,
		Program:         c.cardProgram,
		TransactionType: "GET_CVV",
		CustomerId:      customerId,
		AccountNumber:   accountNumber,
		CardId:          cardId,
	}
}

func (c *NetXDCardApiClient) GetCvvOfCard(req GetCvvOfCardRequest) (NetXDApiResponse[GetCvvOfCardResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[GetCvvOfCardResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[GetCvvOfCardResult]
	err := c.call("ledger.CARD.request", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
