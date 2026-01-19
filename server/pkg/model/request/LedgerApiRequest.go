package request

type LedgerApiRequest struct {
	Signature string `json:"signature" validate:"required"`
	Mfp       string `json:"mfp"`
	PayloadId string `json:"payloadId" validate:"required"`
}
