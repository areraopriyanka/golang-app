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

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestChallengeMobileVerificationOtpWithExpiredOtp() {
	defer SetupMockForLedger(suite).Close()

	config.Config.Otp.OtpExpiryDuration = 300000

	userStatus := constant.PHONE_VERIFICATION_OTP_SENT
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	createdAt := clock.Now().Add(-time.Millisecond * time.Duration(600000))
	userOtpRecord := suite.createOtpRecord(testUser, "/onboarding/customer/mobile", createdAt)

	challengeMobileVerificationOtpRequest := request.ChallengeOtpRequest{
		OtpId: userOtpRecord.OtpId,
		Otp:   "123456",
	}
	requestBody, err := json.Marshal(challengeMobileVerificationOtpRequest)
	suite.Require().NoError(err, "Failed to insert marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/onboarding/customer/mobile", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err = handler.VerifyMobileVerificationOtp(customContext)
	suite.Require().NotNil(err, "Handler should return an error for expired OTP")

	e.HTTPErrorHandler(err, c)

	var responseBody response.BadRequestErrors
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(http.StatusBadRequest, rec.Code, "Expected status code 400 badRequest")
	suite.Equal("invalid_otp", responseBody.Errors[0].FieldName, "Invalid field name")
	suite.Equal("One Time Password is expired.", responseBody.Errors[0].Error, "Invalid error message")

	// Validate that the OTP record was correctly saved in the database
	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", userOtpRecord.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to get otp record from db")
	suite.Equal("/onboarding/customer/mobile", otpRecord.ApiPath, "Api path must match sibling route")
	suite.Equal(otpRecord.OtpStatus, constant.OTP_EXPIRED)
}

func (suite *IntegrationTestSuite) TestChallengeMobileVerificationOtpForValidUserState() {
	defer SetupMockForLedger(suite).Close()

	config.Config.Otp.OtpExpiryDuration = 300000

	userStatus := constant.PHONE_VERIFICATION_OTP_SENT

	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	createdAt := clock.Now()
	userOtpRecord := suite.createOtpRecord(testUser, "/onboarding/customer/mobile", createdAt)

	challengeMobileVerificationOtpRequest := request.ChallengeOtpRequest{
		OtpId: userOtpRecord.OtpId,
		Otp:   "123456",
	}
	requestBody, err := json.Marshal(challengeMobileVerificationOtpRequest)
	suite.Require().NoError(err, "Failed to insert marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/onboarding/customer/mobile", bytes.NewReader(requestBody))

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err = handler.VerifyMobileVerificationOtp(customContext)
	suite.Require().NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	// Validate that the OTP record was correctly saved in the database
	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", userOtpRecord.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to get otp record from db")
	suite.Equal("/onboarding/customer/mobile", otpRecord.ApiPath, "Api path must match sibling route")
	suite.Equal(constant.OTP_VERIFIED, otpRecord.OtpStatus)

	// Verify user_status from DB
	suite.ValidateUserStatusInDB(testUser.Id, constant.PHONE_NUMBER_VERIFIED)
}

func (suite *IntegrationTestSuite) TestChallengeMobileVerificationOtpForInValidUserState() {
	defer SetupMockForLedger(suite).Close()

	config.Config.Otp.OtpExpiryDuration = 300000

	userStatus := constant.KYC_PASS
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	createdAt := clock.Now()
	userOtpRecord := suite.createOtpRecord(testUser, "/onboarding/customer/mobile", createdAt)

	challengeMobileVerificationOtpRequest := request.ChallengeOtpRequest{
		OtpId: userOtpRecord.OtpId,
		Otp:   "123456",
	}
	requestBody, err := json.Marshal(challengeMobileVerificationOtpRequest)
	suite.Require().NoError(err, "Failed to insert marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/onboarding/customer/mobile", bytes.NewReader(requestBody))

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err = handler.VerifyMobileVerificationOtp(customContext)

	suite.Require().NotNil(err, "Handler should return an error for invalid state")
	errResp := err.(*response.ErrorResponse)
	suite.Equal(http.StatusPreconditionFailed, errResp.StatusCode, "Expected status code 412 StatusPreconditionFailed")
	suite.Equal(constant.INVALID_USER_STATE, errResp.ErrorCode, "Expected error code INVALID_USER_STATE")
}

func (suite *IntegrationTestSuite) TestChallengeMobileVerificationOtpWithDuplicateMobile() {
	defer SetupMockForLedger(suite).Close()

	config.Config.Otp.OtpExpiryDuration = 300000

	// Create a user with an existing mobile number
	_ = suite.createTestUser(PartialMasterUserRecordDao{})

	// Create a new user with mobile verification pending
	sessionId := uuid.New().String()
	testUser := dao.MasterUserRecordDao{
		Id:         sessionId,
		FirstName:  "Test",
		LastName:   "User",
		Email:      "test.user@gmail.com",
		MobileNo:   "",
		UserStatus: constant.PHONE_VERIFICATION_OTP_SENT,
	}

	err := suite.TestDB.Create(&testUser).Error
	suite.Require().NoError(err, "Failed to create test user")

	otpSessionId := uuid.New().String()

	userOtpRecord := dao.MasterUserOtpDao{
		OtpId:     otpSessionId,
		Otp:       "123456",
		OtpStatus: constant.OTP_SENT,
		Email:     testUser.Email,
		ApiPath:   "/onboarding/customer/mobile",
		UserId:    testUser.Id,
		MobileNo:  "+14159871234",
		IP:        "1.1.1.1",
		CreatedAt: clock.Now(),
	}
	err = suite.TestDB.Select("otp_id", "otp", "otp_status", "email", "api_path", "user_id", "mobile_no", "ip", "created_at").Create(&userOtpRecord).Error
	suite.Require().NoError(err, "Failed to insert otp record")

	challengeMobileVerificationOtpRequest := request.VerifyMobileVerificationOtpRequest{
		OtpId: userOtpRecord.OtpId,
		Otp:   "123456",
	}
	requestBody, err := json.Marshal(challengeMobileVerificationOtpRequest)
	suite.Require().NoError(err, "Failed to insert marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/onboarding/customer/mobile", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err = handler.VerifyMobileVerificationOtp(customContext)
	suite.Require().NotNil(err, "Handler should return an error for duplicate mobile number")
	e.HTTPErrorHandler(err, c)

	var responseBody response.BadRequestErrors
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(http.StatusBadRequest, rec.Code, "Expected status code 400 badRequest")
	suite.Equal("mobile_number", responseBody.Errors[0].FieldName, "Invalid field name")
	suite.Equal("This mobile number is already assigned to another account.", responseBody.Errors[0].Error, "Invalid error message")
}
