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

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"golang.org/x/crypto/bcrypt"
)

func (suite *IntegrationTestSuite) TestSetCardPinOTPViaSms() {
	suite.configOTP()
	tokenWithBearer, _ := suite.loginAndGetToken()
	rec := suite.makeSetCardPinOTPRequest(tokenWithBearer, map[string]string{"Type": "SMS"})

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.OtpResponseWithMaskedNumber
	err := json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")
	suite.Require().NotEmpty(responseBody.OtpId, "OtpId should not be empty")
	suite.Require().NotEmpty(responseBody.OtpExpiryDuration, "OtpExpiryDuration should not be empty")

	// Validate that the OTP record was correctly saved in the database
	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", responseBody.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to fetch otp record from database")
	suite.Equal(responseBody.OtpId, otpRecord.OtpId, "OtpId in DB must match response")
	suite.Equal("SMS", otpRecord.OtpType, "Expected OTP type to be SMS")
	suite.Equal(config.Config.Otp.OtpExpiryDuration, responseBody.OtpExpiryDuration, "OTP expiry mismatch")
	suite.Equal("XXX-XXX-0000", responseBody.MaskedMobileNo, "Masked mobile not provided in response")
	suite.Equal("/account/cards/pin/otp", otpRecord.ApiPath, "Api path must be valid")
}

func (suite *IntegrationTestSuite) TestSetCardPinOTPViaCall() {
	suite.configOTP()
	tokenWithBearer, _ := suite.loginAndGetToken()
	rec := suite.makeSetCardPinOTPRequest(tokenWithBearer, map[string]string{"Type": "CALL"})

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.OtpResponse
	err := json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")
	suite.Require().NotEmpty(responseBody.OtpId, "OtpId should not be empty")
	suite.Require().NotEmpty(responseBody.OtpExpiryDuration, "OtpExpiryDuration should not be empty")

	// Validate that the OTP record was correctly saved in the database
	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", responseBody.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to fetch otp record from database")
	suite.Equal(responseBody.OtpId, otpRecord.OtpId, "OtpId in DB must match response")
	suite.Equal("CALL", otpRecord.OtpType, "Expected OTP type to be CALL")
	suite.Equal(config.Config.Otp.OtpExpiryDuration, responseBody.OtpExpiryDuration, "OTP expiry mismatch")
	suite.Equal("/account/cards/pin/otp", otpRecord.ApiPath, "Api path must be valid")
}

func (suite *IntegrationTestSuite) TestSetCardPinOTPWrongOTPType() {
	suite.configOTP()
	tokenWithBearer, _ := suite.loginAndGetToken()
	rec := suite.makeSetCardPinOTPRequest(tokenWithBearer, map[string]string{"Type": "videoconference"})

	var responseBody response.BadRequestErrors
	err := json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(http.StatusBadRequest, rec.Code, "Expected status code 400 badRequest")
	suite.Equal("type", responseBody.Errors[0].FieldName, "Invalid field name")
}

func (suite *IntegrationTestSuite) configOTP() {
	config.Config.Otp.OtpExpiryDuration = 300000
	config.Config.Twilio.From = "example_from_address"
	config.Config.Otp.OtpDigits = 6
	config.Config.Twilio.ApiBase = "http://localhost:5003"
	config.Config.Twilio.AuthToken = "fakekeyformock"
	config.Config.Twilio.AccountSid = "ACffffffffffffffffffffffffffffffff"
	utils.InitializeTwilioClient(config.Config.Twilio)
}

