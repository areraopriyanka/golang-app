package utils

import (
	"regexp"
)

// IsValidUnencryptedPIN checks if the given string is numeric and exactly 4 digits long.
func IsValidUnencryptedPIN(pin string) bool {
	// Define the regex once at package level for efficiency
	pinRegex := regexp.MustCompile(`^\d{4}$`)
	if pin == "" {
		return false //, "PIN cannot be empty."
	}

	if !pinRegex.MatchString(pin) {
		return false //, "PIN must be exactly 4 numeric digits."
	}

	return true //, "PIN is valid."
}

// IsValidUnencrypted3DigitCVV checks if the given string is numeric and exactly 3 digits long.
func IsValidUnencrypted3DigitCVV(cvv string) bool {
	// Define the regex for a 3-digit CVV once at package level for efficiency.
	cvv3DigitRegex := regexp.MustCompile(`^\d{3}$`)

	// 1. Check for empty string.
	if cvv == "" {
		return false // "CVV cannot be empty."
	}

	// 2. Use the regex to check for exactly 3 numeric digits.
	if !cvv3DigitRegex.MatchString(cvv) {
		return false // "CVV must be exactly 3 numeric digits."
	}

	// If all checks pass
	return true // "CVV is valid."
}
