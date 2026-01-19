package security

import (
	"process-api/pkg/clock"

	"github.com/labstack/echo/v4"
)

func ResetTokenMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Before(func() {
			statusCode := c.Response().Status

			if statusCode < 200 || statusCode >= 300 {
				c.Logger().Infof("Skipping token reset for response status: %d", statusCode)
				return
			}

			skipTokenReset, _ := c.Get("SkipTokenReset").(bool)
			if skipTokenReset {
				c.Logger().Infof("Skipping token reset")
				return
			}

			now := clock.Now()

			switch cc := c.(type) {
			case *LoggedInRegisteredUserContext:
				newToken, err := ResetToken(cc.UserId, cc.PublicKey, &now)
				if err != nil {
					c.Logger().Errorf("Error resetting registered user token: %s", err)
					return
				}
				c.Response().Header().Set("Authorization", newToken)

			case *LoggedInUnregisteredUserContext:
				newToken, err := ResetUnregisteredToken(cc.UserId, &now)
				if err != nil {
					c.Logger().Errorf("Error resetting unregistered user token: %s", err)
					return
				}
				c.Response().Header().Set("Authorization", newToken)

			case *OnboardingUserContext:
				newToken, err := ResetOnboardingToken(cc.UserId, &now)
				if err != nil {
					c.Logger().Errorf("Error resetting onboarding token: %s", err)
					return
				}
				c.Response().Header().Set("Authorization", newToken)

			case *RecoverOnboardingUserContext:
				newToken, err := ResetRecoverOnboardingToken(cc.UserId, &now)
				if err != nil {
					c.Logger().Errorf("Error resetting onboarding recovery token: %s", err)
					return
				}
				c.Response().Header().Set("Authorization", newToken)

			default:
				c.Logger().Errorf("Unknown context type for token reset: %T", cc)
				return
			}
		})

		return next(c)
	}
}
