package security

import (
	"crypto/rand"
	"math/big"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
)

// Note: crypto/rand must be used over math/rand; go docs for math/rand state:
// > This package's outputs might be easily predictable regardless of how it's seeded.
// > For random numbers suitable for security-sensitive work, see the crypto/rand package.

// Generate a password for the user's ledger user account.
func GenerateLedgerPassword() (string, error) {
	const (
		lowercase      = "abcdefghijklmnopqrstuvwxyz"
		uppercase      = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits         = "0123456789"
		symbols        = "!@#$%" // Undocumented, but these cause server errors for other endpoint calls: ^&*()-+[]{}<>?:;~
		allChars       = lowercase + uppercase + digits + symbols
		passwordLength = 32
	)
	charSets := []string{lowercase, uppercase, digits, symbols}
	password := make([]byte, passwordLength)

	// Satisfy the minimum 1 lower, 1 upper, 1 digit, 1 special; they'll be shuffled later
	for i, charSet := range charSets {
		char, err := randomCharFromSet(charSet)
		if err != nil {
			return "", errtrace.Wrap(err)
		}
		password[i] = char
	}

	// Fill the rest at random, since min reqs have been met
	for i := len(charSets); i < passwordLength; i++ {
		char, err := randomCharFromSet(allChars)
		if err != nil {
			return "", errtrace.Wrap(err)
		}
		password[i] = char
	}
	shuffleErr := utils.ShuffleBytes(password)
	if shuffleErr != nil {
		return "", errtrace.Wrap(shuffleErr)
	}
	return string(password), nil
}

func randomCharFromSet(charSet string) (byte, error) {
	max := big.NewInt(int64(len(charSet)))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, errtrace.Wrap(err)
	}
	return charSet[n.Int64()], nil
}
