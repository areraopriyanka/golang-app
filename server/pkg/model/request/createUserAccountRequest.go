package request

type CreateUserAccountRequest struct {
	Email    string `json:"email" validate:"required,validateEmail" mask:"true"`
	Password string `json:"password" validate:"required,maxByte=72" mask:"true"`
}
