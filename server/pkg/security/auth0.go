package security

import (
	"context"
	"net/http"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/logging"
	"process-api/pkg/utils"
	"slices"
	"strings"
	"time"

	"braces.dev/errtrace"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

type Auth0Client struct {
	provider     *oidc.Provider
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

type Auth0Claims struct {
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Name          string   `json:"name"`
	Picture       string   `json:"picture"`
	Roles         []string `json:"https://dreamfi.com/roles"`
	jwt.RegisteredClaims
}

type AdminUserContext struct {
	echo.Context
	UserID string
	Email  string
	Name   string
	Roles  []string
}

var auth0Client *Auth0Client

func InitAuth0(cfg config.Auth0Configs) error {
	provider, err := oidc.NewProvider(context.Background(), "https://"+cfg.Domain+"/")
	if err != nil {
		return err
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientId,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectUrl,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "offline_access"},
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.ClientId,
	})

	auth0Client = &Auth0Client{
		provider:     provider,
		oauth2Config: oauth2Config,
		verifier:     verifier,
	}

	return nil
}

func GetAuthURL(state string) string {
	return auth0Client.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

type RawIDToken = string

func ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, RawIDToken, *Auth0Claims, error) {
	token, err := auth0Client.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, "", nil, errtrace.Wrap(err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, "", nil, echo.NewHTTPError(http.StatusInternalServerError, "No id_token field in oauth2 token")
	}

	idToken, err := auth0Client.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, "", nil, errtrace.Wrap(err)
	}

	var claims Auth0Claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, "", nil, errtrace.Wrap(err)
	}

	return token, rawIDToken, &claims, nil
}

func AdminAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		logger := logging.GetEchoContextLogger(c)

		// Check for Auth0 ID token cookie
		idTokenCookie, err := c.Cookie("admin_id_token")
		if err != nil {
			logger.Info("No admin ID token cookie found")
			return c.Redirect(http.StatusFound, "/admin/login")
		}

		idToken, err := auth0Client.verifier.Verify(c.Request().Context(), idTokenCookie.Value)
		if err != nil {
			logger.Info("Invalid Auth0 ID token", "error", err)
			return c.Redirect(http.StatusFound, "/admin/login")
		}

		var claims Auth0Claims
		if err := idToken.Claims(&claims); err != nil {
			logger.Info("Failed to parse Auth0 claims", "error", err)
			return c.Redirect(http.StatusFound, "/admin/login")
		}

		// check after half time has passed for 15 minute auth0 token
		if time.Until(claims.ExpiresAt.Time) < 7*time.Minute+30*time.Second {
			logger.Info("Token close to expiry, attempting refresh")

			sess, err := session.Get("admin_session", c)
			if err != nil {
				logger.Info("No session available", "error", err)
				return c.Redirect(http.StatusFound, "/admin/login")
			}

			encryptedRefreshToken, ok := sess.Values["refresh_token"].([]byte)
			if !ok {
				logger.Info("No refresh token in session")
				return c.Redirect(http.StatusFound, "/admin/login")
			}

			if expiry, ok := sess.Values["refresh_token_expiry"].(int64); ok {
				if clock.Now().Unix() > expiry {
					logger.Info("Refresh token expired")
					ClearAdminSession(c)
					return c.Redirect(http.StatusFound, "/admin/login")
				}
			}

			refreshToken, err := utils.DecryptKmsBinary(encryptedRefreshToken)
			if err != nil {
				logger.Info("Failed to decrypt refresh token", "error", err)
				ClearAdminSession(c)
				return c.Redirect(http.StatusFound, "/admin/login")
			}

			newToken, newRawIDToken, newClaims, err := RefreshAuth0Token(c.Request().Context(), refreshToken)
			if err != nil {
				logger.Info("Failed to refresh token", "error", err)
				ClearAdminSession(c)
				return c.Redirect(http.StatusFound, "/admin/login")
			}

			err = CreateAdminSession(c, newToken, newRawIDToken)
			if err != nil {
				logger.Error("Failed to update session after refresh", "error", err)
				return c.Redirect(http.StatusFound, "/admin/login")
			}

			claims = *newClaims
			logger.Debug("Token refreshed successfully")
		}

		env := utils.GetEnv()

		var requiredRole string
		if env == constant.PROD {
			requiredRole = "production-operations"
		} else {
			requiredRole = "sandbox-operations"
		}

		hasRequiredRole := slices.Contains(claims.Roles, requiredRole)

		if !hasRequiredRole {
			logger.Info("User does not have appropriate operations role for production or sandbox", "email", claims.Email, "roles", claims.Roles, "requiredRole", requiredRole, "environment", env)
			return c.HTML(http.StatusForbidden, "<h1>Access Denied</h1><p>You do not have permission to access this area.</p>")
		}

		adminCtx := &AdminUserContext{
			Context: c,
			UserID:  claims.Subject,
			Email:   claims.Email,
			Name:    claims.Name,
			Roles:   claims.Roles,
		}

		return next(adminCtx)
	}
}

func CreateAdminSession(c echo.Context, oauth2Token *oauth2.Token, rawIDToken RawIDToken) error {
	idTokenCookie := &http.Cookie{
		Name:     "admin_id_token",
		Value:    rawIDToken,
		Expires:  oauth2Token.Expiry,
		HttpOnly: true,
		Secure:   strings.HasPrefix(config.Config.Server.BaseUrl, "https://"),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}

	if oauth2Token.RefreshToken != "" {
		sess, err := session.Get("admin_session", c)
		if err != nil {
			return errtrace.Wrap(err)
		}

		encryptedRefreshToken, err := utils.EncryptKmsBinary(oauth2Token.RefreshToken)
		if err != nil {
			return errtrace.Wrap(err)
		}

		sess.Values["refresh_token"] = encryptedRefreshToken
		sess.Values["refresh_token_expiry"] = clock.Now().Add(time.Hour * 24 * 30).Unix() // 30 days

		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return errtrace.Wrap(err)
		}
	}

	c.SetCookie(idTokenCookie)
	return nil
}

func ClearAdminSession(c echo.Context) {
	idTokenCookie := &http.Cookie{
		Name:     "admin_id_token",
		Value:    "",
		Expires:  clock.Now().Add(-time.Hour),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   strings.HasPrefix(config.Config.Server.BaseUrl, "https://"),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
	c.SetCookie(idTokenCookie)

	sess, err := session.Get("admin_session", c)
	if err != nil {
		logging.Logger.Error("failed to get session on ClearAdminSession", "error", err)
	}

	sess.Values["refresh_token"] = nil
	sess.Values["refresh_token_expiry"] = nil
	if err = sess.Save(c.Request(), c.Response()); err != nil {
		logging.Logger.Error("failed to save session on ClearAdminSession", "error", err)
	}
}

func RefreshAuth0Token(ctx context.Context, refreshToken string) (*oauth2.Token, RawIDToken, *Auth0Claims, error) {
	tokenSource := auth0Client.oauth2Config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, "", nil, errtrace.Wrap(err)
	}

	rawIDToken, ok := newToken.Extra("id_token").(string)
	if !ok {
		return nil, "", nil, echo.NewHTTPError(http.StatusInternalServerError, "No id_token in refreshed token")
	}

	idToken, err := auth0Client.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, "", nil, errtrace.Wrap(err)
	}

	var claims Auth0Claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, "", nil, errtrace.Wrap(err)
	}

	return newToken, rawIDToken, &claims, nil
}
