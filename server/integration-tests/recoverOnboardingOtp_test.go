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
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) RunRecoverOnboardingOtp(otpType string) {
	userStatus := constant.USER_CREATED
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	createChallengePasswordOtpRequest := handler.RecoverOnboardingOtpRequest{
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
	c.SetPath("/recover-onboarding/send-otp")

	customContext := security.GenerateRecoverOnboardingUserContext(userRecord.Id, c)

	err = handler.RecoverOnboardingOTP(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var responseBody handler.RecoverOnboardingOtpResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")

	suite.Equal("XXX-XXX-1234", responseBody.MaskedMobileNo, "Masked Mobile no contains last 4 of user's phone no")

	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", responseBody.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to fetch otp record from database")
	suite.Equal(otpRecord.OtpId, responseBody.OtpId)
	suite.Equal(otpRecord.OtpType, otpType)
	suite.Equal(config.Config.Otp.OtpExpiryDuration, responseBody.OtpExpiryDuration)
	suite.Equal(otpRecord.ApiPath, "/recover-onboarding/send-otp", "Api path must be set for get otp route")
	suite.Equal(otpRecord.MobileNo, userRecord.MobileNo)
}

func (suite *IntegrationTestSuite) TestRecoverOnboardingOtpViaSms() {
	suite.RunRecoverOnboardingOtp(constant.SMS)
}

func (suite *IntegrationTestSuite) TestRecoverOnboardingOtpViaCall() {
	suite.RunRecoverOnboardingOtp(constant.CALL)
}

func (suite *IntegrationTestSuite) TestRecoverOnboardingOtpViaEmail() {
	suite.RunRecoverOnboardingOtp(constant.EMAIL)
}

func (suite *IntegrationTestSuite) TestChallengeRecoverOnboardingOtp() {
	unfreeze := clock.FreezeNow()
	defer unfreeze()

	config.Config.Otp.OtpExpiryDuration = 300000

	userStatus := constant.USER_CREATED
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	otpSessionId := uuid.New().String()

	userOtpRecord := dao.MasterUserOtpDao{
		OtpId:     otpSessionId,
		Otp:       "123456",
		OtpStatus: constant.OTP_SENT,
		MobileNo:  "1234567890",
		ApiPath:   "/recover-onboarding/send-otp",
		UserId:    userRecord.Id,
		IP:        "1.1.1.1",
		CreatedAt: clock.Now(),
	}

	err := suite.TestDB.Select("otp_id", "otp", "otp_status", "mobile_no", "api_path", "user_id", "ip", "created_at").Create(&userOtpRecord).Error
	suite.Require().NoError(err, "Failed to insert otp record")

	challengeRecoverOnboardingOtpRequest := handler.ChallengeRecoverOnboardingOtpRequest{
		OtpId: userOtpRecord.OtpId,
		Otp:   "123456",
	}
	requestBody, err := json.Marshal(challengeRecoverOnboardingOtpRequest)
	suite.Require().NoError(err, "Failed to insert marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/recover-onboarding/verify-otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/recover-onboarding/verify-otp")

	customContext := security.GenerateRecoverOnboardingUserContext(userRecord.Id, c)

	err = handler.ChallengeRecoverOnboardingOTP(customContext)
	suite.NoError(err, "Handler should not return an error")

	var responseBody handler.ChallengeRecoverOnboardingOtpResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")

	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", userOtpRecord.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to get otp record from db")
	suite.Equal(otpRecord.ApiPath, "/recover-onboarding/send-otp", "Api path must match sibling route")
	suite.Equal(otpRecord.OtpStatus, constant.OTP_VERIFIED)
	suite.WithinDuration(*otpRecord.UsedAt, clock.Now(), 0, "OTP record used at set to now when otp is challenged successfully")

	var user dao.MasterUserRecordDao

	err = suite.TestDB.Model(&dao.MasterUserRecordDao{}).Where("id=?", userRecord.Id).Find(&user).Error
	suite.Require().NoError(err, "Failed to fetch user data")

	suite.Equal(user.UserStatus, responseBody.UserStatus, "User record status should match response body")
	suite.Equal(user.FirstName, responseBody.UserData.FirstName, "User first name should match response body")
	suite.Equal(user.LastName, responseBody.UserData.LastName, "User last name should match response body")
}
