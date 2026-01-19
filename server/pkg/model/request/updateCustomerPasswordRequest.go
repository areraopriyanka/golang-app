package request

type UpdateCustomerPasswordRequest struct {
	Password string `json:"password" validate:"required,maxByte=72" mask:"true"`
}
