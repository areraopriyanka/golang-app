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
	"process-api/pkg/utils"
)

func (suite *IntegrationTestSuite) RunSendChangePasswordOtp(otpType string) {
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	createChallengePasswordOtpRequest := handler.CreateChangePasswordOtpRequest{
		Type: otpType,
	}
	requestBody, err := json.Marshal(createChallengePasswordOtpRequest)
	suite.Require().NoError(err, "Failed to marshall request body")

	config.Config.Otp.OtpExpiryDuration = 300000
	config.Config.Twilio.From = "example_from_address"
	config.Config.Otp.OtpDigits = 6
	config.Config.Twilio.ApiBase = "http://localhost:5003"
	config.Config.Twilio.AuthToken = "fakekeyformock"
	config.Config.Twilio.AccountSid = "ACffffffffffffffffffffffffffffffff"
	utils.InitializeTwilioClient(config.Config.Twilio)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/account/change-password/send-otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/change-password/send-otp")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, "example_public_key", c)

	err = handler.SendChangePasswordOTP(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var responseBody response.OtpResponseWithMaskedNumber
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")

	suite.Equal("XXX-XXX-1234", responseBody.MaskedMobileNo, "Masked Mobile no contains last 4 of user's phone no")

	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", responseBody.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to fetch otp record from database")
	suite.Equal(otpRecord.OtpId, responseBody.OtpId)
	suite.Equal(otpRecord.OtpType, otpType)
	suite.Equal(config.Config.Otp.OtpExpiryDuration, responseBody.OtpExpiryDuration)
	suite.Equal(otpRecord.ApiPath, "/account/change-password/send-otp", "Api path must be set for get otp route")
	suite.Equal(otpRecord.MobileNo, userRecord.MobileNo)
}

func (suite *IntegrationTestSuite) TestSendChangePasswordOtpViaSms() {
	suite.RunSendChangePasswordOtp(constant.SMS)
}

func (suite *IntegrationTestSuite) TestSendChangePasswordOtpViaCall() {
	suite.RunSendChangePasswordOtp(constant.CALL)
}

func (suite *IntegrationTestSuite) TestChallengeChangePasswordOtp() {
	unfreeze := clock.FreezeNow()
	defer unfreeze()
	config.Config.Otp.OtpExpiryDuration = 300000

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	createdAt := clock.Now()
	userOtpRecord := suite.createOtpRecord(userRecord, "/account/change-password/send-otp", createdAt)

	challengeChangepasswordOtpRequest := request.ChallengeOtpRequest{
		OtpId: userOtpRecord.OtpId,
		Otp:   "123456",
	}
	requestBody, err := json.Marshal(challengeChangepasswordOtpRequest)
	suite.Require().NoError(err, "Failed to insert marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/account/change-password/verify-otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/change-password/verify-otp")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, "example_public_key", c)

	err = handler.ChallengeChangePasswordOTP(customContext)
	suite.NoError(err, "Handler should not return an error")

	var responseBody handler.ChallengeChangePasswordOtpResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")

	suite.NotEmpty(responseBody.ResetToken, "ResetToken must be provided by handler")

	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", userOtpRecord.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to get otp record from db")
	suite.Equal(otpRecord.ApiPath, "/account/change-password/send-otp", "Api path must match sibling route")
	suite.Equal(otpRecord.OtpStatus, constant.OTP_VERIFIED)
	suite.WithinDuration(*otpRecord.UsedAt, clock.Now(), 0, "OTP record used at set to now when otp is challenged successfully")

	var user dao.MasterUserRecordDao

	err = suite.TestDB.Model(&dao.MasterUserRecordDao{}).Where("id=?", userRecord.Id).Find(&user).Error
	suite.Require().NoError(err, "Failed to fetch user data")

	suite.Equal(user.ResetToken, responseBody.ResetToken, "User record reset token must be updated to match response body")
}
