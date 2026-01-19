package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) RunSendMobileVerificationOtpForValidUserState(otpType string) {
	e := handler.NewEcho()

	userStatus := constant.AGE_VERIFICATION_PASSED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	requestData := request.SendMobileVerificationOtpRequest{
		MobileNo: "2152409264",
		Type:     otpType,
	}
	requestBody, err := json.Marshal(requestData)
	suite.Require().NoError(err, "Failed to marshall request body")

	suite.configOTP()

	req := httptest.NewRequest(http.MethodPut, "/onboarding/customer/mobile", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err = handler.SendMobileVerificationOtp(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusCreated, rec.Code, "Expected status code 201 Created")

	var responseBody response.OtpResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")

	// Validate that the OTP record was correctly saved in the database
	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", responseBody.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to fetch otp record from database")
	suite.Equal(otpRecord.OtpId, responseBody.OtpId)
	suite.Equal(otpRecord.OtpType, otpType)
	suite.Equal(config.Config.Otp.OtpExpiryDuration, responseBody.OtpExpiryDuration)
	suite.Equal(otpRecord.ApiPath, "/onboarding/customer/mobile", "Api path must be valid")
	suite.Equal(otpRecord.MobileNo, "+12152409264")

	// Verify user_status from DB
	suite.ValidateUserStatusInDB(testUser.Id, constant.PHONE_VERIFICATION_OTP_SENT)
}

func (suite *IntegrationTestSuite) TestSendMobileVerificationOtpViaSms() {
	suite.RunSendMobileVerificationOtpForValidUserState(constant.SMS)
}

func (suite *IntegrationTestSuite) TestSendMobileVerificationOtpViaCall() {
	suite.RunSendMobileVerificationOtpForValidUserState(constant.CALL)
}

func (suite *IntegrationTestSuite) ValidateUserStatusInDB(sessionId, userStatus string) {
	var user dao.MasterUserRecordDao
	result := suite.TestDB.Where("id=?", sessionId).Find(&user)
	suite.Require().NoError(result.Error, "failed to re-query user record")
	suite.Equal(userStatus, user.UserStatus, "Invalid userStatus")
}

func (suite *IntegrationTestSuite) TestSendMobileVerificationOtpForInvalidUserState() {
	e := handler.NewEcho()

	userStatus := constant.USER_CREATED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	requestData := request.SendMobileVerificationOtpRequest{
		MobileNo: "2152409264",
		Type:     constant.SMS,
	}
	requestBody, err := json.Marshal(requestData)
	suite.Require().NoError(err, "Failed to marshall request body")

	suite.configOTP()

	req := httptest.NewRequest(http.MethodPut, "/onboarding/customer/mobile", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err = handler.SendMobileVerificationOtp(customContext)

	suite.Require().NotNil(err, "Handler should return an error for invalid state")
	errResp := err.(*response.ErrorResponse)
	suite.Equal(http.StatusPreconditionFailed, errResp.StatusCode, "Expected status code 412 StatusPreconditionFailed")
	suite.Equal(constant.INVALID_USER_STATE, errResp.ErrorCode, "Expected error code INVALID_USER_STATE")
}
