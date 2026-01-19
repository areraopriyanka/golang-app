package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/handler"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"time"
)

func (suite *IntegrationTestSuite) TestVerifyResetPasswordValidOtp() {
	config.Config.Otp.OtpExpiryDuration = 300000

	e := handler.NewEcho()
	testUser := suite.createTestUser(PartialMasterUserRecordDao{})

	createdAt := clock.Now()
	userOtpRecord := suite.createOtpRecord(testUser, "/reset-password/send-otp", createdAt)

	request := request.ChallengeOtpRequest{
		OtpId: userOtpRecord.OtpId,
		Otp:   userOtpRecord.Otp,
	}

	requestBody, marshalErr := json.Marshal(request)
	suite.Require().NoError(marshalErr, "Error while marshaling")

	req := httptest.NewRequest(http.MethodPost, "/reset-password/verify-otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler.VerifyResetPasswordOTP(c)
	suite.NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.VerifyResetPasswordOtpResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")
	suite.Require().NotEmpty(responseBody.ResetToken, "ResetToken should not be empty")
}

func (suite *IntegrationTestSuite) TestVerifyResetPasswordExpiredOtp() {
	config.Config.Otp.OtpExpiryDuration = 300000

	e := handler.NewEcho()

	testUser := suite.createTestUser(PartialMasterUserRecordDao{})
	// Pass createdAt as 2 hours before the current time to check the expiry condition
	createdAt := clock.Now().Add(-2 * time.Hour)
	userOtpRecord := suite.createOtpRecord(testUser, "/reset-password/send-otp", createdAt)

	request := request.ChallengeOtpRequest{
		OtpId: userOtpRecord.OtpId,
		Otp:   userOtpRecord.Otp,
	}

	requestBody, marshalErr := json.Marshal(request)
	suite.Require().NoError(marshalErr, "Error while marshaling")

	req := httptest.NewRequest(http.MethodPost, "/reset-password/verify-otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler.VerifyResetPasswordOTP(c)
	suite.Require().NotNil(err, "Handler should return an error for expired OTP")

	e.HTTPErrorHandler(err, c)

	var responseBody response.BadRequestErrors
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(http.StatusBadRequest, rec.Code, "Expected status code 400 badRequest")
	suite.Equal("invalid_otp", responseBody.Errors[0].FieldName, "Invalid field name")
	suite.Equal("One Time Password is expired.", responseBody.Errors[0].Error, "Invalid error message")
}

// we are doing so to prevent leaking information about the existence/absence of emails in our database.
func (suite *IntegrationTestSuite) TestVerifyResetPasswordOtpForInvalidOTPId() {
	config.Config.Otp.OtpExpiryDuration = 300000

	e := handler.NewEcho()

	testUser := suite.createTestUser(PartialMasterUserRecordDao{})

	createdAt := clock.Now()
	userOtpRecord := suite.createOtpRecord(testUser, "/reset-password/send-otp", createdAt)

	request := request.ChallengeOtpRequest{
		OtpId: "invalid-id",
		Otp:   userOtpRecord.Otp,
	}

	requestBody, marshalErr := json.Marshal(request)
	suite.Require().NoError(marshalErr, "Error while marshaling")

	req := httptest.NewRequest(http.MethodPost, "/reset-password/verify-otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler.VerifyResetPasswordOTP(c)
	suite.Require().NotNil(err, "Handler should return an error for invalid OTP id")

	e.HTTPErrorHandler(err, c)

	var responseBody response.BadRequestErrors
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(http.StatusBadRequest, rec.Code, "Expected status code 400 badRequest")
	suite.Equal("invalid_otp", responseBody.Errors[0].FieldName, "Invalid field name")
	suite.Equal("Please enter valid one time password.", responseBody.Errors[0].Error, "Invalid error message")
}
