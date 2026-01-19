package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (suite *IntegrationTestSuite) TestLatestCreditScore() {
	config.Config.Debtwise.ApiBase = "http://localhost:5006"
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	latestCreditScoreId := uuid.New().String()

	dateCreated, _ := time.Parse("2006-01-02", "2025-07-11")

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

	latestCreditScoreRecord := dao.UserCreditScoreDao{
		Id:                  latestCreditScoreId,
		Date:                dateCreated,
		EncryptedCreditData: encryptedCreditData,
		UserId:              userRecord.Id,
	}

	err = suite.TestDB.Create(&latestCreditScoreRecord).Error
	suite.Require().NoError(err, "Failed to insert latest credit report")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/account/debtwise/latest_credit_score", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/account/debtwise/latest_credit_score")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, "examplePublicKey", c)

	err = handler.LatestCreditScoreHandler(customContext)
	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200")

	var response handler.CreditScoreResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	suite.Require().NoError(err, "Should not fail to unmarshal response")

	suite.Require().NotEmpty(response.CreditScoreHistory, "Credit Score History field must be present")

	suite.Equal(response.CreditAge.Amount, "5 yrs 0 mos", "Credit Age should return formatted Year Month string")

	var creditScoreRecord dao.UserCreditScoreDao
	result := suite.TestDB.Where("user_id=?", userRecord.Id).Find(&creditScoreRecord)
	suite.Require().NoError(result.Error, "failed to fetch credit score record")

	jsonString, err := utils.DecryptKmsBinary(creditScoreRecord.EncryptedCreditData)
	suite.Require().NoError(err, "failed to decryptKmsDecryptKmsBinary encrypted credit data")

	var creditData dao.EncryptableCreditScoreData
	err = json.Unmarshal([]byte(jsonString), &creditData)
	suite.Require().NoError(err, "failed to unmarshall encrypted credit data")

	suite.Equal(response.Score, creditData.Score, "Score was stored successfully")
	suite.Equal(response.Increase, creditData.Increase, "Increase was stored successfully")
	suite.Equal(response.Date, creditScoreRecord.Date.Format("2006-01-02"), "Date was stored successfully")

	suite.Equal(response.PaymentHistory.Amount, creditData.PaymentHistoryAmount, "Payment History Amount was stored successfully")
	suite.Equal(response.PaymentHistory.Factor, creditData.PaymentHistoryFactor, "Payment History Factor was stored successfully")

	suite.Equal(response.DerogatoryMarks.Amount, creditData.DerogatoryMarksAmount, "Derogatory Marks Amount was stored successfully")
	suite.Equal(response.DerogatoryMarks.Factor, creditData.DerogatoryMarksFactor, "Derogatory Marks Factor was stored successfully")

	suite.Equal(response.CreditUtilization.Amount, creditData.CreditUtilizationAmount, "Credit Utilization Amount was stored successfully")
	suite.Equal(response.CreditUtilization.Factor, creditData.CreditUtilizationFactor, "Credit Utilization Factor was stored successfully")

	suite.Equal(response.DerogatoryMarks.Amount, creditData.DerogatoryMarksAmount, "Derogatory Marks Amount was stored successfully")
	suite.Equal(response.DerogatoryMarks.Factor, creditData.DerogatoryMarksFactor, "Derogatory Marks Factor was stored successfully")

	suite.NotEqual(response.CreditAge.Amount, creditData.CreditAgeAmount, "Credit Age is formatted string in response but float in record")
	suite.IsType(creditData.CreditAgeAmount, float64(0))
}

