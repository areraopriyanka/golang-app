package response

type CreateOnboardingUserResponse struct {
	Id string `json:"userId" validate:"required"`
}
