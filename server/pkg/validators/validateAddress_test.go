package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidAddress(t *testing.T) {
	t.Run("Valid address", func(t *testing.T) {
		assert.True(t, validateAddress("456 Elm Street, Apt #12"), "Expected true")
	})

	t.Run("Empty address", func(t *testing.T) {
		assert.True(t, validateAddress(""), "Expected true")
	})

	t.Run("Address which contains an exclamation mark", func(t *testing.T) {
		assert.False(t, validateAddress("456 Elm Street, Apt #12!"), "Expected false")
	})

	t.Run("Address with leading space", func(t *testing.T) {
		assert.False(t, validateAddress(" 1600 Pennsylvania Ave NW"), "Expected false")
	})

	t.Run("Address with trailing space", func(t *testing.T) {
		assert.False(t, validateAddress("1600 Pennsylvania Ave NW "), "Expected false")
	})

	t.Run("Address with leading and trailing space", func(t *testing.T) {
		assert.False(t, validateAddress(" 1600 Pennsylvania Ave NW "), "Expected false")
	})

	VISA_DISALLOWED_CHARS := []string{"<", ">", "*", "\\", "%", "(", ")", "=", ";", "{", "}", "?", "|", "[", "]", "~", "@", "&", "`"}
	for _, char := range VISA_DISALLOWED_CHARS {
		t.Run("Address containing disallowed character: "+char, func(t *testing.T) {
			assert.False(t, validateAddress("1600 Pennsylvania Ave "+char), "Expected false")
		})
	}
}
