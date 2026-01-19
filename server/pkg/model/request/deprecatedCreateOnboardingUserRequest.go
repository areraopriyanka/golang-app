package request

type CreateOrUpdateOnboardingUserRequest struct {
	FirstName string `json:"firstName" validate:"required,validateName"`
	LastName  string `json:"lastName" validate:"required,validateName"`
	Suffix    string `json:"suffix"`
	Email     string `json:"email" validate:"required,validateEmail" mask:"true"`
}
