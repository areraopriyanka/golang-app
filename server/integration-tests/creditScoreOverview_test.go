package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (suite *IntegrationTestSuite) TestCreditScoreOverviewForOnboardedUser() {
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
		DebtwiseCustomerNumber:   &debtwiseCustomerNumber,
		DebtwiseOnboardingStatus: "complete",
		UserStatus:               "ACTIVE",
	}

	err := suite.TestDB.Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	creditScoreId := uuid.New().String()

	encryptableCreditData := dao.EncryptableCreditScoreData{
		Score:                   550,
		Increase:                -5,
		DebtwiseCustomerNumber:  *userRecord.DebtwiseCustomerNumber,
		PaymentHistoryAmount:    97,
		PaymentHistoryFactor:    "Good",
		CreditUtilizationAmount: 10,
		CreditUtilizationFactor: "Good",
		DerogatoryMarksAmount:   2,
		DerogatoryMarksFactor:   "Good",
		CreditAgeAmount:         5,
		CreditAgeFactor:         "Good",
		CreditMixAmount:         3,
		CreditMixFactor:         "Good",
		NewCreditAmount:         1,
		NewCreditFactor:         "Good",
		TotalAccountsAmount:     6,
		TotalAccountsFactor:     "Good",
	}

	jsonData, err := json.Marshal(encryptableCreditData)
	suite.Require().NoError(err, "json marshalling should not return error")

	encryptedCreditData, err := utils.EncryptKmsBinary(string(jsonData))
	suite.Require().NoError(err, "encryption call should not return error")

	creditScoreRecord := dao.UserCreditScoreDao{
		Id:                  creditScoreId,
		UserId:              sessionId,
		EncryptedCreditData: encryptedCreditData,
	}

	err = suite.TestDB.Create(&creditScoreRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/account/debtwise/credit_score_overview", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/account/debtwise/credit_score_overview")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "examplePublicKey", c)

	err = handler.CreditScoreOverviewHandler(customContext)
	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200")

	var response handler.CreditScoreOverviewResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	suite.Require().NoError(err, "Should not fail to unmarshal response")

	suite.Require().True(response.IsUserOnboarded, "Response should indicate user is onboarded")
	suite.Require().Equal(550, *response.Score)
	suite.Require().Equal(-5, *response.Increase)
}

func (suite *IntegrationTestSuite) TestCreditScoreOverviewForNotOnboardedUser() {
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
		DebtwiseCustomerNumber:   &debtwiseCustomerNumber,
		DebtwiseOnboardingStatus: "inProgress",
		UserStatus:               "ACTIVE",
	}

	err := suite.TestDB.Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/account/debtwise/credit_score_overview", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/account/debtwise/credit_score_overview")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "examplePublicKey", c)

	err = handler.CreditScoreOverviewHandler(customContext)
	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200")

	var response handler.CreditScoreOverviewResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	suite.Require().NoError(err, "Should not fail to unmarshal response")

	suite.Require().False(response.IsUserOnboarded, "Response should indicate user is onboarded")
}
