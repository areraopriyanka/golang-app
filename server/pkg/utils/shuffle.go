package utils

import (
	"crypto/rand"
	"math/big"
)

// Shuffles an array of bytes in-place.
// https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
func ShuffleBytes(data []byte) error {
	n := len(data)
	// for i from n−1 down to 1 do
	for i := n - 1; i >= 1; i-- {
		// j ← random integer such that 0 ≤ j ≤ i
		jInt, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		// exchange a[j] and a[i]
		j := int(jInt.Int64())
		dataI := data[i]
		dataJ := data[j]
		data[i] = dataJ
		data[j] = dataI
	}
	return nil
}
