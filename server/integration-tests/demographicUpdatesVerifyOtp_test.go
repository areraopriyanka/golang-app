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

func (suite *IntegrationTestSuite) TestChallengeDemographicUpdatesOtp() {
	config.Config.Otp.OtpExpiryDuration = 300000

	e := handler.NewEcho()

	user := suite.createTestUser(PartialMasterUserRecordDao{})

	createdAt := clock.Now()
	userOtpRecord := suite.createOtpRecord(user, "/account/customer/demographic-update/mobile", createdAt)

	request := request.ChallengeOtpRequest{
		OtpId: userOtpRecord.OtpId,
		Otp:   userOtpRecord.Otp,
	}

	requestBody, marshalErr := json.Marshal(request)
	suite.Require().NoError(marshalErr, "Error while marshaling")

	req := httptest.NewRequest(http.MethodPost, "/account/customer/demographic-update/mobile/verify", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, "publicKey", c)

	err := handler.DemographicUpdateVerifyOtp(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	// Validate that the OTP record was correctly saved in the database
	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", userOtpRecord.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to get otp record from db")
	suite.Equal("/account/customer/demographic-update/mobile", otpRecord.ApiPath, "Api path must match sibling route")
	suite.Equal(constant.OTP_VERIFIED, otpRecord.OtpStatus)
}

func (suite *IntegrationTestSuite) TestChallengeDemographicUpdatesOtpFailure() {
	config.Config.Otp.OtpExpiryDuration = 300000

	e := handler.NewEcho()

	user := suite.createTestUser(PartialMasterUserRecordDao{})

	createdAt := clock.Now().Add(-2 * time.Hour)
	userOtpRecord := suite.createOtpRecord(user, "/account/customer/demographic-update/mobile", createdAt)

	request := request.ChallengeOtpRequest{
		OtpId: userOtpRecord.OtpId,
		Otp:   userOtpRecord.Otp,
	}

	requestBody, marshalErr := json.Marshal(request)
	suite.Require().NoError(marshalErr, "Error while marshaling")

	req := httptest.NewRequest(http.MethodPost, "/account/customer/demographic-update/mobile/verify", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, "publicKey", c)

	err := handler.DemographicUpdateVerifyOtp(customContext)
	suite.Require().NotNil(err, "Handler should return an error for expired OTP")

	e.HTTPErrorHandler(err, c)

	var responseBody response.BadRequestErrors
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(http.StatusBadRequest, rec.Code, "Expected status code 400 badRequest")
	suite.Equal("invalid_otp", responseBody.Errors[0].FieldName, "Invalid field name")
	suite.Equal("One Time Password is expired.", responseBody.Errors[0].Error, "Invalid error message")
}
