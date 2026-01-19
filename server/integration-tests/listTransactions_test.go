package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestListTransactions() {
	defer SetupMockForLedger(suite).Close()

	sessionId := suite.SetupTestData()
	e := handler.NewEcho()

	userPublicKey := suite.createUserPublicKeyRecord(sessionId)

	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/transactions")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, userPublicKey.PublicKey, c)

	err := handler.ListTransactions(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody handler.ListTransactionsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(responseBody.Count, int64(1), "Count should match")
	suite.Require().NotEmpty(responseBody.Transactions, "Transactions should not be empty")
}

func (suite *IntegrationTestSuite) TestListTransactionsWithEmptyTransactions() {
	defer SetupMockForLedger(suite).Close()

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH0000060090",
		CardId:        "fcf2f39199174939fe437",
		AccountNumber: "emptyTransactions",
		AccountStatus: "ACTIVE",
		UserId:        userRecord.Id,
	}

	err := suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/accounts/transactions", nil)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/accounts/transactions")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, userPublicKey.PublicKey, c)

	err = handler.ListTransactions(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody handler.ListTransactionsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(int64(0), responseBody.Count, "Count should match")
	suite.Require().Empty(responseBody.Transactions, "Transactions should be empty")
}

func (suite *IntegrationTestSuite) TestListTransactionsWithClosedAccount() {
	defer SetupMockForLedger(suite).Close()

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH0000060090",
		CardId:        "fcf2f39199174939fe437",
		AccountNumber: "987546218371925",
		AccountStatus: "CLOSED",
		UserId:        userRecord.Id,
	}

	err := suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/accounts/transactions", nil)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/accounts/transactions")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, userPublicKey.PublicKey, c)

	err = handler.ListTransactions(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody handler.ListTransactionsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(int64(1), responseBody.Count, "Count should match")
	suite.Require().NotEmpty(responseBody.Transactions, "Transactions should not be empty")
}

func (suite *IntegrationTestSuite) TestListTransactionsWithClosedAccountForEmptyTransaction() {
	defer SetupMockForLedger(suite).Close()

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH0000060090",
		CardId:        "fcf2f39199174939fe437",
		AccountNumber: "emptyTransactions",
		AccountStatus: "CLOSED",
		UserId:        userRecord.Id,
	}

	err := suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/accounts/transactions", nil)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/accounts/transactions")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, userPublicKey.PublicKey, c)

	err = handler.ListTransactions(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody handler.ListTransactionsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(int64(0), responseBody.Count, "Count should match")
	suite.Require().Empty(responseBody.Transactions, "Transactions should be empty")
}
