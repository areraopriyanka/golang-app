package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"process-api/pkg/config"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
)

func (suite *IntegrationTestSuite) RunSendResetPasswordOtp(email string) {
	suite.configEmail()
	config.Config.Otp.OtpExpiryDuration = 300000

	_ = suite.createTestUser(PartialMasterUserRecordDao{})

	e := handler.NewEcho()

	queryParams := url.Values{}
	queryParams.Set("email", email)
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/reset-password/send-otp?%s", queryParams.Encode()), nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler.SendResetPasswordOTP(c)
	suite.NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.GenerateEmailOtpResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.NotEmpty(responseBody.OtpExpiryDuration, "OtpExpiryDuration should not be empty")
	suite.NotEmpty(responseBody.OtpId, "OtpId should not be empty")
	suite.Equal(responseBody.OtpExpiryDuration, 300000)
}

func (suite *IntegrationTestSuite) TestSendResetPasswordOtpWithEmail() {
	suite.RunSendResetPasswordOtp("testuser@gmail.com")
}

func (suite *IntegrationTestSuite) TestSendResetPasswordOtpWithoutEmail() {
	suite.RunSendResetPasswordOtp("")
}
