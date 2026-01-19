package request

type LoginRequest struct {
	Username          string `json:"username" validate:"required"`
	Password          string `json:"password" validate:"required" mask:"true"`
	PublicKey         string `json:"publicKey" validate:"max=255" mask:"true"`
	SardineSessionKey string `json:"sardineSessionKey" mask:"true"`
}
