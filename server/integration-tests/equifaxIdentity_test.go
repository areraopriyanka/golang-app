package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/db/dao"
	"process-api/pkg/debtwise"
	"process-api/pkg/handler"
	"process-api/pkg/security"
	"time"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestEquifaxIdentity() {
	config.Config.Debtwise.ApiBase = "http://localhost:5006"
	h := suite.newHandler()

	dob, _ := time.Parse("01/02/2006", "10/10/1990")
	sessionId := uuid.New().String()
	debtwiseCustomerNumber := 1
	userRecord := dao.MasterUserRecordDao{
		Id:                       sessionId,
		DOB:                      dob,
		FirstName:                "John",
		LastName:                 "Smith",
		StreetAddress:            "123 A St.",
		City:                     "Salem",
		State:                    "OR",
		ZipCode:                  "00000",
		Email:                    "john_smith@email.com",
		MobileNo:                 "990909000",
		DebtwiseOnboardingStatus: "inProgress",
		DebtwiseCustomerNumber:   &debtwiseCustomerNumber,
		UserStatus:               "ACTIVE",
	}

	err := suite.TestDB.Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/debtwise/equifax/identity", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/debtwise/equifax/identity")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "examplePublicKey", c)

	err = h.EquifaxIdentityHandler(customContext)
	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200")

	var responseBody handler.EquifaxIdentityResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Should not fail to unmarshal response")

	suite.Equal(responseBody.FirstName, userRecord.FirstName, "Handler should return first name from user record")
	suite.Equal(responseBody.LastName, userRecord.LastName, "Handler should return last name from user record")
	suite.Equal(responseBody.DateOfBirth, "1990-10-10", "Handler should return date of birth in ISO8601 format")
	suite.Equal(*responseBody.MaskedSsn, "***-**-0909", "Handler should return mobile number as home phone from user record")
	suite.IsType(&[]debtwise.EquifaxAddress{}, responseBody.Addresses, "Addresses should be a slice of EquifaxAddress")
}
