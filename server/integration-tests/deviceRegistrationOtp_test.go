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
	"time"
)

func (suite *IntegrationTestSuite) TestGetDeviceRegisterOtpWithEmail() {
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	getDeviceRegistrationOtpRequest := request.GetDeviceRegistrationOtpRequest{
		Type: "EMAIL",
	}
	requestBody, err := json.Marshal(getDeviceRegistrationOtpRequest)
	suite.Require().NoError(err, "Failed to marshall request body")

	// Use sendgrid mock server
	config.Config.Email.ApiBase = "http://localhost:5001"
	config.Config.Otp.OtpExpiryDuration = 300000
	config.Config.Email.TemplateDirectory = "../email-templates/"
	config.Config.Email.Domain = "mg.netxd.com"
	config.Config.Email.ApiKey = "sendgrid-api-key"
	config.Config.Email.FromAddr = "Support@DreamFi.com"
	config.Config.Otp.OtpDigits = 6

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/register-device/otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/register-device/otp")

	customContext := security.GenerateLoggedInUnregisteredUserContext(userRecord.Id, c)

	err = handler.GetDeviceRegistrationOtp(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var responseBody response.OtpResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")

	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", responseBody.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to fetch otp record from database")
	suite.Equal(otpRecord.OtpId, responseBody.OtpId)
	suite.Equal(otpRecord.OtpType, "EMAIL")
	suite.Equal(config.Config.Otp.OtpExpiryDuration, responseBody.OtpExpiryDuration)
	suite.Equal(otpRecord.Email, "testuser@gmail.com")
}

func (suite *IntegrationTestSuite) RunGetDeviceRegisterOtpWithMobile(otpType string) {
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	getDeviceRegistrationOtpRequest := request.GetDeviceRegistrationOtpRequest{
		Type: otpType,
	}
	requestBody, err := json.Marshal(getDeviceRegistrationOtpRequest)
	suite.Require().NoError(err, "Failed to marshall request body")

	config.Config.Otp.OtpExpiryDuration = 300000
	config.Config.Twilio.From = "example_from_address"
	config.Config.Otp.OtpDigits = 6
	config.Config.Twilio.ApiBase = "http://localhost:5003"
	config.Config.Twilio.AuthToken = "fakekeyformock"
	config.Config.Twilio.AccountSid = "ACffffffffffffffffffffffffffffffff"
	utils.InitializeTwilioClient(config.Config.Twilio)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/register-device/otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/register-device/otp")

	customContext := security.GenerateLoggedInUnregisteredUserContext(userRecord.Id, c)

	err = handler.GetDeviceRegistrationOtp(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var responseBody response.OtpResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")

	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", responseBody.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to fetch otp record from database")
	suite.Equal(otpRecord.OtpId, responseBody.OtpId)
	suite.Equal(otpRecord.OtpType, otpType)
	suite.Equal(config.Config.Otp.OtpExpiryDuration, responseBody.OtpExpiryDuration)
	suite.Equal(otpRecord.ApiPath, "/register-device/otp", "Api path must be set for get otp route")
	suite.Equal(otpRecord.MobileNo, userRecord.MobileNo)
}

func (suite *IntegrationTestSuite) TestGetDeviceRegisterOtpWithMobile() {
	suite.RunGetDeviceRegisterOtpWithMobile(constant.SMS)
}

func (suite *IntegrationTestSuite) TestGetDeviceRegisterOtpWithCall() {
	suite.RunGetDeviceRegisterOtpWithMobile(constant.CALL)
}

