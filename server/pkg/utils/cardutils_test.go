package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidPin(t *testing.T) {
	// Test Case 1
	t.Run("Is Numeric", func(t *testing.T) {
		assert.True(t, IsValidUnencryptedPIN("1234"), "Expected true to be numeric")
	})

	// Test Case 2
	t.Run("Is Numeric and 4 digit", func(t *testing.T) {
		assert.True(t, IsValidUnencryptedPIN("1234"), "Expected true to be numeric and four digits")
	})

	// Test Case 3
	t.Run("Is not numeric nor 4 digit", func(t *testing.T) {
		assert.False(t, IsValidUnencryptedPIN("r8gjfgh76fughkjh"), "Expected false to be numeric and four digits")
	})

	t.Run("Is not numeric nor 4 digit", func(t *testing.T) {
		assert.False(t, IsValidUnencryptedPIN("123"), "Expected false to be numeric and four digits")
	})

	t.Run("Valid 4-digit numeric PIN (leading zeros allowed) ", func(t *testing.T) {
		assert.True(t, IsValidUnencryptedPIN("0000"), "Expected true to be numeric and four digits")
	})
}

func TestIsValidCVV(t *testing.T) {
	// Test Case 1
	t.Run("Is Numeric and 3 digit", func(t *testing.T) {
		assert.True(t, IsValidUnencrypted3DigitCVV("124"), "Expected true to be numeric and three digits")
	})

	// Test Case 2
	t.Run("Is not numeric nor 4 digit", func(t *testing.T) {
		assert.False(t, IsValidUnencrypted3DigitCVV("dfete"), "Expected false to be numeric and three digits")
	})
	// Test Case 3
	t.Run("Valid 3-digit numeric CVV (leading zeros allowed) ", func(t *testing.T) {
		assert.True(t, IsValidUnencrypted3DigitCVV("000"), "Expected true to be numeric and three digits")
	})
}
