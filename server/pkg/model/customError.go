package model

import (
	"encoding/json"
	"errors"
)

type Error struct {
	StatusCode int
	Response   json.RawMessage
}

var ErrOtpExpired = errors.New("OTP is expired")
