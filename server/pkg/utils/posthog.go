package utils

import (
	"fmt"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/logging"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/posthog/posthog-go"
)

type Posthog struct {
	client posthog.Client
}

var PosthogClient *Posthog

func InitializePosthogClient(posthogConfig config.PosthogConfigs) error {
	client, err := posthog.NewWithConfig(posthogConfig.ProjectKey, posthog.Config{
		Endpoint: posthogConfig.BaseUrl,
	})
	if err != nil {
		return fmt.Errorf("failed to create PostHog client: %w", err)
	}
	PosthogClient = &Posthog{client: client}
	return nil
}

func (ph *Posthog) SendErrorToPosthog(c echo.Context, err error, trace string) {
	exception := map[string]interface{}{
		"type":  "ServerError",
		"value": err.Error(),
	}
	props := posthog.NewProperties().
		Set("$exception_message", err.Error()).
		Set("$exception_list", []map[string]interface{}{exception}).
		Set("$current_url", c.Request().URL.Path).
		Set("exception_stacktrace", trace)

	distinctId := uuid.New().String()
	if userId, ok := c.Get("user_id").(string); ok && userId != "" {
		distinctId = userId
	}

	err = ph.client.Enqueue(posthog.Capture{
		// DistinctId is a mandatory field. PostHog will ignore an event if it is not specified.
		// We are passing the userId to this field. If userId is not present, we generate a UUID.
		DistinctId: distinctId,
		Event:      "$exception",
		Properties: props,
		Timestamp:  clock.Now(),
	})
	if err != nil {
		logging.Logger.Error("Could not enqueue", "error", err)
	}
}

func ClosePosthogClient() error {
	if PosthogClient != nil && PosthogClient.client != nil {
		return PosthogClient.client.Close()
	}
	return nil
}
