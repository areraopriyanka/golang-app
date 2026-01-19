package request

type GenerateEmailVerificationOtpRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type VerifyEmailOtpRequest struct {
	Email string `json:"email" mask:"true"`
	Otp   string `json:"otp"`
}
