package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidateZipCode(t *testing.T) {
	t.Run("Valid zip code", func(t *testing.T) {
		assert.True(t, validateZipCode("09867"), "Expected true")
	})

	t.Run("9 digit zip code", func(t *testing.T) {
		assert.False(t, validateZipCode("098678978"), "Expected false")
	})

	t.Run("9 digit zip code with hyphen", func(t *testing.T) {
		assert.False(t, validateZipCode("09867-8978"), "Expected false")
	})
}
