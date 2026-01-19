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

func (suite *IntegrationTestSuite) TestUpdateLinkToken_Success() {
	h := suite.newHandler()
	e := handler.NewEcho()

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)
	token, err := security.GenerateOnboardedJwt(userRecord.Id, userPublicKey.PublicKey, nil)
	suite.Require().NoError(err, "Failed to generate JWT token")

	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	item := suite.createPlaidItemWithCheckingAndSavingsAccounts(ps, userRecord)

	var account dao.PlaidAccountDao
	err = suite.TestDB.Model(dao.PlaidAccountDao{}).Where("plaid_item_id=?", item.PlaidItemID).First(&account).Error
	suite.Require().NoError(err, "Failed to get an account")
	requestPayload := handler.PlaidUpdateLinkTokenRequest{Platform: "ios", AccountID: &account.ID}
	requestBody, err := json.Marshal(requestPayload)
	suite.Require().NoError(err, "Failed to marshall request")

	h.BuildRoutes(e, "", "test")
	req := httptest.NewRequest(http.MethodPost, "/account/plaid/link/token/update", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var body handler.PlaidLinkTokenResponse
	suite.NoError(json.NewDecoder(bytes.NewReader(rec.Body.Bytes())).Decode(&body))
	suite.Equal("link-sandbox-5c2ad447-5e16-4905-90f1-317cfc0a7067", body.LinkToken, "linkToken must not be empty, and must match mocked value")
}

func (suite *IntegrationTestSuite) TestUpdateLinkToken_WrongUser() {
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
	item := suite.createPlaidItemWithCheckingAndSavingsAccounts(ps, wrongUser)

	var account dao.PlaidAccountDao
	err = suite.TestDB.Model(dao.PlaidAccountDao{}).Where("plaid_item_id=?", item.PlaidItemID).First(&account).Error
	suite.Require().NoError(err, "Failed to get an account")
	requestPayload := handler.PlaidUpdateLinkTokenRequest{Platform: "ios", AccountID: &account.ID}

	requestBody, err := json.Marshal(requestPayload)
	suite.Require().NoError(err, "Failed to marshall request")

	h.BuildRoutes(e, "", "test")
	req := httptest.NewRequest(http.MethodPost, "/account/plaid/link/token/update", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Require().Equal(http.StatusNotFound, rec.Code, "Should 404 if the user attempts to update item that does not belong to them")
}
