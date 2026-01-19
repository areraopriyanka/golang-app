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

func (suite *IntegrationTestSuite) TestSalesforceGetBalance() {
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
	req := httptest.NewRequest(http.MethodGet, "/salesforce/accounts/217140/balance", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/salesforce/accounts/:ledgerAccountID/balance")
	c.SetParamNames("ledgerAccountID")
	c.SetParamValues("217140")

	claims := &validator.ValidatedClaims{
		CustomClaims: &salesforce.SalesforceClaims{
			Scope: "read:balance",
		},
	}
	c.Set("salesforce_claims", claims)

	err = salesforce.SalesforceGetBalance(c)

	suite.Require().NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody salesforce.BalanceResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(int64(10000), responseBody.AvailableBalanceCents, "Balance should match ledger response")
}

func (suite *IntegrationTestSuite) TestSalesforceGetBalanceUnauthorizedMissingScope() {
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
	req := httptest.NewRequest(http.MethodGet, "/salesforce/accounts/217140/balance", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/salesforce/accounts/:ledgerAccountID/balance")
	c.SetParamNames("ledgerAccountID")
	c.SetParamValues("217140")

	claims := &validator.ValidatedClaims{
		CustomClaims: &salesforce.SalesforceClaims{
			Scope: "read:transactions", // wrong scope
		},
	}
	c.Set("salesforce_claims", claims)

	err = salesforce.SalesforceGetBalance(c)

	suite.Require().Error(err, "Handler should return an error for missing scope")
	errResp := err.(response.ErrorResponse)
	suite.Require().Equal(http.StatusForbidden, errResp.StatusCode, "Expected status code 403 Forbidden")
}

func (suite *IntegrationTestSuite) TestSalesforceGetBalanceAccountNotFound() {
	defer SetupMockForLedger(suite).Close()

	// Don't create any account in the database

	e := handler.NewEcho()
	nonExistentAccountID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/salesforce/accounts/"+nonExistentAccountID+"/balance", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/salesforce/accounts/:ledgerAccountID/balance")
	c.SetParamNames("ledgerAccountID")
	c.SetParamValues(nonExistentAccountID)

	claims := &validator.ValidatedClaims{
		CustomClaims: &salesforce.SalesforceClaims{
			Scope: "read:balance",
		},
	}
	c.Set("salesforce_claims", claims)

	err := salesforce.SalesforceGetBalance(c)

	suite.Require().Error(err, "Handler should return an error for non-existent account")
	errResp := err.(response.ErrorResponse)
	suite.Require().Equal(http.StatusNotFound, errResp.StatusCode, "Expected status code 404 Not Found")
}

func (suite *IntegrationTestSuite) TestSalesforceGetBalanceMissingClaims() {
	defer SetupMockForLedger(suite).Close()

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/salesforce/accounts/217140/balance", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/salesforce/accounts/:ledgerAccountID/balance")
	c.SetParamNames("ledgerAccountID")
	c.SetParamValues("217140")

	// Don't set any salesforce_claims in the context

	err := salesforce.SalesforceGetBalance(c)

	suite.Require().Error(err, "Handler should return an error for missing claims")
	errResp := err.(response.ErrorResponse)
	suite.Require().Equal(http.StatusUnauthorized, errResp.StatusCode, "Expected status code 401 Unauthorized")
}
