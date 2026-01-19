package utils

import "github.com/google/uuid"

func GenerateReferenceNumber() string {
	return uuid.New().String()
}
