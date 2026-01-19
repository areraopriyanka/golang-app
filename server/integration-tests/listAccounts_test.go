package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/handler"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestListAccounts() {
	defer SetupMockForLedger(suite).Close()

	user := suite.createUser()
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)
	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPut, "/dashboard/list-accounts", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/dashboard/list-accounts")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)

	err := handler.ListAccounts(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody handler.ListAccountAndFirstNameResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().NotEmpty(responseBody.FirstName, "FirstName should not be empty")
	suite.Require().Equal(responseBody.FirstName, "Test")
	suite.NotEmpty(responseBody.Accounts, "Account list should not be empty")
}
