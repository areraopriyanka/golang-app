package clock

import (
	"time"
)

var Now = time.Now

func Freeze(t time.Time) func() {
	original := Now
	Now = func() time.Time { return t }
	return func() { Now = original }
}

func FreezeNow() func() {
	original := Now
	// truncate avoids precision differences with pgsql
	now := time.Now().Truncate(time.Second)
	Now = func() time.Time { return now }
	return func() { Now = original }
}
