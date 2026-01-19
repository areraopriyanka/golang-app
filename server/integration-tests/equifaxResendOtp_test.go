package test

import (
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/security"
	"time"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestEquifaxResendOtp() {
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

	req := httptest.NewRequest(http.MethodPost, "/debtwise/equifax/otp/resend", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/debtwise/equifax/otp/resend")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "examplePublicKey", c)

	err = h.EquifaxInvokeHandler(customContext)
	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200")

	result := suite.TestDB.Where("id=?", userRecord.Id).Find(&userRecord)
	suite.Require().NoError(result.Error, "failed to re-query user record")

	suite.Require().Equal("inProgress", userRecord.DebtwiseOnboardingStatus, "Debtwise onboarding status should be set to inProgress.")
}
