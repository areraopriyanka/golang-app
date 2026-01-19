package ledger

import "braces.dev/errtrace"

type AddUserKeyRequest struct {
	UserName  string `json:"userName"`
	PublicKey string `json:"publicKey"`
	Status    string `json:"status"`
}

type AddUserKeyResponse struct {
	KeyID  string `json:"keyID"`  // validate:"required" Docs say this isn't required... but that's the whole point of this endpoint
	Status string `json:"status"` // validate:"required,eq=ACTIVE"
	ApiKey string `json:"apiKey"`
}

func BuildAddUserKeyRequest(email string, publicKey string) AddUserKeyRequest {
	return AddUserKeyRequest{
		UserName:  email,
		PublicKey: publicKey,
		Status:    "ACTIVE", // ledger docs indicate this should be hardcoded
	}
}

func (c *NetXDLedgerApiClient) AddUserKey(req AddUserKeyRequest) (NetXDApiResponse[AddUserKeyResponse], error) {
	var response NetXDApiResponse[AddUserKeyResponse]
	err := c.call("CustomerService.AddUserKey", c.url, req, &response)
	return response, errtrace.Wrap(err)
}
