package salesforce

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		wantToken   string
		wantErr     bool
		expectedErr string
	}{
		{
			name:       "valid bearer token",
			authHeader: "Bearer abc123xyz",
			wantToken:  "abc123xyz",
			wantErr:    false,
		},
		{
			name:       "valid bearer token with case insensitivity",
			authHeader: "bearer abc123xyz",
			wantToken:  "abc123xyz",
			wantErr:    false,
		},
		{
			name:        "empty authorization header",
			authHeader:  "",
			wantToken:   "",
			wantErr:     true,
			expectedErr: "authorization header is empty",
		},
		{
			name:        "missing Bearer prefix",
			authHeader:  "abc123xyz",
			wantToken:   "",
			wantErr:     true,
			expectedErr: "authorization header is malformed",
		},
		{
			name:        "missing token after Bearer",
			authHeader:  "Bearer",
			wantToken:   "",
			wantErr:     true,
			expectedErr: "authorization header is malformed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := extractBearerToken(tt.authHeader)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.EqualError(t, err, tt.expectedErr)
				}
				assert.Equal(t, "", token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantToken, token)
			}
		})
	}
}

func TestSalesforceClaimsValidate(t *testing.T) {
	claims := SalesforceClaims{
		Scope: "read:accounts write:accounts",
	}

	err := claims.Validate(context.TODO())
	assert.NoError(t, err, "Validate should always return nil")
}

func TestSalesforceClaimsHasScope(t *testing.T) {
	tests := []struct {
		name      string
		claims    SalesforceClaims
		scope     string
		wantMatch bool
	}{
		{
			name: "single scope matches",
			claims: SalesforceClaims{
				Scope: "read:accounts",
			},
			scope:     "read:accounts",
			wantMatch: true,
		},
		{
			name: "single scope does not match",
			claims: SalesforceClaims{
				Scope: "read:accounts",
			},
			scope:     "write:accounts",
			wantMatch: false,
		},
		{
			name: "multiple scopes with match",
			claims: SalesforceClaims{
				Scope: "read:accounts write:accounts delete:accounts",
			},
			scope:     "write:accounts",
			wantMatch: true,
		},
		{
			name: "multiple scopes without match",
			claims: SalesforceClaims{
				Scope: "read:accounts write:accounts",
			},
			scope:     "delete:accounts",
			wantMatch: false,
		},
		{
			name: "empty scope string",
			claims: SalesforceClaims{
				Scope: "",
			},
			scope:     "read:accounts",
			wantMatch: false,
		},
		{
			name: "checking for empty scope",
			claims: SalesforceClaims{
				Scope: "read:accounts write:accounts",
			},
			scope:     "",
			wantMatch: false,
		},
		{
			name: "partial match should not succeed",
			claims: SalesforceClaims{
				Scope: "read:accounts",
			},
			scope:     "read",
			wantMatch: false,
		},
		{
			name: "scope at the beginning",
			claims: SalesforceClaims{
				Scope: "read:accounts write:accounts delete:accounts",
			},
			scope:     "read:accounts",
			wantMatch: true,
		},
		{
			name: "scope at the end",
			claims: SalesforceClaims{
				Scope: "read:accounts write:accounts delete:accounts",
			},
			scope:     "delete:accounts",
			wantMatch: true,
		},
		{
			name: "scope in the middle",
			claims: SalesforceClaims{
				Scope: "read:accounts write:accounts delete:accounts",
			},
			scope:     "write:accounts",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.claims.HasScope(tt.scope)
			assert.Equal(t, tt.wantMatch, result)
		})
	}
}
