package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/logging"
	"process-api/pkg/plaid"
	"process-api/pkg/security"
	"process-api/pkg/utils"
)

func (suite *IntegrationTestSuite) TestPlaidAccountsReconnected_Success() {
	h := suite.newHandler()
	e := handler.NewEcho()

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)
	token, err := security.GenerateOnboardedJwt(userRecord.Id, userPublicKey.PublicKey, nil)
	suite.Require().NoError(err, "Failed to generate JWT token")

	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	item := suite.createPlaidItemWithCheckingAndSavingsAccounts(ps, userRecord)
	err = dao.PlaidItemDao{}.SetItemError(item.PlaidItemID, "LOGIN_REQUIRED")
	suite.Require().NoError(err, "Failed to set plaid item error")

	account0 := plaid.PlaidLinkAccount{ID: "vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK9c0", Type: "depository", Subtype: "checking"}
	accounts := []plaid.PlaidLinkAccount{account0}
	requestPayload := handler.PlaidAccountsReconnectedRequest{LinkSessionID: "test-link-session-id", Accounts: accounts}
	requestBody, err := json.Marshal(requestPayload)
	suite.Require().NoError(err, "Failed to marshall request")

	h.BuildRoutes(e, "", "test")
	req := httptest.NewRequest(http.MethodPost, "/account/plaid/accounts/reconnected", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	updatedAccount, err := dao.PlaidItemDao{}.GetItemByPlaidItemID(item.PlaidItemID)
	suite.Require().NoError(err, "Could not get updated account")
	suite.Require().NotNil(updatedAccount, "Could not get updated account")
	suite.Require().Empty(updatedAccount.ItemError, "Did not clear item error")
	suite.Require().False(updatedAccount.IsPendingDisconnect, "Did not clear isPendingDisconnect")
}

func (suite *IntegrationTestSuite) TestPlaidAccountsReconnected_NotFound() {
	h := suite.newHandler()
	e := handler.NewEcho()

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)
	token, err := security.GenerateOnboardedJwt(userRecord.Id, userPublicKey.PublicKey, nil)
	suite.Require().NoError(err, "Failed to generate JWT token")

	account0 := plaid.PlaidLinkAccount{ID: "vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK9c0", Type: "depository", Subtype: "checking"}
	accounts := []plaid.PlaidLinkAccount{account0}
	requestPayload := handler.PlaidAccountsReconnectedRequest{LinkSessionID: "test-link-session-id", Accounts: accounts}
	requestBody, err := json.Marshal(requestPayload)
	suite.Require().NoError(err, "Failed to marshall request")

	h.BuildRoutes(e, "", "test")
	req := httptest.NewRequest(http.MethodPost, "/account/plaid/accounts/reconnected", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	suite.Require().Equal(http.StatusNotFound, rec.Code, "Should 404 if plaid item cannot be found")
}

func (suite *IntegrationTestSuite) TestPlaidAccountsReconnected_WrongUser() {
	h := suite.newHandler()
	e := handler.NewEcho()

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)
	token, err := security.GenerateOnboardedJwt(userRecord.Id, userPublicKey.PublicKey, nil)
	suite.Require().NoError(err, "Failed to generate JWT token")

	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	wrongUser := suite.createTestUser(PartialMasterUserRecordDao{
		Email:    utils.Pointer("wrong@example.com"),
		MobileNo: utils.Pointer("+14159871235"),
	})
	_ = suite.createPlaidItemWithCheckingAndSavingsAccounts(ps, wrongUser)

	account0 := plaid.PlaidLinkAccount{ID: "vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK9c0", Type: "depository", Subtype: "checking"}
	accounts := []plaid.PlaidLinkAccount{account0}
	requestPayload := handler.PlaidAccountsReconnectedRequest{LinkSessionID: "test-link-session-id", Accounts: accounts}
	requestBody, err := json.Marshal(requestPayload)
	suite.Require().NoError(err, "Failed to marshall request")

	h.BuildRoutes(e, "", "test")
	req := httptest.NewRequest(http.MethodPost, "/account/plaid/accounts/reconnected", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	suite.Require().Equal(http.StatusNotFound, rec.Code, "Should 404 if the user attempts to update item that does not belong to them")
}
