package request

type SetCardPinRequest struct {
	Signature string `json:"signature" validate:"required" mask:"true"`
	PayloadId string `json:"payloadId" validate:"required"`
	OtpId     string `json:"otpId" validate:"required"`
	Otp       string `json:"otp" validate:"required"`
}
