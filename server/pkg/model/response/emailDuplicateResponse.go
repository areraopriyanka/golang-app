package response

type EmailDuplicateResponse struct {
	IsEmailDuplicate bool `json:"isEmailInUse" validate:"required"`
}
