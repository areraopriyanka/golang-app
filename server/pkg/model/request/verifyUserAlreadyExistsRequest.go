package request

type VerifyUserAlreadyExistsRequest struct {
	Email     string `json:"email" mask:"true"`
	Mobile_No string `json:"mobileNo" mask:"true"`
}
