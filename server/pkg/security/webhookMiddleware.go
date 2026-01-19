package security

import (
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/logging"

	"github.com/labstack/echo/v4"
	twiliovalidator "github.com/twilio/twilio-go/client"
)

// Removed webhookMiddleware and other webhook related code for now since it's not in use.
// Moved it to another branch: `ledger-webhook-code-backup`.
// If we start working on the ledger webhook again, we can pull it from there.

func TwilioWebhookMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		twilioSignature := c.Request().Header.Get("X-Twilio-Signature")

		if twilioSignature == "" {
			logging.Logger.Error("Missing twilio signature in the request header")
			return c.NoContent(http.StatusBadRequest)
		}
		url := config.Config.Twilio.CallbackUrl

		formParams := map[string]string{}
		if err := c.Request().ParseForm(); err != nil {
			logging.Logger.Error("Error while parsing the form data", "Error", err.Error())
			return c.NoContent(http.StatusBadRequest)
		}
		for key, values := range c.Request().Form {
			formParams[key] = values[0]
		}

		validator := twiliovalidator.NewRequestValidator(config.Config.Twilio.AuthToken)

		// Twilio creates the signature by hashing the request URL and form data using twilio Auth Token,
		if !validator.Validate(url, formParams, twilioSignature) {
			logging.Logger.Error("Invalid signature from twilio")
			return c.NoContent(http.StatusBadRequest)
		}
		return next(c)
	}
}
