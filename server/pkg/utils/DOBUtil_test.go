package utils

import (
	"process-api/pkg/clock"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsValidAge(t *testing.T) {
	// Test Case 1
	t.Run("Exact 18 years old", func(t *testing.T) {
		dob := clock.Now().AddDate(-18, 0, 0)
		// current time
		now := func() time.Time {
			return clock.Now().UTC()
		}
		assert.True(t, IsValidAge(dob, now), "Expected true for exactly 18 years old")
	})

	// Test Case 2
	t.Run("Older than 18 years old", func(t *testing.T) {
		dob := clock.Now().AddDate(-20, 0, 0)
		// current time
		now := func() time.Time {
			return clock.Now().UTC()
		}
		assert.True(t, IsValidAge(dob, now), "Expected true for older than 18 years old")
	})

	// Test Case 3
	t.Run("Younger than 18 years old", func(t *testing.T) {
		dob := clock.Now().AddDate(-17, 0, 0)
		// current time
		now := func() time.Time {
			return clock.Now().UTC()
		}
		assert.False(t, IsValidAge(dob, now), "Expected false for younger than 18 years old")
	})

	// Test Case 4
	t.Run("Just under 18 years old", func(t *testing.T) {
		dob := clock.Now().AddDate(-18, 0, 1)
		// current time
		now := func() time.Time {
			return clock.Now().UTC()
		}
		assert.False(t, IsValidAge(dob, now), "Expected false for just under 18 years old")
	})

	// Test Case 5
	t.Run("Exact 18 years old", func(t *testing.T) {
		dob := time.Date(2005, 12, 25, 0, 0, 0, 0, time.UTC)
		// Simulating current date to 25th December 2023
		now := func() time.Time {
			return time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
		}
		assert.True(t, IsValidAge(dob, now), "Expected true for exactly 18 years old")
	})

	// Test Case 6
	t.Run("Leap year birthday on leap day", func(t *testing.T) {
		dob := time.Date(2004, 2, 29, 0, 0, 0, 0, time.UTC)
		// current time
		now := func() time.Time {
			return clock.Now().UTC()
		}
		assert.True(t, IsValidAge(dob, now), "Expected true for someone born on February 29, 2004")
	})
}