func (suite *IntegrationTestSuite) loginAndGetToken() (tokenWithBearer string, userId string) {
	userId = uuid.New().String()
	dob, _ := time.Parse("01/02/2006", "01/02/2000")
	email := "test+1@example.com"
	password := "Password123!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	suite.Require().NoError(err)

	userRecord := dao.MasterUserRecordDao{
		Id:                         userId,
		DOB:                        dob,
		FirstName:                  "John",
		LastName:                   "Doe",
		StreetAddress:              "123 A St.",
		City:                       "Providence",
		State:                      "RI",
		ZipCode:                    "00000",
		Email:                      email,
		LedgerCustomerNumber:       "100000000052004",
		KmsEncryptedLedgerPassword: nil,
		MobileNo:                   "4010000000",
		Password:                   hashedPassword,
		UserStatus:                 constant.ACTIVE,
	}
	err = suite.TestDB.Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	keyId := "exampleKeyId"
	publicKey := "examplePublicKey"
	userPublicKey := dao.UserPublicKey{
		UserId:    userId,
		KeyId:     keyId,
		PublicKey: publicKey,
	}
	err = suite.TestDB.Create(&userPublicKey).Error
	suite.Require().NoError(err, "Failed to insert user public key")

	e := handler.NewEcho()
	e.POST("/login", handler.Login)

	loginBody, _ := json.Marshal(request.LoginRequest{
		Username:  email,
		Password:  password,
		PublicKey: publicKey,
	})

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected successful login")

	tokenWithBearer = rec.Header().Get("Authorization")
	suite.Require().NotEmpty(tokenWithBearer, "Expected token in Authorization header")

	return tokenWithBearer, userId
}

func (suite *IntegrationTestSuite) makeSetCardPinOTPRequest(tokenWithBearer string, request interface{}) *httptest.ResponseRecorder {
	e := handler.NewEcho()
	e.HTTPErrorHandler = utils.CustomHTTPErrorHandler
	e.Use(security.LoggedInRegisteredUserMiddleware)
	e.POST("/account/cards/pin/otp", handler.CreateSetCardPinOTP)

	requestBody, err := json.Marshal(request)
	suite.Require().NoError(err, "Failed to marshall request")
	req := httptest.NewRequest(http.MethodPost, "/account/cards/pin/otp", bytes.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", tokenWithBearer)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func (suite *IntegrationTestSuite) TestChallengeSetCardPinOtp() {
	unfreeze := clock.FreezeNow()
	defer unfreeze()
	defer SetupMockForLedger(suite).Close()

	config.Config.Otp.OtpExpiryDuration = 300000

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	createdAt := clock.Now()
	userOtpRecord := suite.createOtpRecord(userRecord, "/account/cards/pin/otp", createdAt)

	challengeSetCardPinOtpRequest := handler.ChallengeSetCardPinOtpRequest{
		OtpId:    userOtpRecord.OtpId,
		OtpValue: "123456",
	}
	requestBody, err := json.Marshal(challengeSetCardPinOtpRequest)
	suite.Require().NoError(err, "Failed to insert marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/pin/otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/pin/otp")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, "examplePublicKey", c)

	err = handler.ChallengeSetCardPinOtp(customContext)
	suite.NoError(err, "Handler should not return an error")

	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", userOtpRecord.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to get otp record from db")
	suite.Equal("/account/cards/pin/otp", otpRecord.ApiPath, "Api path must match sibling route")
	suite.Equal(constant.OTP_VERIFIED, otpRecord.OtpStatus)
	suite.WithinDuration(*otpRecord.UsedAt, clock.Now(), 0, "OTP record used at set to now when otp is challenged successfully")
}

func (suite *IntegrationTestSuite) TestChallengeSetCardPinOtpFailure() {
	defer SetupMockForLedger(suite).Close()

	config.Config.Otp.OtpExpiryDuration = 300000

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	createdAt := clock.Now().Add(-time.Duration(config.Config.Otp.OtpExpiryDuration)*time.Millisecond - time.Minute)
	userOtpRecord := suite.createOtpRecord(userRecord, "/account/cards/pin/otp", createdAt)

	challengeSetCardPinOtpRequest := handler.ChallengeSetCardPinOtpRequest{
		OtpId:    userOtpRecord.OtpId,
		OtpValue: "123456",
	}
	requestBody, err := json.Marshal(challengeSetCardPinOtpRequest)
	suite.Require().NoError(err, "Failed to insert marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/pin/otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/pin/otp")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, "examplePublicKey", c)

	err = handler.ChallengeSetCardPinOtp(customContext)
	suite.Require().NotNil(err, "Handler should return an error for expired OTP")

	e.HTTPErrorHandler(err, c)

	var responseBody response.BadRequestErrors
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(http.StatusBadRequest, rec.Code, "Expected status code 400 badRequest")
	suite.Equal("invalid_otp", responseBody.Errors[0].FieldName, "Invalid field name")
	suite.Equal("One Time Password is expired.", responseBody.Errors[0].Error, "Invalid error message")
}
