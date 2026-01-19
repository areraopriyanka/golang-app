package response

type ValidateCvvResponse struct {
	IsValidCvv bool `json:"isValidCvv" validate:"required"`
}
