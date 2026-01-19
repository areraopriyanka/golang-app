package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidateTransactionNote(t *testing.T) {
	validNotes := []string{
		"Too short",
		"V",
		"A 10 chars",
		"1234567890",
		"Hello",
		"",
	}
	for _, note := range validNotes {
		t.Run("Valid note "+note, func(t *testing.T) {
			assert.True(t, validateTransactionNote(note), "should be valid: "+note)
		})
	}

	t.Run("Note with leading space", func(t *testing.T) {
		assert.False(t, validateTransactionNote(" This starts with a space"), "Expected false")
	})
	t.Run("Note with greater than 40 chars", func(t *testing.T) {
		assert.False(t, validateTransactionNote(" Note with greater than 40 chars"), "Expected false")
	})
	t.Run("Note with trailing space", func(t *testing.T) {
		assert.False(t, validateTransactionNote("This ends with a space "), "Expected false")
	})

	t.Run("Note with leading and trailing space", func(t *testing.T) {
		assert.False(t, validateTransactionNote(" Leading and trailing space "), "Expected false")
	})
}
