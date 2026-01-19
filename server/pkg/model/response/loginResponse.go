package response

type LoginResponse struct {
	Id               string `json:"userId" validate:"required"`
	CustomerNo       string `json:"customerNo" mask:"true"`
	Status           string `json:"status"`
	IsUserRegistered bool   `json:"isUserRegistered"`
	IsUserOnboarding bool   `json:"isUserOnboarding"`
	MobileNo         string `json:"mobileNo" mask:"true"`
	Email            string `json:"email" mask:"true"`
}
