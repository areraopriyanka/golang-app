package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/salesforce"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestSalesforceGetTransactions() {
	defer SetupMockForLedger(suite).Close()

	user := suite.createUser()
	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH0000060090",
		CardId:        "fcf2f39199174939fe437",
		AccountNumber: "987546218371925",
		AccountId:     "217140",
		AccountStatus: "ACTIVE",
		UserId:        user.Id,
	}

	err := suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and account data")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/salesforce/accounts/217140/transactions", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/salesforce/accounts/:ledgerAccountID/transactions")
	c.SetParamNames("ledgerAccountID")
	c.SetParamValues("217140")

	claims := &validator.ValidatedClaims{
		CustomClaims: &salesforce.SalesforceClaims{
			Scope: "read:transactions",
		},
	}
	c.Set("salesforce_claims", claims)

	err = salesforce.SalesforceGetTransactions(c)

	suite.Require().NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody []salesforce.Transaction
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Len(responseBody, 1, "Should return 1 transaction")

	transaction := responseBody[0]
	suite.Require().Equal("ledger.ach.transfer_ach_pull_1741989935373707193", transaction.Id, "Transaction ID should match")
	suite.Require().Equal("217140", transaction.AccountId, "Account ID should match")
	suite.Require().Equal(int64(12345), transaction.AmountCents, "Amount should match")
	suite.Require().Equal("DONOR_FIRSTNAME", transaction.Merchant, "Merchant name should match debtor account")
	suite.Require().Equal("ACH_PULL", transaction.Type, "Transaction type should match")
	suite.Require().Equal("COMPLETED", transaction.Status, "Transaction status should match")
	suite.Require().NotEmpty(transaction.Date, "Date should not be empty")
}

func (suite *IntegrationTestSuite) TestSalesforceGetTransactionsEmpty() {
	defer SetupMockForLedger(suite).Close()

	user := suite.createUser()
	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH0000060090",
		CardId:        "fcf2f39199174939fe437",
		AccountNumber: "emptyTransactions", // Special account number that triggers empty response
		AccountId:     "217140",
		AccountStatus: "ACTIVE",
		UserId:        user.Id,
	}

	err := suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and account data")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/salesforce/accounts/217140/transactions", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/salesforce/accounts/:ledgerAccountID/transactions")
	c.SetParamNames("ledgerAccountID")
	c.SetParamValues("217140")

	claims := &validator.ValidatedClaims{
		CustomClaims: &salesforce.SalesforceClaims{
			Scope: "read:transactions",
		},
	}
	c.Set("salesforce_claims", claims)

	err = salesforce.SalesforceGetTransactions(c)

	// The ledger client converts NOT_FOUND_TRANSACTION_ENTRIES errors into empty results
	// So the handler should succeed and return an empty array
	suite.Require().NoError(err, "Handler should not return an error for empty transactions")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody []salesforce.Transaction
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")
	suite.Require().Empty(responseBody, "Should return an empty transactions array")
}

func (suite *IntegrationTestSuite) TestSalesforceGetTransactionsUnauthorizedMissingScope() {
	defer SetupMockForLedger(suite).Close()

	user := suite.createUser()
	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH0000060090",
		CardId:        "fcf2f39199174939fe437",
		AccountNumber: "987546218371925",
		AccountId:     "217140",
		AccountStatus: "ACTIVE",
		UserId:        user.Id,
	}

	err := suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and account data")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/salesforce/accounts/217140/transactions", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/salesforce/accounts/:ledgerAccountID/transactions")
	c.SetParamNames("ledgerAccountID")
	c.SetParamValues("217140")

	claims := &validator.ValidatedClaims{
		CustomClaims: &salesforce.SalesforceClaims{
			Scope: "read:balance", // wrong scope
		},
	}
	c.Set("salesforce_claims", claims)

	err = salesforce.SalesforceGetTransactions(c)

	suite.Require().Error(err, "Handler should return an error for missing scope")
	errResp := err.(response.ErrorResponse)
	suite.Require().Equal(http.StatusForbidden, errResp.StatusCode, "Expected status code 403 Forbidden")
}

func (suite *IntegrationTestSuite) TestSalesforceGetTransactionsAccountNotFound() {
	defer SetupMockForLedger(suite).Close()

	// Don't create any account in the database

	e := handler.NewEcho()
	nonExistentAccountID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/salesforce/accounts/"+nonExistentAccountID+"/transactions", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/salesforce/accounts/:ledgerAccountID/transactions")
	c.SetParamNames("ledgerAccountID")
	c.SetParamValues(nonExistentAccountID)

	claims := &validator.ValidatedClaims{
		CustomClaims: &salesforce.SalesforceClaims{
			Scope: "read:transactions",
		},
	}
	c.Set("salesforce_claims", claims)

	err := salesforce.SalesforceGetTransactions(c)

	suite.Require().Error(err, "Handler should return an error for non-existent account")
	errResp := err.(response.ErrorResponse)
	suite.Require().Equal(http.StatusNotFound, errResp.StatusCode, "Expected status code 404 Not Found")
}

func (suite *IntegrationTestSuite) TestSalesforceGetTransactionsMissingClaims() {
	defer SetupMockForLedger(suite).Close()

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/salesforce/accounts/217140/transactions", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/salesforce/accounts/:ledgerAccountID/transactions")
	c.SetParamNames("ledgerAccountID")
	c.SetParamValues("217140")

	// Don't set any salesforce_claims in the context

	err := salesforce.SalesforceGetTransactions(c)

	suite.Require().Error(err, "Handler should return an error for missing claims")
	errResp := err.(response.ErrorResponse)
	suite.Require().Equal(http.StatusUnauthorized, errResp.StatusCode, "Expected status code 401 Unauthorized")
}

func (suite *IntegrationTestSuite) TestSalesforceGetTransactionsMultipleScopes() {
	defer SetupMockForLedger(suite).Close()

	user := suite.createUser()
	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH0000060090",
		CardId:        "fcf2f39199174939fe437",
		AccountNumber: "987546218371925",
		AccountId:     "217140",
		AccountStatus: "ACTIVE",
		UserId:        user.Id,
	}

	err := suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and account data")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/salesforce/accounts/217140/transactions", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/salesforce/accounts/:ledgerAccountID/transactions")
	c.SetParamNames("ledgerAccountID")
	c.SetParamValues("217140")

	// Set up Salesforce claims with multiple scopes including read:transactions
	claims := &validator.ValidatedClaims{
		CustomClaims: &salesforce.SalesforceClaims{
			Scope: "read:balance read:transactions write:accounts",
		},
	}
	c.Set("salesforce_claims", claims)

	err = salesforce.SalesforceGetTransactions(c)

	suite.Require().NoError(err, "Handler should not return an error when read:transactions scope is present among multiple scopes")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody []salesforce.Transaction
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")
	suite.Require().Len(responseBody, 1, "Should return 1 transaction")
}
