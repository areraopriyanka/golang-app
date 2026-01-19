package security

import (
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/model/response"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

type BaseShareContext struct {
	echo.Context
}

type LoggedInRegisteredUserContext struct {
	BaseShareContext
	UserId    string
	PublicKey string
}

func LoggedInRegisteredUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		jwtType, err := GetJwtType(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, err)
		}
		if jwtType != "onboarded" {
			return response.ErrorResponse{ErrorCode: constant.INVALID_TOKEN, Message: constant.INVALID_TOKEN_MSG, StatusCode: http.StatusUnauthorized, MaybeInnerError: errtrace.New("")}
		}

		userState, err := GetJwtUserState(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, err)
		}
		// If the token does not belong to an onboarded user
		if userState != constant.ACTIVE {
			return response.ErrorResponse{ErrorCode: constant.INVALID_TOKEN, Message: constant.INVALID_TOKEN_MSG, StatusCode: http.StatusUnauthorized, MaybeInnerError: errtrace.New("")}
		}

		userId, err := GetJwtUserId(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, err)
		}

		if userId == "" {
			return response.ErrorResponse{ErrorCode: constant.USER_ID_MISSING, Message: constant.USER_ID_MISSING_MSG, StatusCode: http.StatusUnauthorized, MaybeInnerError: errtrace.New("")}
		}

		publicKey, err := GetJwtPublicKey(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, err)
		}
		if publicKey == "" {
			return response.ErrorResponse{ErrorCode: constant.PUBLIC_KEY_MISSING, Message: constant.PUBLIC_KEY_MISSING_MSG, StatusCode: http.StatusUnauthorized, MaybeInnerError: errtrace.New("")}
		}

		cc := GenerateLoggedInRegisteredUserContext(userId, publicKey, c)
		c.Set("user_id", userId)
		return next(cc)
	}
}

func GenerateLoggedInRegisteredUserContext(userId string, publicKey string, c echo.Context) *LoggedInRegisteredUserContext {
	return &LoggedInRegisteredUserContext{
		BaseShareContext: BaseShareContext{c},
		UserId:           userId,
		PublicKey:        publicKey,
	}
}
