package security

import (
	"strings"
	"testing"
	"unicode"
)

func TestGenerateLedgerPasswordLength(t *testing.T) {
	password, err := GenerateLedgerPassword()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedLength := 32
	if len(password) != expectedLength {
		t.Errorf("Password length mismatch: got %d, want %d", len(password), expectedLength)
	}
}

func TestGenerateLedgerPasswordMinRequirements(t *testing.T) {
	password, err := GenerateLedgerPassword()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var hasLower, hasUpper, hasDigit, hasSymbol bool
	for _, char := range password {
		switch {
		case unicode.IsLower(rune(char)):
			hasLower = true
		case unicode.IsUpper(rune(char)):
			hasUpper = true
		case unicode.IsDigit(rune(char)):
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()-_=+[]{}<>?/|", rune(char)):
			hasSymbol = true
		}
	}

	if !hasLower {
		t.Error("Password does not contain a lowercase letter")
	}
	if !hasUpper {
		t.Error("Password does not contain an uppercase letter")
	}
	if !hasDigit {
		t.Error("Password does not contain a digit")
	}
	if !hasSymbol {
		t.Error("Password does not contain a symbol")
	}
}

func TestGenerateLedgerPasswordRandomness(t *testing.T) {
	const iterations = 10000
	passwords := make(map[string]struct{}, iterations)

	for i := 0; i < iterations; i++ {
		password, err := GenerateLedgerPassword()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if _, exists := passwords[password]; exists {
			t.Errorf("Duplicate password generated: %s", password)
		}
		passwords[password] = struct{}{}
	}
}
