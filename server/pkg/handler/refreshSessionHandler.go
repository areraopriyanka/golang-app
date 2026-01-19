package handler

import (
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"github.com/labstack/echo/v4"
)

// @summary RefreshSession
// @description Issues new JWT token for user with expiring session
// @tags auth
// @Success 200 "OK"
// @failure 401 {object} response.ErrorResponse
// @router /refresh-session [get]
func RefreshSession(c echo.Context) error {
	logger := logging.GetEchoContextLogger(c)
	switch ctx := c.(type) {
	case *security.LoggedInRegisteredUserContext:
		userCtx := ctx
		logger.Debug("Session refreshed for registered user", "userId", userCtx.UserId)
		return c.NoContent(http.StatusOK)

	case *security.OnboardingUserContext:
		onboardingCtx := ctx
		logger.Debug("Session refreshed for onboarding user", "userId", onboardingCtx.UserId)
		return c.NoContent(http.StatusOK)

	case *security.LoggedInUnregisteredUserContext:
		unregisteredCtx := ctx
		logger.Debug("Session refreshed for unregistered user", "userId", unregisteredCtx.UserId)
		return c.NoContent(http.StatusOK)

	case *security.RecoverOnboardingUserContext:
		recoverCtx := ctx
		logger.Debug("Session refreshed for recover onboarding user", "userId", recoverCtx.UserId)
		return c.NoContent(http.StatusOK)

	default:
		return response.ErrorResponse{
			ErrorCode:  constant.UNAUTHORIZED_ACCESS_ERROR,
			Message:    constant.UNAUTHORIZED_ACCESS_ERROR_MSG,
			StatusCode: http.StatusUnauthorized,
			LogMessage: "Unknown user context",
		}
	}
}
