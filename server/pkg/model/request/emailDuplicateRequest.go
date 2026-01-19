package request

type EmailDuplicateRequest struct {
	Email string `json:"email" validate:"required" mask:"true"`
}
