package request

type UserAlreadyRegisteredRequest struct {
	Contact Contact `json:"contact"`
}
type Contact struct {
	Email       string `json:"email" mask:"true"`
	PhoneNumber string `json:"phoneNumber" mask:"true"`
}
