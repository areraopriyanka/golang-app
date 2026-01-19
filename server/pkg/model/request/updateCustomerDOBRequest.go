package request

type UpdateCustomerDOBRequest struct {
	DOB string `json:"DOB" validate:"required,validateDOB" mask:"true"`
}
