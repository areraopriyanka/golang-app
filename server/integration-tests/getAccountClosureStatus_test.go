package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/handler"
	"process-api/pkg/security"
	"process-api/pkg/utils"
)

func (suite *IntegrationTestSuite) TestGetAccountClosureStatusClosable() {
	defer SetupMockForLedger(suite).Close()

	user := suite.createTestUser(PartialMasterUserRecordDao{LedgerCustomerNumber: utils.Pointer("6E6F62616C616E636573")})
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)
	_ = suite.createUserAccountCard(user.Id)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPut, "/account/closure-status", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/closure-status")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)

	err := handler.GetAccountClosureStatus(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody handler.AccountClosureStatus
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(responseBody.IsAccountClosable, true, "Account closabity is true")
	suite.Require().Equal(responseBody.AccountBalance, 0.0, "Account balance should be present")
	suite.Require().Equal(responseBody.PreAuthBalance, 0.0, "PreAuth balance should be present")
	suite.Require().Equal(responseBody.HoldBalance, 0.0, "Hold balance should be present")
	suite.Require().Equal(responseBody.PendingTransactions, uint(0), "Pending transactions should be present")
}

func (suite *IntegrationTestSuite) TestGetAccountClosureStatusNotClosable() {
	defer SetupMockForLedger(suite).Close()

	user := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)
	_ = suite.createUserAccountCard(user.Id)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPut, "/account/closure-status", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/closure-status")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)

	err := handler.GetAccountClosureStatus(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody handler.AccountClosureStatus
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(responseBody.IsAccountClosable, false, "Account closabity is false")
	suite.Require().Equal(responseBody.AccountBalance, 10000.0, "Account balance should be present")
	suite.Require().Equal(responseBody.PreAuthBalance, 0.0, "PreAuth balance should be present")
	suite.Require().Equal(responseBody.HoldBalance, 0.0, "Hold balance should be present")
	suite.Require().Equal(responseBody.PendingTransactions, uint(0), "Pending transactions should be present")
}
