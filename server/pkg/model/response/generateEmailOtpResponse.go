package response

type GenerateEmailOtpResponse struct {
	OtpId             string `json:"otpId" validate:"required"`
	OtpExpiryDuration int    `json:"otpExpiryDuration" validate:"required"`
}
