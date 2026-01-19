package utils

import (
	"time"
)

func IsValidAge(dob time.Time, now func() time.Time) bool {
	currentDate := now()
	age := currentDate.Year() - dob.Year()

	if currentDate.Before(dob.AddDate(age, 0, 0)) {
		age--
	}
	if age >= 18 {
		return true
	}

	return false
}
