package utils_test

import (
	"bytes"
	"io"
	"process-api/pkg/config"
	"process-api/pkg/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func stubIoReader() *io.Reader {
	data := make([]byte, 32)
	var reader io.Reader = bytes.NewReader(data)
	return &reader
}

func TestGenerateOTPWithRealRand(t *testing.T) {
	config.Config.Otp.OtpDigits = 6
	config.Config.Otp.UseHardcodedOtp = false

	otp, err := utils.GenerateOTP(nil)

	if !assert.NoError(t, err, "GenerateOTP() failed") {
		return
	}

	assert.Equal(t, config.Config.Otp.OtpDigits, len(otp))
}

func TestGenerateOTPDigitLength(t *testing.T) {
	config.Config.Otp.OtpDigits = 4
	config.Config.Otp.UseHardcodedOtp = false

	rand := stubIoReader()

	otp, err := utils.GenerateOTP(rand)

	if !assert.NoError(t, err, "GenerateOTP() failed") {
		return
	}

	assert.Equal(t, "0000", otp)
}

func TestGenerateOTPWithoutHardCoded(t *testing.T) {
	config.Config.Otp.OtpDigits = 6
	config.Config.Otp.UseHardcodedOtp = false

	rand := stubIoReader()

	otp, err := utils.GenerateOTP(rand)

	if !assert.NoError(t, err, "GenerateOTP() failed") {
		return
	}

	assert.Equal(t, "000000", otp)
}

func TestGenerateOTPWithHardCoded(t *testing.T) {
	config.Config.Otp.OtpDigits = 6
	config.Config.Otp.HardcodedOtp = "123456"
	config.Config.Otp.UseHardcodedOtp = true

	rand := stubIoReader()

	otp, err := utils.GenerateOTP(rand)

	if !assert.NoError(t, err, "GenerateOTP() failed") {
		return
	}

	assert.Equal(t, "123456", otp)
}
