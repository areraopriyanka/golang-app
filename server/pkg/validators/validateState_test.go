package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidateState(t *testing.T) {
	t.Run("Valid state", func(t *testing.T) {
		assert.True(t, validateState("CA"), "Expected true")
	})

	t.Run("Empty state code", func(t *testing.T) {
		assert.False(t, validateState(""), "Expected false")
	})

	t.Run("State with leading space", func(t *testing.T) {
		assert.False(t, validateState(" CA"), "Expected false")
	})

	t.Run("State with trailing space", func(t *testing.T) {
		assert.False(t, validateState("CA "), "Expected false")
	})

	t.Run("Invalid state code", func(t *testing.T) {
		assert.False(t, validateState("XY"), "Expected false")
	})
}
