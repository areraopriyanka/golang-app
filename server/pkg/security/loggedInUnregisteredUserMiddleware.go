package security

import (
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/model/response"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

type LoggedInUnregisteredUserContext struct {
	BaseShareContext
	UserId string
}

func LoggedInUnregisteredUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		jwtType, err := GetJwtType(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, err)
		}
		if jwtType != "unregistered-onboarded" {
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
		if err == nil && publicKey != "" {
			return response.ErrorResponse{ErrorCode: "PUBLIC_KEY_PRESENT", Message: "User already has a registered device", StatusCode: http.StatusUnauthorized, MaybeInnerError: errtrace.New("")}
		}

		cc := GenerateLoggedInUnregisteredUserContext(userId, c)
		c.Set("user_id", userId)
		return next(cc)
	}
}

func GenerateLoggedInUnregisteredUserContext(userId string, c echo.Context) *LoggedInUnregisteredUserContext {
	return &LoggedInUnregisteredUserContext{
		BaseShareContext: BaseShareContext{c},
		UserId:           userId,
	}
}
