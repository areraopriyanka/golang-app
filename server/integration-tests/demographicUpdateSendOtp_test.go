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

func (suite *IntegrationTestSuite) RunDemographicUpdateSendOtp(otpType string) {
	e := handler.NewEcho()

	testUser := suite.createTestUser(PartialMasterUserRecordDao{})

	requestData := request.SendMobileVerificationOtpRequest{
		Type: otpType,
	}
	requestBody, err := json.Marshal(requestData)
	suite.Require().NoError(err, "Failed to marshall request body")

	suite.configOTP()
	suite.configEmail()

	req := httptest.NewRequest(http.MethodPut, "/account/customer/demographic-update/mobile", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateLoggedInRegisteredUserContext(testUser.Id, "publicKey", c)

	err = handler.DemographicUpdateSendOtp(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.DemographicUpdateSendOtpResponse

	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")
	suite.Require().NotEmpty(responseBody.OtpId, "OtpId should not be empty")
	suite.Require().NotEmpty(responseBody.OtpExpiryDuration, "OtpExpiryDuration should not be empty")
	suite.Equal(config.Config.Otp.OtpExpiryDuration, responseBody.OtpExpiryDuration)
	suite.Equal(responseBody.MaskedMobileNo, "XXX-XXX-1234", "Masked mobile number should be in the correct format")
	suite.Equal(responseBody.MaskedEmail, "te******@gmail.com", "Masked email should be in the correct format")

	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", responseBody.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to fetch otp record from database")
	suite.Equal(responseBody.OtpId, otpRecord.OtpId)
	suite.Equal(otpType, otpRecord.OtpType)
	suite.Equal("/account/customer/demographic-update/mobile", otpRecord.ApiPath, "Api path must be valid")
	suite.Equal(testUser.MobileNo, otpRecord.MobileNo)
}

func (suite *IntegrationTestSuite) TestDemographicUpdateSendOtpWithSMS() {
	suite.RunDemographicUpdateSendOtp(constant.SMS)
}

func (suite *IntegrationTestSuite) TestDemographicUpdateSendOtpWithCall() {
	suite.RunDemographicUpdateSendOtp(constant.CALL)
}

func (suite *IntegrationTestSuite) TestDemographicUpdateSendOtpWithEmail() {
	suite.RunDemographicUpdateSendOtp(constant.EMAIL)
}
