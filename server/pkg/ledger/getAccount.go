// Generated from https://apidocs.netxd.com/developers/docs/account_apis/Get%20Account/
package ledger

import (
	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

type GetAccountRequest struct {
	ID string `json:"ID" validate:"required"`
}

type GetAccountResult struct {
	Account struct {
		Id             string   `json:"id"`
		Name           string   `json:"name"`
		Number         string   `json:"number"`
		NickName       string   `json:"nickName"`
		CreatedDate    string   `json:"createdDate"`
		UpdatedDate    string   `json:"updatedDate"`
		Balance        int64    `json:"balance"`
		Debit          bool     `json:"debit"`
		MinimumBalance int64    `json:"minimumBalance"`
		HoldBalance    int64    `json:"holdBalance"`
		SubLedgerCode  string   `json:"subLedgerCode"`
		Tags           []string `json:"tags"`
		Final          bool     `json:"final"`
		Parent         struct {
			ID     string `json:"ID"`
			Code   string `json:"code"`
			Name   string `json:"name"`
			Number string `json:"number"`
		} `json:"parent"`
		CustomerID      string `json:"customerID"`
		CustomerName    string `json:"customerName"`
		InstitutionName string `json:"institutionName"`
		AccountCategory string `json:"accountCategory"`
		AccountType     string `json:"accountType"`
		Currency        string `json:"currency"`
		CurrencyCode    string `json:"currencyCode"`
		LegalReps       []struct {
			ID          string `json:"ID"`
			Name        string `json:"name"`
			CreatedDate string `json:"createdDate"`
			UpdatedDate string `json:"updatedDate"`
		} `json:"legalReps"`
		Status        string `json:"status"`
		InstitutionID string `json:"institutionID"`
		GlAccount     string `json:"glAccount"`
		DDAAccount    bool   `json:"DDAAccount"`
		Address       struct {
			AddressLine1 string `json:"addressLine1"`
			City         string `json:"city"`
			State        string `json:"state"`
			Country      string `json:"country"`
			Zip          string `json:"zip"`
		} `json:"address"`
		IsVerify              bool   `json:"isVerify"`
		MinimumRouteApprovers int64  `json:"minimumRouteApprovers"`
		NewRouteAlert         bool   `json:"newRouteAlert"`
		CeTransactionNumber   string `json:"ceTransactionNumber"`
		AccountLevel          string `json:"accountLevel"`
		LedgerBalance         int64  `json:"ledgerBalance"`
		PreAuthBalance        int64  `json:"preAuthBalance"`
		AccountFinderSync     bool   `json:"accountFinderSync"`
		IsGLVerify            bool   `json:"isGLVerify"`
		IsShadowAccount       bool   `json:"isShadowAccount"`
		Sweep                 bool   `json:"sweep"`
		IsClosed              bool   `json:"isClosed"`
		Program               string `json:"program"`
		ProductID             string `json:"productID"`
		ExternalLedger        bool   `json:"externalLedger"`
	} `json:"account"`
}

func (c *NetXDLedgerApiClient) GetAccount(accountID string) (NetXDApiResponse[GetAccountResult], error) {
	req := &GetAccountRequest{ID: accountID}
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[GetAccountResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[GetAccountResult]
	err := c.call("AccountService.GetAccount", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
