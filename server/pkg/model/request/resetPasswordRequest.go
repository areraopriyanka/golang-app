package request

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" validate:"required" mask:"true"`
	NewPassword string `json:"newPassword" validate:"required" mask:"true"`
	ResetToken  string `json:"resetToken" validate:"required" mask:"true"`
}

type ResetPasswordRequest struct {
	ResetToken string `json:"resetToken" validate:"required" mask:"true"`
	Password   string `json:"password" validate:"required" mask:"true"`
}
