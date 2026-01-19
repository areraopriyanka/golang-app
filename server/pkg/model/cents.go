package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"braces.dev/errtrace"
)

type TransferCents uint32

const maxCents = 100_000_00 // $100,000.00 expressed in cents

func (c *TransferCents) String() string {
	if c == nil {
		return ""
	}
	return strconv.FormatUint(uint64(*c), 10)
}

func (c *TransferCents) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return errtrace.Wrap(errors.New("cents: empty input"))
	}

	// Allow both JSON string { "cents": "123" }
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return errtrace.Wrap(fmt.Errorf("cents: %w", err))
		}
		return c.fromString(s)
	}

	// ... and JSON number: `{ "cents": 123 }`
	var v uint64
	if err := json.Unmarshal(b, &v); err != nil {
		return errtrace.Wrap(fmt.Errorf("cents: %w", err))
	}
	return c.fromUint64(v)
}

func (c *TransferCents) fromString(s string) error {
	if s == "" {
		return errors.New("cents: empty string")
	}
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("cents: %w", err))
	}
	return c.fromUint64(v)
}

func (c *TransferCents) fromUint64(v uint64) error {
	if v == 0 || v > maxCents {
		return errtrace.Wrap(fmt.Errorf("cents: %d out of range (1-%d))", v, maxCents))
	}
	*c = TransferCents(v)
	return nil
}
