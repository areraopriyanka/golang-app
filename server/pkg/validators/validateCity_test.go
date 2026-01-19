package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidCity(t *testing.T) {
	t.Run("Valid city name", func(t *testing.T) {
		assert.True(t, validateCity("California"), "Expected true")
	})

	t.Run("Empty city name", func(t *testing.T) {
		assert.False(t, validateCity(""), "Expected false")
	})

	t.Run("City name with space in between", func(t *testing.T) {
		assert.True(t, validateCity("San Francisco"), "Expected true")
	})

	t.Run("Leading space", func(t *testing.T) {
		assert.False(t, validateCity(" New York"), "Expected false")
	})

	t.Run("Trailing space", func(t *testing.T) {
		assert.False(t, validateCity("California "), "Expected false")
	})

	t.Run("Leading and trailing spaces", func(t *testing.T) {
		assert.False(t, validateCity(" California "), "Expected false")
	})

	t.Run("City name with digits", func(t *testing.T) {
		assert.False(t, validateCity("California12"), "Expected false")
	})

	t.Run("City name with underscore", func(t *testing.T) {
		assert.False(t, validateCity("New_York"), "Expected false")
	})

	VISA_DISALLOWED_CHARS := []string{"<", ">", "*", "\\", "/", "%", "(", ")", "=", ";", "{", "}", "?", "|", "@", "[", "]", "~", "&", "'"}
	for _, char := range VISA_DISALLOWED_CHARS {
		t.Run("Address containing disallowed character: "+char, func(t *testing.T) {
			assert.False(t, validateCity("Clevland "+char), "Expected false")
		})
	}
}