func (suite *IntegrationTestSuite) TestCreditScoreJob() {
	config.Config.Debtwise.ApiBase = "http://localhost:5006"
	h := suite.newHandler()

	debtwiseOnboardingStatus := "complete"
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{DebtwiseOnboardingStatus: &debtwiseOnboardingStatus})

	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)
	e := handler.NewEcho()

	req := httptest.NewRequest(http.MethodPut, "/account/debtwise/equifax/complete", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/debtwise/equifax/complete")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, userPublicKey.PublicKey, c)

	err := h.CompleteDebtwiseOnboarding(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.MaybeRiverJobResponse[handler.CreditScoreResponse]
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	jobId := responseBody.JobId

	suite.IsType(int(0), jobId, "Job ID is present and of type int")
	suite.Require().NotEmpty(responseBody.Payload, "In mock environment credit score is included as payload in polling interval")
	suite.Require().Equal("completed", responseBody.State)

	var statusResponseBody response.MaybeRiverJobResponse[handler.CreditScoreResponse]

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/debtwise/equifax/complete/%d", jobId), nil)
	req2.Header.Set("Content-Type", "application/json")
	c2 := e.NewContext(req2, rec2)
	c2.SetPath("/debtwise/equifax/complete/:jobId")
	c2.SetParamNames("jobId")
	c2.SetParamValues(fmt.Sprintf("%d", jobId))
	customContext2 := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, userPublicKey.PublicKey, c2)

	err = h.GetCreditScoreJobStatus(customContext2)
	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec2.Code, "Expected status code 200 OK")

	err = json.Unmarshal(rec2.Body.Bytes(), &statusResponseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal("completed", statusResponseBody.State, "Response indicates the status of the job as completed")
	suite.Require().NotEmpty(statusResponseBody.JobId, "JobId is present in status response body")

	var creditScoreRecord dao.UserCreditScoreDao
	result := suite.TestDB.Where("user_id=?", userRecord.Id).Find(&creditScoreRecord)
	suite.Require().NoError(result.Error, "failed to fetch credit score record")

	suite.Require().NotEmpty(creditScoreRecord.EncryptedCreditData, "encrypted credut data should be present")
	jsonString, err := utils.DecryptKmsBinary(creditScoreRecord.EncryptedCreditData)
	suite.Require().NoError(err, "failed to decryptKmsDecryptKmsBinary encrypted credit data")

	var creditData dao.EncryptableCreditScoreData
	err = json.Unmarshal([]byte(jsonString), &creditData)
	suite.Require().NoError(err, "failed to unmarshall encrypted credit data")

	// Assertions to verify credit store record was stored correctly upon job completion
	suite.Require().Equal(creditData.Score, statusResponseBody.Payload.Score, "Score was stored successfully and returned in response as payload")
	suite.Require().Equal(creditData.Increase, statusResponseBody.Payload.Increase, "Increase was stored successfully and returned in response as payload")
	suite.Require().Equal(creditScoreRecord.Date.Format("2006-01-02"), statusResponseBody.Payload.Date, "Date was stored successfully and returned in response as payload")

	suite.Require().Equal(creditData.PaymentHistoryAmount, statusResponseBody.Payload.PaymentHistory.Amount, "Payment History Amount was stored successfully and returned in response as payload")
	suite.Require().Equal(creditData.PaymentHistoryFactor, statusResponseBody.Payload.PaymentHistory.Factor, "Payment History Factor was stored successfully and returned in response as payload")

	suite.Require().Equal(creditData.DerogatoryMarksAmount, statusResponseBody.Payload.DerogatoryMarks.Amount, "Derogatory Marks Amount was stored successfully and returned in response as payload")
	suite.Require().Equal(creditData.DerogatoryMarksFactor, statusResponseBody.Payload.DerogatoryMarks.Factor, "Derogatory Marks Factor was stored successfully and returned in response as payload")

	suite.Require().Equal(creditData.CreditUtilizationAmount, statusResponseBody.Payload.CreditUtilization.Amount, "Credit Utilization Amount was stored successfully and returned in response as payload")
	suite.Require().Equal(creditData.CreditUtilizationFactor, statusResponseBody.Payload.CreditUtilization.Factor, "Credit Utilization Factor was stored successfully and returned in response as payload")

	suite.Require().Equal(creditData.NewCreditAmount, statusResponseBody.Payload.NewCredit.Amount, "New Credit Amount was stored successfully and returned in response as payload")
	suite.Require().Equal(creditData.NewCreditFactor, statusResponseBody.Payload.NewCredit.Factor, "New Credit Factor was stored successfully and returned in response as payload")
}
