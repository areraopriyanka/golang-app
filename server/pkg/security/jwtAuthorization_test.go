package security

import (
	"log/slog"
	"os"
	"process-api/pkg/clock"
	"process-api/pkg/logging"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExpiredJwt(t *testing.T) {
	logging.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	expiredTime := clock.Now().Add(time.Minute * -31)
	token, _ := GenerateOnboardedJwt("exampleUserId", "examplePublicKey", &expiredTime)
	claims := GetClaimsFromToken(token)

	assert.Nil(t, claims)
}
