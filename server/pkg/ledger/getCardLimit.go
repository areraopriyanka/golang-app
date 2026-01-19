// Generated from https://apidocs.netxd.com/developers/docs/card_apis/Card%20Controls/Get%20Card%20Limit
package ledger

import (
	"fmt"
	"process-api/pkg/clock"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type GetCardLimitRequest struct {
	Reference       string `json:"reference"`
	TransactionType string `json:"transactionType" validate:"required"`
	CustomerId      string `json:"customerId" validate:"required"`
	AccountNumber   string `json:"accountNumber" validate:"required"`
	Product         string `json:"product" validate:"required"`
	Channel         string `json:"channel" validate:"required"`
	Program         string `json:"program" validate:"required"`
	CardId          string `json:"cardId" validate:"required"`
}

type GetCardLimitResult struct {
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
		Limits                 []struct {
			Type        string `json:"type"`
			Value       string `json:"value"`
			CycleLength string `json:"cycleLength"`
			CycleType   string `json:"cycleType"`
			Remaining   string `json:"remaining"`
		} `json:"limits"`
		TransactionMade bool `json:"transactionMade"`
		IsReIssue       bool `json:"isReIssue"`
		IsReplace       bool `json:"isReplace"`
	} `json:"card"`
	Api struct {
		Type              string `json:"type"`
		Reference         string `json:"reference"`
		DateCreated       int64  `json:"dateCreated"`
		OriginalReference string `json:"originalReference"`
	} `json:"api"`
}

func (c *NetXDCardApiConfig) BuildGetCardLimitRequest(customerId string, accountNumber string, cardId string) GetCardLimitRequest {
	return GetCardLimitRequest{
		Reference:       fmt.Sprintf("getcardlimit_%s_%d", customerId, clock.Now().UnixNano()),
		TransactionType: "GET_CARD_LIMIT",
		CustomerId:      customerId,
		AccountNumber:   accountNumber,
		Product:         c.cardProduct,
		Channel:         c.cardChannel,
		Program:         c.cardProgram,
		CardId:          cardId,
	}
}

func (c *NetXDCardApiClient) GetCardLimit(req GetCardLimitRequest) (NetXDApiResponse[GetCardLimitResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[GetCardLimitResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[GetCardLimitResult]
	err := c.call("ledger.CARD.request", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
