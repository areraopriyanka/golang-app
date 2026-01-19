package utils

import "testing"

func TestShuffleBytes(t *testing.T) {
	data := []byte("isogram")
	letterCounts := make(map[byte]int)
	positionCounts := make(map[byte][]int)
	for _, char := range data {
		positionCounts[char] = make([]int, len(data))
	}
	// Increasing iterations allows for the `randomnessThreshold` to get smaller,
	// at the cost of increased running time.
	iterations := 100000
	for i := 0; i < iterations; i++ {
		shuffled := append([]byte{}, data...)
		err := ShuffleBytes(shuffled)
		if err != nil {
			t.Errorf("Error shuffling!: %s", err.Error())
		}
		for pos, char := range shuffled {
			letterCounts[char]++
			positionCounts[char][pos]++
		}
	}
	// The expected probability that a char will occur at any given position after shuffling
	// Each character should have _about_ an equal chance of being shuffled
	// to any given position in the byte array. Because this is a random process,
	// it won't be exact, so a threshold is chosen that will deem the results "random-enough".
	randomnessThreshold := 0.05
	expected := iterations / len(data)
	for _, counts := range positionCounts {
		for pos, count := range counts {
			diff := count - expected
			if diff < 0 {
				diff = expected - count
			}
			ratio := float64(diff) / float64(expected)
			randomEnough := ratio < randomnessThreshold
			if !randomEnough {
				t.Errorf("Error with shuffle distribution! %f at position %d. Count was %d; expected was %d", ratio, pos, count, expected)
			}
		}
	}
}