func (suite *IntegrationTestSuite) TestChallengeDeviceRegisterOtp() {
	unfreeze := clock.FreezeNow()
	defer unfreeze()
	defer SetupMockForLedger(suite).Close()

	config.Config.Otp.OtpExpiryDuration = 300000

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	createdAt := clock.Now()
	userOtpRecord := suite.createOtpRecord(userRecord, "/register-device/otp", createdAt)

	challengeDeviceRegistrationOtpRequest := request.ChallengeDeviceRegistrationOtpRequest{
		OtpId:     userOtpRecord.OtpId,
		OtpValue:  "123456",
		PublicKey: "example_public_key",
	}
	requestBody, err := json.Marshal(challengeDeviceRegistrationOtpRequest)
	suite.Require().NoError(err, "Failed to insert marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/register-device/otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/register-device/otp")

	customContext := security.GenerateLoggedInUnregisteredUserContext(userRecord.Id, c)

	err = handler.ChallengeDeviceRegistrationOtp(customContext)
	suite.NoError(err, "Handler should not return an error")

	tokenWithBearer := rec.Header().Get("Authorization")
	suite.NotEmpty(tokenWithBearer, "Authorization token should be present in response headers")

	claims := security.GetClaimsFromToken(tokenWithBearer)

	suite.Equal(claims.Subject, userRecord.Id, "JWT claim subject should match userId from user record")
	suite.WithinDuration(clock.Now().Add(30*time.Minute), claims.ExpiresAt.Time, 0, "Expiration should match expectation")

	var responseBody response.ChallengeDeviceRegistrationOtpResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")

	suite.Equal(responseBody.KeyId, "ledger_provided_key_id")

	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", userOtpRecord.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to get otp record from db")
	suite.Equal(otpRecord.ApiPath, "/register-device/otp", "Api path must match sibling route")
	suite.Equal(otpRecord.OtpStatus, constant.OTP_VERIFIED)
	suite.WithinDuration(*otpRecord.UsedAt, clock.Now(), 0, "OTP record used at set to now when otp is challenged successfully")

	var userPublicKeyRecord dao.UserPublicKey
	err = suite.TestDB.Model(&dao.UserPublicKey{}).Where("user_id = ?", customContext.UserId).First(&userPublicKeyRecord).Error
	suite.Require().NoError(err, "Failed to get user public key record from db")
	suite.Equal(userPublicKeyRecord.UserId, customContext.UserId, "User Public Key record should be created for user")
	suite.Equal(userPublicKeyRecord.KeyId, responseBody.KeyId, "User Public Key record keyId should match response from ledger")
	suite.Equal(userPublicKeyRecord.PublicKey, challengeDeviceRegistrationOtpRequest.PublicKey, "User Public Key record public key should match public key from client")
	suite.NotEmpty(userPublicKeyRecord.KmsEncryptedApiKey, "User api key must not be empty")
}

func (suite *IntegrationTestSuite) TestChallengeDeviceRegisterWithExpiredOtp() {
	defer SetupMockForLedger(suite).Close()

	config.Config.Otp.OtpExpiryDuration = 300000

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	createdAt := clock.Now().Add(-time.Duration(config.Config.Otp.OtpExpiryDuration)*time.Millisecond - time.Minute)
	userOtpRecord := suite.createOtpRecord(userRecord, "/register-device/otp", createdAt)

	challengeDeviceRegistrationOtpRequest := request.ChallengeDeviceRegistrationOtpRequest{
		OtpId:     userOtpRecord.OtpId,
		OtpValue:  "123456",
		PublicKey: "example_public_key",
	}
	requestBody, err := json.Marshal(challengeDeviceRegistrationOtpRequest)
	suite.Require().NoError(err, "Failed to insert marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/register-device/otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/register-device/otp/verify")

	customContext := security.GenerateLoggedInUnregisteredUserContext(userRecord.Id, c)

	err = handler.ChallengeDeviceRegistrationOtp(customContext)
	suite.Require().NotNil(err, "Handler should return an error for expired OTP")

	errorResponse := err.(response.ErrorResponse)

	suite.Equal(http.StatusBadRequest, errorResponse.StatusCode, "Expected statuscode is BadRequest 400")
	suite.Equal(constant.OTP_EXPIRED_ERROR, errorResponse.ErrorCode, "Error code should indicate expired OTP")
	suite.Contains(errorResponse.Message, constant.OTP_EXPIRED_ERROR_MSG, "Error message should indicate expired OTP")

	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", userOtpRecord.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to get otp record from db")
	suite.Equal(otpRecord.ApiPath, "/register-device/otp", "Api path must match sibling route")
	suite.Equal(otpRecord.OtpStatus, constant.OTP_EXPIRED)
}
