package security

import (
	"errors"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/logging"
	"strings"
	"time"

	"braces.dev/errtrace"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func GenerateUnregisteredOnboardedJwt(userId string, now *time.Time) (string, error) {
	if now == nil {
		temp := clock.Now()
		now = &temp
	}

	claims := JwtClaims{
		Type:      "unregistered-onboarded",
		UserState: constant.ACTIVE,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(config.Config.Jwt.TimeoutInMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(*now),
			NotBefore: jwt.NewNumericDate(*now),
			Subject:   userId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(config.Config.Jwt.SecreteKey))
}

func GenerateOnboardedJwt(userId string, publicKey string, now *time.Time) (string, error) {
	if now == nil {
		temp := clock.Now()
		now = &temp
	}

	claims := JwtClaims{
		Type:      "onboarded",
		PublicKey: publicKey,
		UserState: constant.ACTIVE,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(config.Config.Jwt.TimeoutInMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(*now),
			NotBefore: jwt.NewNumericDate(*now),
			Subject:   userId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(config.Config.Jwt.SecreteKey))
}

func GenerateRecoverOnboardingJwt(userId string, now *time.Time) (string, error) {
	if now == nil {
		temp := clock.Now()
		now = &temp
	}

	claims := JwtClaims{
		Type:      "recover-onboarding",
		UserState: constant.ONBOARDING,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(config.Config.Jwt.TimeoutInMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(*now),
			NotBefore: jwt.NewNumericDate(*now),
			Subject:   userId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(config.Config.Jwt.SecreteKey))
}

func GenerateOnboardingJwt(userId string, now *time.Time) (string, error) {
	if now == nil {
		temp := clock.Now()
		now = &temp
	}

	claims := JwtClaims{
		Type:      "onboarding",
		UserState: constant.ONBOARDING,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(config.Config.Jwt.TimeoutInMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(*now),
			NotBefore: jwt.NewNumericDate(*now),
			Subject:   userId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(config.Config.Jwt.SecreteKey))
}

func GetJwtUserId(c echo.Context) (string, error) {
	tokenWithBearer := c.Request().Header.Get("Authorization")
	claims := GetClaimsFromToken(tokenWithBearer)
	if claims == nil {
		return "", errtrace.Wrap(errors.New("failed to retrieve JWT claims"))
	}
	if claims.Subject == "" {
		return "", errtrace.Wrap(errors.New("user ID is missing in JWT claims"))
	}
	return claims.Subject, nil
}

func GetJwtUserState(c echo.Context) (string, error) {
	tokenWithBearer := c.Request().Header.Get("Authorization")
	claims := GetClaimsFromToken(tokenWithBearer)
	if claims == nil {
		return "", errtrace.Wrap(errors.New("failed to retrieve JWT claims"))
	}
	return claims.UserState, nil
}

func GetJwtType(c echo.Context) (string, error) {
	tokenWithBearer := c.Request().Header.Get("Authorization")
	claims := GetClaimsFromToken(tokenWithBearer)
	if claims == nil {
		return "", errtrace.Wrap(errors.New("failed to retrieve JWT claims"))
	}
	return claims.Type, nil
}

func GetJwtPublicKey(c echo.Context) (string, error) {
	tokenWithBearer := c.Request().Header.Get("Authorization")
	claims := GetClaimsFromToken(tokenWithBearer)
	if claims == nil {
		return "", errtrace.Wrap(errors.New("failed to retrieve JWT claims"))
	}
	if claims.PublicKey == "" {
		return "", errtrace.Wrap(errors.New("PublicKey is missing in JWT claims"))
	}
	return claims.PublicKey, nil
}

func GetClaimsFromToken(tokenWithBearer string) *JwtClaims {
	if tokenWithBearer == "" {
		logging.Logger.Error(constant.TOKEN_EMPTY_ERROR_MSG)
		return nil
	}

	// token without Bearer
	tokenString := strings.ReplaceAll(tokenWithBearer, "Bearer", "")
	tokenString = strings.ReplaceAll(tokenString, " ", "")

	claims := &JwtClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Config.Jwt.SecreteKey), nil
	}, jwt.WithExpirationRequired())
	if err != nil {
		logging.Logger.Error("error in GetClaimsFromToken", "error", err.Error())
		return nil
	}

	if claims.ExpiresAt.Before(clock.Now()) {
		logging.Logger.Error("Token is expired")
		return nil
	}

	return claims
}

func ResetToken(userId string, publicKey string, now *time.Time) (string, error) {
	token, err := GenerateOnboardedJwt(userId, publicKey, now)
	if err != nil {
		logging.Logger.Error("Error generating JWT", "error", err)
		return "", errtrace.Wrap(errors.New("could not generate token"))
	}

	return "Bearer " + token, nil
}

func ResetUnregisteredToken(userId string, now *time.Time) (string, error) {
	token, err := GenerateUnregisteredOnboardedJwt(userId, now)
	if err != nil {
		logging.Logger.Error("Error generating unregistered JWT", "error", err)
		return "", errtrace.Wrap(errors.New("could not generate token"))
	}

	return "Bearer " + token, nil
}

func ResetOnboardingToken(userId string, now *time.Time) (string, error) {
	token, err := GenerateOnboardingJwt(userId, now)
	if err != nil {
		logging.Logger.Error("Error generating onboarding JWT", "error", err)
		return "", errtrace.Wrap(errors.New("could not generate token"))
	}

	return "Bearer " + token, nil
}

func ResetRecoverOnboardingToken(userId string, now *time.Time) (string, error) {
	token, err := GenerateRecoverOnboardingJwt(userId, now)
	if err != nil {
		logging.Logger.Error("Error generating recover onboarding JWT", "error", err)
		return "", errtrace.Wrap(errors.New("could not generate token"))
	}

	return "Bearer " + token, nil
}
