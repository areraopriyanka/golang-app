package utils

import (
	"process-api/pkg/validators"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyAndFormatMobileNumberValidNumbers(t *testing.T) {
	phoneNumber, err := validators.VerifyAndFormatMobileNumber("4014567890")
	assert.Nil(t, err)
	assert.Equal(t, "+14014567890", phoneNumber, "should format US phone number as E.164")

	phoneNumber, err = validators.VerifyAndFormatMobileNumber("14014567890")
	assert.Nil(t, err)
	assert.Equal(t, "+14014567890", phoneNumber, "should format US phone number as E.164")

	phoneNumber, err = validators.VerifyAndFormatMobileNumber("+14014567890")
	assert.Nil(t, err)
	assert.Equal(t, "+14014567890", phoneNumber, "should format US phone number as E.164")
}

func TestVerifyAndFormatMobileNumberInvalidNumbers(t *testing.T) {
	phoneNumber, err := validators.VerifyAndFormatMobileNumber("91234567890")
	assert.Error(t, err)
	assert.Equal(t, "", phoneNumber, "should reject India phone number")
}

func TestRemoveCountryCodeFromMobileNumber(t *testing.T) {
	phoneNumber, err := RemoveCountryCodeFromMobileNumber("+12247894561")
	assert.Nil(t, err)
	assert.Equal(t, "2247894561", phoneNumber, "Valid US number without country code")

	phoneNumber, err = RemoveCountryCodeFromMobileNumber("+16463215678")
	assert.Nil(t, err)
	assert.Equal(t, "6463215678", phoneNumber, "Valid US number without country code")

	phoneNumber, err = RemoveCountryCodeFromMobileNumber("+14159871234")
	assert.Nil(t, err)
	assert.Equal(t, "4159871234", phoneNumber, "Valid US number without country code")
}
