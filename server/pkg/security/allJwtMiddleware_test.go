package security

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/logging"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type JwtExample struct {
	userID   string
	typeName string
	token    string
}

type JwtMiddlwareExample struct {
	jwt        JwtExample
	middleware func(echo.HandlerFunc) echo.HandlerFunc
	callback   func(c echo.Context, example JwtMiddlwareExample)
	shouldPass bool
}

func TestAllExamples(t *testing.T) {
	logging.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	config.Config.Jwt.TimeoutInMinutes = 30

	now := clock.Now()

	onboardedUserID := uuid.New().String()
	onboardedJwt, err := GenerateOnboardedJwt(onboardedUserID, "examplePublicKey", &now)
	if !assert.NoError(t, err) {
		return
	}
	onboardedExample := JwtExample{onboardedUserID, "onboardedJwt", onboardedJwt}
	onboardedPass := func(c echo.Context, example JwtMiddlwareExample) {
		cc, ok := c.(*LoggedInRegisteredUserContext)
		assert.True(t, ok)
		assert.Equal(t, example.jwt.userID, cc.UserId)
	}

	onboardingUserID := uuid.New().String()
	onboardingJwt, err := GenerateOnboardingJwt(onboardingUserID, &now)
	if !assert.NoError(t, err) {
		return
	}
	onboardingExample := JwtExample{onboardingUserID, "onboardingJwt", onboardingJwt}
	onboardingPass := func(c echo.Context, example JwtMiddlwareExample) {
		cc, ok := c.(*OnboardingUserContext)
		assert.True(t, ok)
		assert.Equal(t, example.jwt.userID, cc.UserId)
	}

	unregisteredUserID := uuid.New().String()
	unregisteredJwt, err := GenerateUnregisteredOnboardedJwt(unregisteredUserID, &now)
	if !assert.NoError(t, err) {
		return
	}
	unregisteredExample := JwtExample{unregisteredUserID, "unregisteredJwt", unregisteredJwt}
	unregisteredPass := func(c echo.Context, example JwtMiddlwareExample) {
		cc, ok := c.(*LoggedInUnregisteredUserContext)
		assert.True(t, ok)
		assert.Equal(t, example.jwt.userID, cc.UserId)
	}

	recoverUserID := uuid.New().String()
	recoverJwt, err := GenerateRecoverOnboardingJwt(recoverUserID, &now)
	if !assert.NoError(t, err) {
		return
	}
	recoverExample := JwtExample{recoverUserID, "recoverJwt", recoverJwt}
	recoverPass := func(c echo.Context, example JwtMiddlwareExample) {
		cc, ok := c.(*RecoverOnboardingUserContext)
		assert.True(t, ok)
		assert.Equal(t, example.jwt.userID, cc.UserId)
	}

	fail := func(c echo.Context, _ JwtMiddlwareExample) {
		assert.Fail(t, "middleware did not reject jwt")
	}

	examples := []JwtMiddlwareExample{
		{onboardedExample, LoggedInRegisteredUserMiddleware, onboardedPass, true},
		{onboardingExample, LoggedInRegisteredUserMiddleware, fail, false},
		{unregisteredExample, LoggedInRegisteredUserMiddleware, fail, false},
		{recoverExample, LoggedInRegisteredUserMiddleware, fail, false},

		{onboardedExample, LoggedInUnregisteredUserMiddleware, fail, false},
		{onboardingExample, LoggedInUnregisteredUserMiddleware, fail, false},
		{unregisteredExample, LoggedInUnregisteredUserMiddleware, unregisteredPass, true},
		{recoverExample, LoggedInUnregisteredUserMiddleware, fail, false},

		{onboardedExample, OnboardingUserMiddleware, fail, false},
		{onboardingExample, OnboardingUserMiddleware, onboardingPass, true},
		{unregisteredExample, OnboardingUserMiddleware, fail, false},
		{recoverExample, OnboardingUserMiddleware, fail, false},

		{onboardedExample, RecoverOnboardingUserMiddleware, fail, false},
		{onboardingExample, RecoverOnboardingUserMiddleware, fail, false},
		{unregisteredExample, RecoverOnboardingUserMiddleware, fail, false},
		{recoverExample, RecoverOnboardingUserMiddleware, recoverPass, true},

		{onboardedExample, RefreshSessionMiddleware, onboardedPass, true},
		{onboardingExample, RefreshSessionMiddleware, onboardingPass, true},
		{unregisteredExample, RefreshSessionMiddleware, unregisteredPass, true},
		{recoverExample, RefreshSessionMiddleware, recoverPass, true},
	}

	for _, example := range examples {
		fullMiddlewareName := runtime.FuncForPC(reflect.ValueOf(example.middleware).Pointer()).Name()
		parts := strings.Split(fullMiddlewareName, ".")
		middlewareName := parts[len(parts)-1]

		passOrFail := "fail"
		if example.shouldPass {
			passOrFail = "pass"
		}

		t.Run(fmt.Sprintf("%s should %s on %s", middlewareName, passOrFail, example.jwt.typeName), func(t *testing.T) {
			AssertJwtMiddleware(t, example)
		})
	}
}

func AssertJwtMiddleware(t *testing.T, example JwtMiddlwareExample) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	req.Header.Set("Authorization", "Bearer "+example.jwt.token)

	handlerCalled := false

	handler := func(c echo.Context) error {
		handlerCalled = true
		example.callback(c, example)
		return c.String(http.StatusOK, "Success")
	}

	middleware := example.middleware(handler)

	err := middleware(c)

	if example.shouldPass {
		if assert.NoError(t, err, "middleware should not return an error when expected to pass") {
			return
		}

		if assert.True(t, handlerCalled, "middleware should call the handler when expected to pass") {
			return
		}

		if assert.Equal(t, http.StatusOK, rec.Code, "middleware handler should return http 200 when expected to pass") {
			return
		}
		if assert.Contains(t, rec.Body.String(), "Success", "middleware handler should return http body 'Success' when expected to pass") {
			return
		}
	} else {
		if err != nil {
			if assert.ErrorContains(t, err, "401", "middleware should return error with code 401 when expected to fail") {
				return
			}
		} else {
			if assert.Equal(t, http.StatusUnauthorized, rec.Code, "middleware handler should return http 401 when expected to fail") {
				return
			}
		}

		if assert.False(t, handlerCalled, "middleware should not call the handler when expected to fail") {
			return
		}
	}
}
