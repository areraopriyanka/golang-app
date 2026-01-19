package plaid

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"braces.dev/errtrace"
	"github.com/golang-jwt/jwt/v5"
	"github.com/plaid/plaid-go/v34/plaid"
)

type webhookKeyMap struct {
	mu sync.RWMutex
	m  map[string]*ecdsa.PublicKey
}

func newWebhookKeyMap() *webhookKeyMap {
	return &webhookKeyMap{m: make(map[string]*ecdsa.PublicKey)}
}

func (s *webhookKeyMap) Get(keyID string) (*ecdsa.PublicKey, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[keyID]
	return v, ok
}

func (s *webhookKeyMap) GetAllKeyIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.m))
	for id := range s.m {
		ids = append(ids, id)
	}
	return ids
}

func (s *webhookKeyMap) Set(keyID string, publicKey *ecdsa.PublicKey) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[keyID] = publicKey
}

func (s *webhookKeyMap) Delete(keyID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, keyID)
}

var webhookKeyCache = newWebhookKeyMap()

func (ps *PlaidService) VerifyWebhook(webhookBody string, untrustedTokenString string) bool {
	logger := ps.Logger.WithGroup("VerifyWebhook")

	if untrustedTokenString == "" {
		logger.Error("Missing plaid-verification header")
		return false
	}

	// Parse untrusted token string, but don't validate untrustedToken yet
	untrustedToken, _, err := new(jwt.Parser).ParseUnverified(untrustedTokenString, jwt.MapClaims{})
	if err != nil {
		logger.Error("Error parsing unverified token string", "error", err.Error())
		return false
	}
	if untrustedToken == nil {
		logger.Error("Error parsing unverified token string")
		return false
	}

	// Plaid docs:
	// > Ensure that the value of the alg (algorithm) field in the header is "ES256".
	// > Reject the webhook if this is not the case.
	if untrustedToken.Method.Alg() != "ES256" {
		logger.Error("Wrong algorithm used; expected ES256", "got", untrustedToken.Method.Alg())
		return false
	}

	// Extract the key ID (kid) from the token header
	// WARN: the token is not verified here; do not trust its contents
	// or do anything with its data except get the public key from Plaid
	kid, ok := untrustedToken.Header["kid"].(string)
	if !ok {
		logger.Error("Could not cast kid to string")
		return false
	}

	logger.Debug("getting key from cache", "kid", kid, "kids", webhookKeyCache.GetAllKeyIDs())
	publicKey, ok := webhookKeyCache.Get(kid)
	if !ok {
		logger.Debug("couldn't get key from cache", "kid", kid, "kids", webhookKeyCache.GetAllKeyIDs())
		// A new kid might indicate public key rollover; in which case, re-fetch
		// any existing webhook keys to see if any have expired.
		ps.checkForExpiredWebhookKeys()
		// Then resume with attempting to fetch the new webhook public key:
		publicKey, err = ps.fetchWebhookKey(kid)
		if err != nil {
			logger.Error("Webhook key not set; failing verification", "error", err.Error())
			return false
		}
		webhookKeyCache.Set(kid, publicKey)
		logger.Debug("added key to cache", "kid", kid, "kids", webhookKeyCache.GetAllKeyIDs())
	}

	parsedToken, err := jwt.Parse(untrustedTokenString, func(token *jwt.Token) (any, error) {
		return publicKey, nil
	},
		// Need to re-verify signing method since we got `alg` earlier with the untrusted token
		// https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
		jwt.WithValidMethods([]string{jwt.SigningMethodES256.Alg()}),
		jwt.WithLeeway(5*time.Minute),
		jwt.WithIssuedAt())
	if err != nil {
		logger.Error("Error parsing jwt", "error", err.Error())
		return false
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		logger.Error("Error parsing claims", "error", "not ok")
		return false
	}
	if !parsedToken.Valid {
		logger.Error("Error parsing claims", "error", "not valid")
		return false
	}

	// [Docs](https://plaid.com/docs/api/webhooks/webhook-verification/)
	// > Use the issued at time denoted by the iat field to verify that the webhook is
	// > not more than 5 minutes old. Rejecting outdated webhooks can help prevent replay attacks.
	// Note: golang-jwt makes us manually test this [release notes](https://github.com/golang-jwt/jwt/discussions/308#discussion-5097549)
	iat, ok := claims["iat"].(float64)
	if !ok {
		logger.Error("Could not cast iat as float64", "iat", claims["iat"])
		return false
	}
	issuedAt := time.Unix(int64(iat), 0)
	if time.Since(issuedAt) > 5*time.Minute {
		logger.Error("Rejecting webhook event; too old.")
		return false
	}

	requestBodySHA256, ok := claims["request_body_sha256"].(string)
	if !ok {
		logger.Error("Could not cast claims' request_body_sha256 to string")
		return false
	}

	hasher := crypto.SHA256.New()
	hasher.Write([]byte(webhookBody))
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	if requestBodySHA256 != expectedHash {
		logger.Error("Rejecting webhook, as shas do not match")
		return false
	}
	return true
}

