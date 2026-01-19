package security

import (
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/model/response"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

type OnboardingUserContext struct {
	BaseShareContext
	UserId string
}

func OnboardingUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		jwtType, err := GetJwtType(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, err)
		}
		if jwtType != "onboarding" {
			return response.ErrorResponse{ErrorCode: constant.INVALID_TOKEN, Message: constant.INVALID_TOKEN_MSG, StatusCode: http.StatusUnauthorized, MaybeInnerError: errtrace.New("")}
		}

		userState, err := GetJwtUserState(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, err)
		}

		if userState == "" {
			return response.ErrorResponse{ErrorCode: constant.USER_STATE_MISSING, Message: constant.USER_STATE_MISSING_MSG, StatusCode: http.StatusUnauthorized, MaybeInnerError: errtrace.New("")}
		}

		// If the token does not belong to an onboarding user
		if userState != constant.ONBOARDING {
			return response.ErrorResponse{ErrorCode: constant.INVALID_TOKEN, Message: "Token does not belong to an onboarding user.", StatusCode: http.StatusUnauthorized, MaybeInnerError: errtrace.New("")}
		}

		userId, err := GetJwtUserId(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, err)
		}
		if userId == "" {
			return response.ErrorResponse{ErrorCode: constant.USER_ID_MISSING, Message: constant.USER_ID_MISSING_MSG, StatusCode: http.StatusUnauthorized, MaybeInnerError: errtrace.New("")}
		}

		cc := GenerateOnboardingUserContext(userId, c)
		c.Set("user_id", userId)
		return next(cc)
	}
}

func GenerateOnboardingUserContext(userId string, c echo.Context) *OnboardingUserContext {
	return &OnboardingUserContext{
		BaseShareContext: BaseShareContext{c},
		UserId:           userId,
	}
}
