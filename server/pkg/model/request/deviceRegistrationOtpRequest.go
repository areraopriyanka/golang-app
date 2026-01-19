package request

type GetDeviceRegistrationOtpRequest struct {
	Type string `json:"type" validate:"required,otpType"`
}

type ChallengeDeviceRegistrationOtpRequest struct {
	OtpId     string `json:"otpId" validate:"required"`
	OtpValue  string `json:"otpValue" validate:"required"`
	PublicKey string `json:"publicKey" validate:"required"`
}
