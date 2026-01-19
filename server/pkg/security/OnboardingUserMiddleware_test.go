package security

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/logging"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestOnboardingUserMiddleware(t *testing.T) {
	logging.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	config.Config.Jwt.TimeoutInMinutes = 30
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userId := uuid.New().String()
	now := clock.Now()

	token, _ := GenerateOnboardingJwt(userId, &now)

	req.Header.Set("Authorization", "Bearer "+token)

	handlerCalled := false

	handler := func(c echo.Context) error {
		handlerCalled = true
		cc, ok := c.(*OnboardingUserContext)
		assert.True(t, ok)
		assert.Equal(t, userId, cc.UserId)
		return c.String(http.StatusOK, "Success")
	}

	middleware := OnboardingUserMiddleware(handler)

	err := middleware(c)
	assert.NoError(t, err)

	// ensuring handler was called
	assert.True(t, handlerCalled)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Success")
}

func TestOnboardingUserMiddlewareWitActiveJWT(t *testing.T) {
	logging.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	config.Config.Jwt.TimeoutInMinutes = 30
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userId := uuid.New().String()
	now := clock.Now()
	// Generate onboarding jwt
	token, _ := GenerateOnboardedJwt(userId, "examplePublicKey", &now)
	req.Header.Set("Authorization", "Bearer "+token)

	handlerCalled := false
	handler := func(c echo.Context) error {
		// Should NOT be called in this test
		handlerCalled = true
		return c.String(http.StatusOK, "Success")
	}

	middleware := OnboardingUserMiddleware(handler)

	err := middleware(c)
	assert.ErrorContains(t, err, "401")

	// Should NOT call handler because token is invalid
	assert.False(t, handlerCalled)
}
