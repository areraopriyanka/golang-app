package response

type OtpResponse struct {
	OtpExpiryDuration int    `json:"otpExpiryDuration" validate:"required"`
	OtpId             string `json:"otpId" validate:"required"`
}

type OtpResponseWithMaskedNumber struct {
	OtpExpiryDuration int    `json:"otpExpiryDuration" validate:"required"`
	OtpId             string `json:"otpId" validate:"required"`
	MaskedMobileNo    string `json:"maskedMobileNo" validate:"required"`
}
