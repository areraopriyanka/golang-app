package model

import (
	"encoding/json"
	"testing"
)

func TestCents_UnmarshalJSON(t *testing.T) { // o3 generated; hand modified
	type payload struct {
		AmountCents TransferCents `json:"amountCents"`
	}

	tests := []struct {
		name  string
		input string
		want  TransferCents
		err   bool
	}{
		{
			name:  "string value - ok",
			input: `{"amountCents":"12345"}`,
			want:  TransferCents(12345),
		},
		{
			name:  "numeric value - ok",
			input: `{"amountCents":67890}`,
			want:  TransferCents(67890),
		},
		{
			name:  "zero - error",
			input: `{"amountCents":"0"}`,
			err:   true,
		},
		{
			name:  "above maximum - error",
			input: `{"amountCents":10000001}`,
			err:   true,
		},
		{
			name:  "negative (not uint) - error",
			input: `{"amountCents":"-1"}`,
			err:   true,
		},
		{
			name:  "non-numeric string - error",
			input: `{"amountCents":"abc"}`,
			err:   true,
		},
		{
			name:  "empty string - error",
			input: `{"amountCents":""}`,
			err:   true,
		},
		{
			name:  "float - error",
			input: `{"amountCents":"1.23"}`,
			err:   true,
		},
		{
			name:  "bignum float - error",
			input: `{"amountCents":"9223372036854775808.23"}`,
			err:   true,
		},
		{
			name:  "exactly max - ok",
			input: `{"amountCents":10000000}`,
			want:  TransferCents(maxCents),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var p payload
			err := json.Unmarshal([]byte(tc.input), &p)

			if tc.err {
				if err == nil {
					t.Fatalf("expected error, got nil (value=%v)", p.AmountCents)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.AmountCents != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, p.AmountCents)
			}
		})
	}
}
