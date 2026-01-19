package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPOBox(t *testing.T) {
	// Test Case 1
	t.Run("Valid PO Box with number", func(t *testing.T) {
		assert.True(t, IsPOBox("PO Box 123"), "Expected true")
	})

	t.Run("PO Box with period and spaces", func(t *testing.T) {
		assert.True(t, IsPOBox("po.  box"), "Expected true")
	})

	t.Run("PO Box with mixed case", func(t *testing.T) {
		assert.True(t, IsPOBox("Po BoX"), "Expected true")
	})

	t.Run("PO Box with abbreviation", func(t *testing.T) {
		assert.True(t, IsPOBox("P.O.Box"), "Expected true")
	})

	t.Run("Address without PO Box", func(t *testing.T) {
		assert.False(t, IsPOBox("123 Main St"), "Expected false")
	})
}
