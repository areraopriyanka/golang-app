package request

type ChallengeOtpRequest struct {
	OtpId string `json:"otpId" validate:"required"`
	Otp   string `json:"otp" validate:"required"`
}
