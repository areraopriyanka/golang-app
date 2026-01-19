package security

import (
	"github.com/golang-jwt/jwt/v5"
)

type JwtClaims struct {
	Type      string `json:"type,omitempty"`
	UserState string `json:"userState,omitempty"`
	PublicKey string `json:"publicKey,omitempty"`
	jwt.RegisteredClaims
}
