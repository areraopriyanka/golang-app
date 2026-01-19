package handler

import (
	"log/slog"
	"os"
	"process-api/pkg/logging"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapLedgerCardStatus(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logging.Logger = logger

	// Test Case 1
	t.Run("inactive returned for expected raw status", func(t *testing.T) {
		assert.Equal(t, MapLedgerCardStatus("CARD_IS_NOT_ACTIVATED", logger), "inactive")
		assert.Equal(t, MapLedgerCardStatus("RETURNED_UNDELIVERED", logger), "inactive")
	})

	// Test Case 2
	t.Run("active returned for expected raw status", func(t *testing.T) {
		assert.Equal(t, MapLedgerCardStatus("ACTIVE", logger), "active")
		assert.Equal(t, MapLedgerCardStatus("ACTIVATED", logger), "active")
	})

	// Test Case 3
	t.Run("frozen returned for expected raw status", func(t *testing.T) {
		assert.Equal(t, MapLedgerCardStatus("TEMPRORY_BLOCKED_BY_CLIENT", logger), "frozen")
		assert.Equal(t, MapLedgerCardStatus("TEMPRORY_BLOCKED_BY_ADMIN", logger), "frozen")
	})

	t.Run("cancelled returned for expected raw status", func(t *testing.T) {
		assert.Equal(t, MapLedgerCardStatus("CARD_REQUEST_NOT_PROCESSED", logger), "cancelled")
		assert.Equal(t, MapLedgerCardStatus("EXPIRED_CARD", logger), "cancelled")
		assert.Equal(t, MapLedgerCardStatus("LOST_STOLEN", logger), "cancelled")
		assert.Equal(t, MapLedgerCardStatus("CLOSED", logger), "cancelled")
		assert.Equal(t, MapLedgerCardStatus("DEACTIVATED", logger), "cancelled")
	})
}
