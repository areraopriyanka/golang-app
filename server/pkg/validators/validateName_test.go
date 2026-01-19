package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidateName(t *testing.T) {
	t.Run("Valid name", func(t *testing.T) {
		assert.True(t, validateName("Alice Mary"), "Expected true")
	})

	t.Run("Empty name", func(t *testing.T) {
		assert.False(t, validateName(""), "Expected false")
	})

	t.Run("Name which contains an exclamation mark", func(t *testing.T) {
		assert.False(t, validateName("John!"), "Expected false")
	})

	t.Run("Name with leading space", func(t *testing.T) {
		assert.False(t, validateName(" Alice Mary"), "Expected false")
	})

	t.Run("Name with trailing space", func(t *testing.T) {
		assert.False(t, validateName("Alice Mary "), "Expected false")
	})

	VISA_DISALLOWED_CHARS := []string{"<", ">", "*", "\\", "/", "%", "(", ")", "=", ";", "{", "}", "?", "|", "@", "[", "]", "~", "&", "'", "&", "`"}
	for _, char := range VISA_DISALLOWED_CHARS {
		t.Run("Address containing disallowed character: "+char, func(t *testing.T) {
			assert.False(t, validateName("Bob "+char), "Expected false")
		})
	}
}