func (ps *PlaidService) checkForExpiredWebhookKeys() {
	logger := ps.Logger.WithGroup("checkForExpiredWebhookKeys")
	for _, kid := range webhookKeyCache.GetAllKeyIDs() {
		publicKey, err := ps.fetchWebhookKey(kid)
		if err != nil {
			logger.Error("Error with existing plaid webhook", "error", err.Error())
			webhookKeyCache.Delete(kid)
		} else if publicKey != nil {
			webhookKeyCache.Set(kid, publicKey)
		}
	}
}

func (ps *PlaidService) fetchWebhookKey(kid string) (*ecdsa.PublicKey, error) {
	ctx := context.Background()
	webhookRequest := *plaid.NewWebhookVerificationKeyGetRequest(kid)
	webhookResponse, _, respErr := ps.Plaid.PlaidApi.WebhookVerificationKeyGet(ctx).WebhookVerificationKeyGetRequest(webhookRequest).Execute()

	if respErr != nil {
		return nil, errtrace.Wrap(fmt.Errorf("error calling WebhookVerificationKeyGet %w", respErr))
	}
	jwk := webhookResponse.GetKey()

	if jwk.ExpiredAt.IsSet() {
		return nil, errtrace.Wrap(errtrace.Wrap(errors.New("jwk public key is expired")))
	}

	if jwk.Kty != "EC" {
		return nil, errtrace.Wrap(fmt.Errorf("unexpected kty: %s", jwk.Kty))
	}
	if jwk.Crv != "P-256" {
		return nil, errtrace.Wrap(fmt.Errorf("unexpected crv: %s", jwk.Crv))
	}

	// This was inspired by Plaid's docs [example implementation](https://plaid.com/docs/api/webhooks/webhook-verification/#example-implementation)
	// However, their impl was truncated at `y, _ := base64.URLEncoding.DecodeString(cachedKey.Y + `
	// examples implementations exist on GitHub e.g., [greed](https://github.com/jms-guy/greed/blob/d0e3d6a5382c49609dc69d76c75d27cb4a6dd72b/backend/internal/auth/jwt_token.go#L119)
	publicKey := new(ecdsa.PublicKey)
	publicKey.Curve = elliptic.P256()
	x, err := base64.URLEncoding.DecodeString(jwk.X + "=")
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("error decoding jwt public key X value %w", err))
	}
	xc := new(big.Int)
	publicKey.X = xc.SetBytes(x)
	y, err := base64.URLEncoding.DecodeString(jwk.Y + "=")
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("error decoding jwt public key Y value %w", err))
	}
	yc := new(big.Int)
	publicKey.Y = yc.SetBytes(y)
	ps.Logger.Debug("Successfully added webhook public key")
	return publicKey, nil
}
