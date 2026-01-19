package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"time"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestEquifaxInvoke() {
	config.Config.Debtwise.ApiBase = "http://localhost:5006"
	h := suite.newHandler()

	dob, _ := time.Parse("01/02/2006", "01/02/2000")
	sessionId := uuid.New().String()
	debtwiseCustomerNumber := 1
	userRecord := dao.MasterUserRecordDao{
		Id:                       sessionId,
		DOB:                      dob,
		FirstName:                "John",
		LastName:                 "Doe",
		StreetAddress:            "123 A St.",
		City:                     "Salem",
		State:                    "OR",
		ZipCode:                  "00000",
		Email:                    "user@example.com",
		MobileNo:                 "4010000000",
		DebtwiseCustomerNumber:   &debtwiseCustomerNumber,
		DebtwiseOnboardingStatus: "inProgress",
		UserStatus:               "ACTIVE",
	}

	err := suite.TestDB.Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	e := handler.NewEcho()

	req := httptest.NewRequest(http.MethodPost, "/debtwise/equifax/invoke", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/debtwise/equifax/invoke")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "examplePublicKey", c)

	err = h.EquifaxInvokeHandler(customContext)
	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200")

	var responseBody handler.EquifaxInvokeResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Should not fail to unmarshal response")

	nextSteps := []string{"validate_otp", "consent"}
	suite.Contains(nextSteps, responseBody.NextStep, "Next step should be validateOtp or consent.")

	suite.Equal("XXX-XXX-0000", *responseBody.MaskedMobileNo, "Mobile number should be returned and masked")

	result := suite.TestDB.Where("id=?", userRecord.Id).Find(&userRecord)
	suite.Require().NoError(result.Error, "failed to re-query user record")

	suite.Require().Equal("inProgress", userRecord.DebtwiseOnboardingStatus, "Debtwise onboarding status should be set to inProgress.")
}

func (suite *IntegrationTestSuite) TestEquifaxInvokeFailure() {
	config.Config.Debtwise.ApiBase = "http://localhost:5006"
	h := suite.newHandler()

	dob, _ := time.Parse("01/02/2006", "01/02/2000")
	sessionId := uuid.New().String()
	debtwiseCustomerNumber := 1
	userRecord := dao.MasterUserRecordDao{
		Id:                       sessionId,
		DOB:                      dob,
		FirstName:                "DWFail",
		LastName:                 "Doe",
		StreetAddress:            "123 A St.",
		City:                     "Salem",
		State:                    "OR",
		ZipCode:                  "00000",
		Email:                    "user@example.com",
		MobileNo:                 "4010000000",
		DebtwiseCustomerNumber:   &debtwiseCustomerNumber,
		DebtwiseOnboardingStatus: "inProgress",
		UserStatus:               "ACTIVE",
	}

	err := suite.TestDB.Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	e := handler.NewEcho()

	req := httptest.NewRequest(http.MethodPost, "/debtwise/equifax/invoke", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/debtwise/equifax/invoke")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "examplePublicKey", c)

	err = h.EquifaxInvokeHandler(customContext)
	suite.Require().NotNil(err, "Handler should return error")

	errorResponse := err.(response.ErrorResponse)

	suite.Equal(http.StatusInternalServerError, errorResponse.StatusCode, "Expected status code 500 when debtwise returns 422")
}

func (suite *IntegrationTestSuite) TestEquifaxInvokeForOnboardedUser() {
	config.Config.Debtwise.ApiBase = "http://localhost:5006"
	h := suite.newHandler()

	dob, _ := time.Parse("01/02/2006", "01/02/2000")
	sessionId := uuid.New().String()
	debtwiseCustomerNumber := 1
	userRecord := dao.MasterUserRecordDao{
		Id:                       sessionId,
		DOB:                      dob,
		FirstName:                "John",
		LastName:                 "Doe",
		StreetAddress:            "123 A St.",
		City:                     "Salem",
		State:                    "OR",
		ZipCode:                  "00000",
		Email:                    "user@example.com",
		MobileNo:                 "4010000000",
		DebtwiseCustomerNumber:   &debtwiseCustomerNumber,
		DebtwiseOnboardingStatus: "complete",
		UserStatus:               "ACTIVE",
	}

	err := suite.TestDB.Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	e := handler.NewEcho()

	req := httptest.NewRequest(http.MethodPost, "/debtwise/equifax/invoke", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/debtwise/equifax/invoke")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "examplePublicKey", c)

	err = h.EquifaxInvokeHandler(customContext)
	suite.Require().NotNil(err, "Handler should return an error")

	errorResponse := err.(response.ErrorResponse)

	suite.Equal(http.StatusBadRequest, errorResponse.StatusCode, "Expected status code 400")

	result := suite.TestDB.Where("id=?", userRecord.Id).Find(&userRecord)
	suite.Require().NoError(result.Error, "failed to re-query user record")

	suite.Require().Equal("complete", userRecord.DebtwiseOnboardingStatus, "Debtwise onboarding status should remain complete")
}
