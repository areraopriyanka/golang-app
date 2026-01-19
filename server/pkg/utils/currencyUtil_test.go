package utils

import "testing"

// Initial test implementation provided by o3. Every line modified and corrected by hand.
func TestUSDtoCents(t *testing.T) {
	f1 := 0.1
	f2 := 0.2
	f3 := f1 + f2 // this results in 0.30000000000000004 due to ieee-754 double precision
	cases := []struct {
		name     string
		usd      float64
		expected int64
	}{
		{"zero", 0.00, 0},
		{"penny", 0.01, 1},
		{"dime", 0.10, 10},
		{"quarter", 0.25, 25},
		{"29 cents", 0.29, 29},
		{"dollar", 1.00, 100},
		{"large value", 12345678.91, 1_234_567_891},

		{"quirky ieee-754 double precisions 0.30000000000000004", f3, 30},

		// half-cent cases to test bankers rounding
		{"round half to even down", 1.005, 100},
		{"round half to even up", 1.015, 102},

		// Negative values
		{"negative simple", -1.23, -123},
		{"negative half-cent", -1.005, -100},
	}

	for _, testCase := range cases {
		got := USDtoCents(testCase.usd)
		if got != testCase.expected {
			t.Errorf("%s: USDtoCents(%v) = %d, expected %d", testCase.name, testCase.usd, got, testCase.expected)
		}
	}
}

func TestNilableUSDtoCentsWithNil(t *testing.T) {
	if got := NilableUSDtoCents(nil); got != nil {
		t.Errorf("NilableUSDtoCents(nil) = %v, want nil", *got)
	}
}

func TestNilableUSDtoCentsWithPtrFloat64(t *testing.T) {
	val := 12.34
	expected := int64(1234)
	got := NilableUSDtoCents(&val)
	if got != nil {
		if *got != expected {
			t.Errorf("NilableUSDtoCents(%v) = %d, want %d", val, *got, expected)
		}
	} else {
		t.Fatalf("NilableUSDtoCents(%v) returned nil pointer", val)
	}
}

func TestCentsToUSD(t *testing.T) {
	cases := []struct {
		name  string
		cents int64
		usd   float64
	}{
		{"zero", 0, 0.00},
		{"penny", 1, 0.01},
		{"dime", 10, 0.10},
		{"quarter", 25, 0.25},
		{"29 cents", 29, 0.29},
		{"dollar", 100, 1.00},
		{"large value", 12345678, 123456.78},

		{"negative simple", -123, -1.23},
	}

	for _, testCase := range cases {
		got := CentsToUSD(testCase.cents)
		if got != testCase.usd {
			t.Errorf("%s: USDtoCents(%v) = %f, expected %f", testCase.name, testCase.usd, got, testCase.usd)
		}
	}
}
