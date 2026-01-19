package security

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/logging"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestResetTokenMiddlewareForActiveUser(t *testing.T) {
	logging.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	config.Config.Jwt.TimeoutInMinutes = 30

	handler := func(c echo.Context) error {
		return c.NoContent(200)
	}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/")

	userId := "c6b841a-dfc1-4a6a-a12a-13c21035b5be"
	publicKey := "public-key-text"

	customContext := GenerateLoggedInRegisteredUserContext(userId, publicKey, c)

	err := ResetTokenMiddleware(handler)(customContext)
	if !assert.NoError(t, err, "Handler should not return an error") {
		return
	}

	authorization := rec.Result().Header.Get("Authorization")
	if !assert.NotEmpty(t, authorization, "Authorization must be present") {
		return
	}

	claims := GetClaimsFromToken(authorization)
	if !assert.NotNil(t, claims, "claims must be successfully parsed") {
		return
	}

	assert.Equal(t, userId, claims.Subject)
	assert.Equal(t, publicKey, claims.PublicKey)
	assert.Equal(t, constant.ACTIVE, claims.UserState)
}

func TestResetTokenMiddlewareForOnboardingUser(t *testing.T) {
	logging.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	config.Config.Jwt.TimeoutInMinutes = 30

	handler := func(c echo.Context) error {
		return c.NoContent(200)
	}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/")

	userId := "c6b841a-dfc1-4a6a-a12a-13c21035b5be"

	customContext := GenerateOnboardingUserContext(userId, c)

	err := ResetTokenMiddleware(handler)(customContext)
	if !assert.NoError(t, err, "ResetTokenMiddleware should not return an error") {
		return
	}

	authorization := rec.Result().Header.Get("Authorization")
	if !assert.NotEmpty(t, authorization, "Authorization must be present") {
		return
	}

	claims := GetClaimsFromToken(authorization)
	if !assert.NotNil(t, claims, "claims must be successfully parsed") {
		return
	}

	assert.Equal(t, userId, claims.Subject)
	assert.Equal(t, constant.ONBOARDING, claims.UserState)
}

func TestResetTokenMiddlewareOnInternalError(t *testing.T) {
	handler := func(c echo.Context) error {
		return c.NoContent(500)
	}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/")

	userId := "c6b841a-dfc1-4a6a-a12a-13c21035b5be"
	publicKey := "public-key-text"

	customContext := GenerateLoggedInRegisteredUserContext(userId, publicKey, c)

	err := ResetTokenMiddleware(handler)(customContext)
	if !assert.NoError(t, err, "Handler should not return an error") {
		return
	}

	authorization := rec.Result().Header.Get("Authorization")
	assert.Empty(t, authorization, "Authorization must be empty")
}
