package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSearchFieldLength(t *testing.T) {
	// Test Case 1
	t.Run("Valid 18 bytes", func(t *testing.T) {
		assert.True(t, ValidateSearchFieldLength("1600 Amphitheatre"), "Expected true")
	})

	t.Run("Valid exact 127 bytes", func(t *testing.T) {
		assert.True(t, ValidateSearchFieldLength("1234 Elm Street, Apt 56B, Springfield, IL 62704. Please deliver before 5 PM. Leave at the back door. Thanks for your service!"), "Expected true")
	})

	t.Run("Invalid 130 bytes", func(t *testing.T) {
		assert.False(t, ValidateSearchFieldLength("1234 Elm Street, Apartment 56B, Springfield, IL 62704. Please leave package at the back door before 5:00 PM. Thank you for your service"), "Expected false")
	})
}
