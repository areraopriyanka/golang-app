package response

type GeneratePayloadResponse struct {
	PayloadId string `json:"payloadId" validate:"required"`
	Payload   string `json:"payload" validate:"required" mask:"true"`
}
