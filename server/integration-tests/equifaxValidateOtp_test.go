package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/security"
	"time"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestEquifaxValidateOtp() {
	config.Config.Debtwise.ApiBase = "http://localhost:5006"

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
		DebtwiseOnboardingStatus: "inProgress",
		DebtwiseCustomerNumber:   &debtwiseCustomerNumber,
		UserStatus:               "ACTIVE",
	}

	err := suite.TestDB.Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	equifaxValidateOtpRequest := handler.EquifaxValidateOtpRequest{
		Code: "1111",
	}
	requestBody, err := json.Marshal(equifaxValidateOtpRequest)
	suite.Require().NoError(err, "Failed to marshal debtwise request")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/debtwise/equifax/otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/debtwise/equifax/otp")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "examplePublicKey", c)

	err = handler.EquifaxValidateOtp(customContext)
	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200")

	var responseBody handler.EquifaxValidateOtpResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Should not fail to unmarshal response")

	suite.Equal(responseBody.NextStep, "consent", "Next step should be consent")
}

func (suite *IntegrationTestSuite) TestEquifaxValidateOtpFailure() {
	config.Config.Debtwise.ApiBase = "http://localhost:5006"

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
		DebtwiseOnboardingStatus: "inProgress",
		DebtwiseCustomerNumber:   &debtwiseCustomerNumber,
		UserStatus:               "ACTIVE",
	}

	err := suite.TestDB.Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	equifaxValidateOtpRequest := handler.EquifaxValidateOtpRequest{
		Code: "",
	}
	requestBody, err := json.Marshal(equifaxValidateOtpRequest)
	suite.Require().NoError(err, "Failed to marshal debtwise request")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/debtwise/equifax/otp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/debtwise/equifax/otp")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "examplePublicKey", c)

	err = handler.EquifaxValidateOtp(customContext)
	suite.Require().Error(err, "Handler should return an error")
}
