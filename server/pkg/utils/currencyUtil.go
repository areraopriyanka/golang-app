package utils

import (
	"github.com/shopspring/decimal"
)

// It's important to round consistently, and it can get tricky when converting
// float64 (ieee-754 double precision) usd amounts to cents. This is especially
// evident when you get a number like `1.015`, which, by money rounding rules,
// should be rounded to `102` cents, but because the fp representation is
// `101.49999999999999`, the (what should be) exact half-cent doesn't trigger
// the proper rounding, and it gets rounded to `101` instead. Using shopspring's
// decimal libary and `RoundBank` uses the appropriate "round half to even" rule,
// aka "banker's rounding":
// https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even
// https://github.com/shopspring/decimal/blob/a1bdfc355e9c71119322b748c95f7d6b82566e30/decimal.go#L1644
func USDtoCents(usd float64) int64 {
	hundred := decimal.New(100, 0)
	return decimal.NewFromFloat(usd).Mul(hundred).RoundBank(0).IntPart()
}

func NilableUSDtoCents(nilableUSD *float64) *int64 {
	if nilableUSD == nil {
		return nil
	}
	usd := *nilableUSD
	cents := USDtoCents(usd)
	return &cents
}

func CentsToUSD(cents int64) float64 {
	hundred := decimal.New(100, 0)
	usd := decimal.NewFromInt(int64(cents)).Div(hundred)
	return usd.InexactFloat64()
}
