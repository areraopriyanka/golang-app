package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskEmail(t *testing.T) {
	t.Run("One char email username", func(t *testing.T) {
		email := "u@email.com"

		assert.Equal(t, MaskEmail(email), email, "One char email usernames are not masked")
	})

	t.Run("Two char email username", func(t *testing.T) {
		email := "us@email.com"

		assert.Equal(t, MaskEmail(email), "u*@email.com", "Two char email usernames have only one char masked")
	})

	t.Run("Email username over two chars", func(t *testing.T) {
		email := "username@email.com"

		assert.Equal(t, MaskEmail(email), "us******@email.com", "Longer email usernames have two unmasked chars at start")
		assert.Equal(t, len(MaskEmail(email)), len(email), "No chars are lost when email is masked")
	})
}
