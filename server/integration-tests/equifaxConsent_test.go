package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/security"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (suite *IntegrationTestSuite) TestEquifaxConsent() {
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

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/debtwise/equifax/consent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/debtwise/equifax/consent")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "examplePublicKey", c)

	err = handler.EquifaxConsentHandler(customContext)
	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200")

	var responseBody handler.EquifaxConsentResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Should not fail to unmarshal response")

	suite.Equal(responseBody.NextStep, "identity", "Next step should be identity")
}
