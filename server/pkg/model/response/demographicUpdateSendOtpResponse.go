package response

type DemographicUpdateSendOtpResponse struct {
	OtpExpiryDuration int    `json:"otpExpiryDuration" validate:"required"`
	OtpId             string `json:"otpId" validate:"required" mask:"true"`
	MaskedMobileNo    string `json:"maskedMobileNo" validate:"required" mask:"true"`
	MaskedEmail       string `json:"maskedEmail" validate:"required" mask:"true"`
}
