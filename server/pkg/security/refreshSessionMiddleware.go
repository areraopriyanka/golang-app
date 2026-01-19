package security

import (
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/model/response"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

func RefreshSessionMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		jwtType, err := GetJwtType(c)
		if err != nil {
			return response.ErrorResponse{
				ErrorCode:       constant.UNAUTHORIZED_ACCESS_ERROR,
				Message:         constant.UNAUTHORIZED_ACCESS_ERROR_MSG,
				StatusCode:      http.StatusUnauthorized,
				LogMessage:      "Failed to get jwtType",
				MaybeInnerError: errtrace.Wrap(err),
			}
		}

		switch jwtType {
		case "onboarded":
			return LoggedInRegisteredUserMiddleware(next)(c)
		case "unregistered-onboarded":
			return LoggedInUnregisteredUserMiddleware(next)(c)
		case "onboarding":
			return OnboardingUserMiddleware(next)(c)
		case "recover-onboarding":
			return RecoverOnboardingUserMiddleware(next)(c)
		default:
			return response.ErrorResponse{
				ErrorCode:       constant.UNAUTHORIZED_ACCESS_ERROR,
				Message:         constant.UNAUTHORIZED_ACCESS_ERROR_MSG,
				StatusCode:      http.StatusUnauthorized,
				LogMessage:      "Received unknown jwtType",
				MaybeInnerError: errtrace.Wrap(err),
			}
		}
	}
}
