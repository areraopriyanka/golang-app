package request

type SendMobileVerificationOtpRequest struct {
	MobileNo string `json:"mobileNo" validate:"required"`
	Type     string `json:"type" validate:"required,oneof=SMS CALL"`
}

type VerifyMobileVerificationOtpRequest struct {
	OtpId string `json:"otpId" validate:"required"`
	Otp   string `json:"otp" validate:"required"`
}
