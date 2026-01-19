package salesforce

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/model/response"
	"slices"
	"strings"
	"time"

	"braces.dev/errtrace"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/labstack/echo/v4"
)

// @title DreamFi Salesforce Integration API
// @version 0.0.1

type SalesforceClaims struct {
	Scope string `json:"scope"`
}

// Validate does nothing, but we need it to satisfy validator.CustomClaims interface.
func (c SalesforceClaims) Validate(ctx context.Context) error {
	return nil
}

func SalesforceAuth0Middleware() echo.MiddlewareFunc {
	issuerURL, err := url.Parse("https://" + config.Config.Salesforce.InboundAuth0Domain + "/")
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{config.Config.Salesforce.InboundAuth0Audience},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &SalesforceClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to set up the jwt validator: %v", err)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := extractBearerToken(c.Request().Header.Get("Authorization"))
			if err != nil {
				return response.ErrorResponse{
					ErrorCode:       constant.UNAUTHORIZED_ACCESS_ERROR,
					Message:         "Missing or invalid authorization header",
					StatusCode:      http.StatusUnauthorized,
					LogMessage:      err.Error(),
					MaybeInnerError: errtrace.Wrap(err),
				}
			}

			validatedClaims, err := jwtValidator.ValidateToken(c.Request().Context(), token)
			if err != nil {
				return response.ErrorResponse{
					ErrorCode:       constant.UNAUTHORIZED_ACCESS_ERROR,
					Message:         "Failed to validate JWT",
					StatusCode:      http.StatusUnauthorized,
					LogMessage:      err.Error(),
					MaybeInnerError: errtrace.Wrap(err),
				}
			}

			c.Set("salesforce_claims", validatedClaims)

			return next(c)
		}
	}
}

func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errtrace.Wrap(errors.New("authorization header is empty"))
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errtrace.Wrap(errors.New("authorization header is malformed"))
	}

	return parts[1], nil
}

func (c SalesforceClaims) HasScope(scope string) bool {
	return slices.Contains(strings.Split(c.Scope, " "), scope)
}
