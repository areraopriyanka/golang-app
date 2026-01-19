package response

type VerifyResetPasswordOtpResponse struct {
	ResetToken string `json:"resetToken" validate:"required" mask:"true"`
}
