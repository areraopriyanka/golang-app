package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"time"
)

func (suite *IntegrationTestSuite) TestAccountClosureOTPWithEmail() {
	suite.configEmail()
	suite.configOTP()

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	accountClosureOtpRequest := handler.CreateAccountClosureOtpRequest{
		Type: "EMAIL",
	}
	requestBody, err := json.Marshal(accountClosureOtpRequest)
	suite.Require().NoError(err, "Failed to marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/account/close/send-otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/close/send-otp")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, "example_public_key", c)

	err = handler.SendAccountClosureOTP(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var responseBody handler.CreateAccountClosureOtpResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")
	suite.Require().NotEmpty(responseBody.OtpId, "OtpId should not be empty")
	suite.Require().NotEmpty(responseBody.OtpExpiryDuration, "OtpExpiryDuration should not be empty")
	suite.Equal(responseBody.MaskedMobileNo, "XXX-XXX-1234", "Mobile number is masked in correct format")
	suite.Equal(responseBody.MaskedEmail, "te******@gmail.com", "Email is masked in correct format")

	// Validate that the OTP record was correctly saved in the database
	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", responseBody.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to fetch otp record from database")
	suite.Equal(responseBody.OtpId, otpRecord.OtpId, "OtpId in DB must match response")
	suite.Equal("EMAIL", otpRecord.OtpType, "Expected OTP type to be SMS")
	suite.Equal(config.Config.Otp.OtpExpiryDuration, responseBody.OtpExpiryDuration, "OTP expiry mismatch")
	suite.Equal("/account/close/send-otp", otpRecord.ApiPath, "Api path must match")
}

func (suite *IntegrationTestSuite) GetAccountClosureOtpWithMobile(otpType string) {
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	accountClosureOtpRequest := handler.CreateAccountClosureOtpRequest{
		Type: otpType,
	}
	requestBody, err := json.Marshal(accountClosureOtpRequest)
	suite.Require().NoError(err, "Failed to marshall request body")
	suite.configOTP()

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/account/close/send-otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/close/send-otp")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, "example_public_key", c)

	err = handler.SendAccountClosureOTP(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var responseBody handler.CreateAccountClosureOtpResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")
	suite.Require().NotEmpty(responseBody.OtpId, "OtpId should not be empty")
	suite.Require().NotEmpty(responseBody.OtpExpiryDuration, "OtpExpiryDuration should not be empty")
	suite.Equal(responseBody.MaskedMobileNo, "XXX-XXX-1234", "Mobile number is masked in correct format")
	suite.Equal(responseBody.MaskedEmail, "te******@gmail.com", "Email is masked in correct format")

	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", responseBody.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to fetch otp record from database")
	suite.Equal(otpRecord.OtpId, responseBody.OtpId)
	suite.Equal(otpRecord.OtpType, otpType)
	suite.Equal(config.Config.Otp.OtpExpiryDuration, responseBody.OtpExpiryDuration)
	suite.Equal(otpRecord.ApiPath, "/account/close/send-otp", "Api path must be set for get otp route")
	suite.Equal(otpRecord.MobileNo, userRecord.MobileNo)
}

func (suite *IntegrationTestSuite) TestGetAccountClosureOtpWithSMS() {
	suite.GetAccountClosureOtpWithMobile(constant.SMS)
}

func (suite *IntegrationTestSuite) TestGetAccountClosureOtpWithCall() {
	suite.GetAccountClosureOtpWithMobile(constant.CALL)
}

func (suite *IntegrationTestSuite) TestChallengeAccountClosureOtp() {
	unfreeze := clock.FreezeNow()
	defer unfreeze()

	config.Config.Otp.OtpExpiryDuration = 300000

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	createdAt := clock.Now()
	userOtpRecord := suite.createOtpRecord(userRecord, "/account/close/send-otp", createdAt)

	challengeAccountClosureOtpRequest := request.ChallengeOtpRequest{
		OtpId: userOtpRecord.OtpId,
		Otp:   "123456",
	}
	requestBody, err := json.Marshal(challengeAccountClosureOtpRequest)
	suite.Require().NoError(err, "Failed to insert marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/account/close/verify-otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/close/verify-otp")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, "examplePublicKey", c)

	err = handler.ChallengeAccountClosureOTP(customContext)
	suite.NoError(err, "Handler should not return an error")

	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", userOtpRecord.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to get otp record from db")
	suite.Equal("/account/close/send-otp", otpRecord.ApiPath, "Api path must match sibling route")
	suite.Equal(constant.OTP_VERIFIED, otpRecord.OtpStatus)
	suite.WithinDuration(*otpRecord.UsedAt, clock.Now(), 0, "OTP record used at set to now when otp is challenged successfully")
}

func (suite *IntegrationTestSuite) TestChallengeAccountClosureOtpFailure() {
	config.Config.Otp.OtpExpiryDuration = 300000

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	createdAt := clock.Now().Add(-time.Duration(config.Config.Otp.OtpExpiryDuration)*time.Millisecond - time.Minute)
	userOtpRecord := suite.createOtpRecord(userRecord, "/account/close/send-otp", createdAt)

	challengeAccountClosureOtpRequest := request.ChallengeOtpRequest{
		OtpId: userOtpRecord.OtpId,
		Otp:   "123456",
	}
	requestBody, err := json.Marshal(challengeAccountClosureOtpRequest)
	suite.Require().NoError(err, "Failed to insert marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/account/close/verify-otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/close/verify-otp")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, "examplePublicKey", c)

	err = handler.ChallengeAccountClosureOTP(customContext)
	suite.Require().NotNil(err, "Handler should not return an error")

	e.HTTPErrorHandler(err, c)

	var responseBody response.BadRequestErrors
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(http.StatusBadRequest, rec.Code, "Expected status code 400 badRequest")
	suite.Equal("invalid_otp", responseBody.Errors[0].FieldName, "Invalid field name")
	suite.Equal("One Time Password is expired.", responseBody.Errors[0].Error, "Invalid error message")
}
