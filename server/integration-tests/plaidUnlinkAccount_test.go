package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/logging"
	"process-api/pkg/plaid"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestPlaidUnlinkAccount_Success() {
	h := suite.newHandler()
	e := handler.NewEcho()

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)
	_ = suite.createUserAccountCard(userRecord.Id)

	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	// All hardcoded values are set from mockoon
	plaidItemID := "DWVAAPWq4RHGlEaNyGKRTAnPLaEmo8Cvq7nc0"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0c0"
	err := ps.InsertItem(userRecord.Id, plaidItemID, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userRecord.Id, plaidItemID, unencryptedAccessToken)
	suite.Require().NoError(err)
	var item dao.PlaidItemDao
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_item_id=?", plaidItemID).Find(&item).Error
	suite.Require().NoError(err)
	// Make sure it gets both accounts:
	plaidAccountID1 := "vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK9c0"
	var account1 dao.PlaidAccountDao
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_account_id=?", plaidAccountID1).Find(&account1).Error
	suite.Require().NoError(err)
	plaidAccountID2 := "vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK9c1"
	var account2 dao.PlaidAccountDao
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_account_id=?", plaidAccountID2).Find(&account2).Error
	suite.Require().NoError(err)
	requestPayload := handler.PlaidAccountUnlinkRequest{
		PlaidExternalAccountID: account1.ID,
	}

	requestBody, err := json.Marshal(requestPayload)
	suite.Require().NoError(err, "Failed to marshall request")
	req := httptest.NewRequest(http.MethodDelete, "/accounts/plaid/account", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/accounts/plaid/account")
	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, userPublicKey.PublicKey, c)
	err = h.PlaidAccountUnlink(customContext)
	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code)
	var accounts []dao.PlaidAccountDao
	err = db.DB.Where("plaid_account_id = ? OR plaid_account_id = ?", plaidAccountID1, plaidAccountID2).Find(&accounts).Error
	suite.Require().NoError(err, "Failed to query for accounts")
	suite.Require().Len(accounts, 0, "all accounts associated with the deleted account's item must also be deleted")
	var items []dao.PlaidItemDao
	err = db.DB.Where("plaid_item_id = ?", plaidItemID).Find(&items).Error
	suite.Require().NoError(err, "Failed to query for items")
	suite.Require().Len(items, 0, "the Plaid item should also be deleted")
}
