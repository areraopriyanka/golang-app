package response

type BuildPayloadResponse struct {
	PayloadId string `json:"payloadId" validate:"required"`
	Payload   string `json:"payload" validate:"required" mask:"true"`
}
