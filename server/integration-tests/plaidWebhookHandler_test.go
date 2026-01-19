package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/clock"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/logging"
	"process-api/pkg/plaid"
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"time"

	"github.com/jinzhu/gorm"
)

func (suite *IntegrationTestSuite) sendPlaidWebhook(h *handler.Handler, webhookPayload any) *httptest.ResponseRecorder {
	payloadBytes, err := json.Marshal(webhookPayload)
	suite.Require().NoError(err, "Failed to marshal webhook payload")

	kid := "test-key-id-123"
	privateKey := suite.getPlaidWebhookPrivateKey(kid)
	issuedAt := clock.Now().Add(-1 * time.Minute)
	tokenString := suite.createPlaidWebhookJWT(privateKey, string(payloadBytes), kid, issuedAt)
	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/plaid/webhooks", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Plaid-Verification", tokenString)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/plaid/webhooks")

	err = h.PlaidWebhookHandler(c)
	suite.Require().NoError(err, "Handler should not return an error")
	return rec
}

func (suite *IntegrationTestSuite) TestPlaidWebhookHandler_ValidWebhook_Success() {
	h := suite.newHandler()
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	item := suite.createPlaidItemWithCheckingAndSavingsAccounts(ps, userRecord)

	webhookPayload := plaid.WebhookPayload{
		WebhookType: "ITEM",
		WebhookCode: "ERROR",
		ItemID:      item.PlaidItemID,
	}

	rec := suite.sendPlaidWebhook(h, webhookPayload)
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var itemNowPendingExpiration dao.PlaidItemDao
	err := suite.TestDB.Where("plaid_item_id = ?", item.PlaidItemID).First(&itemNowPendingExpiration).Error
	suite.Require().NoError(err, "Failed to fetch updated plaid item")
	suite.Require().Equal("ERROR", *itemNowPendingExpiration.ItemError, "Item error should be set")

	webhookPayload = plaid.WebhookPayload{
		WebhookType: "ITEM",
		WebhookCode: "LOGIN_REPAIRED",
		ItemID:      item.PlaidItemID,
	}
	rec = suite.sendPlaidWebhook(h, webhookPayload)
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var itemNowHealed dao.PlaidItemDao
	err = suite.TestDB.Where("plaid_item_id = ?", item.PlaidItemID).First(&itemNowHealed).Error
	suite.Require().NoError(err, "Failed to fetch updated plaid item")
	suite.Require().Nil(itemNowHealed.ItemError, "Item error should no longer be set")
}

func (suite *IntegrationTestSuite) TestPlaidWebhookHandler_ValidWebhook_ItemIDNotFound_Success() {
	h := suite.newHandler()
	webhookPayload := plaid.WebhookPayload{
		WebhookType: "ITEM",
		WebhookCode: "LOGIN_REPAIRED",
		ItemID:      "test-123",
	}
	rec := suite.sendPlaidWebhook(h, webhookPayload)
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")
}

func (suite *IntegrationTestSuite) TestPlaidWebhookHandler_ValidWebhook_UnhandledCode_Success() {
	h := suite.newHandler()
	webhookPayload := plaid.WebhookPayload{
		WebhookType: "ITEM",
		WebhookCode: "UNHANDLED_CODE",
		ItemID:      "test-123",
	}
	rec := suite.sendPlaidWebhook(h, webhookPayload)
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")
}

func (suite *IntegrationTestSuite) TestPlaidWebhookHandler_ValidWebhook_UnhandledType_Success() {
	h := suite.newHandler()
	webhookPayload := plaid.WebhookPayload{
		WebhookType: "UNHANDLED_TYPE",
		WebhookCode: "LOGIN_REPAIRED",
		ItemID:      "test-123",
	}
	rec := suite.sendPlaidWebhook(h, webhookPayload)
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")
}

func (suite *IntegrationTestSuite) TestPlaidWebhookHandler_Revoke_Success() {
	h := suite.newHandler()
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	// All hardcoded values are set from mockoon
	plaidItemID := "DWVAAPWq4RHGlEaNyGKRTAnPLaEmo8Cvq7nc0"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0c0"
	err := ps.InsertItem(userRecord.Id, plaidItemID, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userRecord.Id, plaidItemID, unencryptedAccessToken)
	suite.Require().NoError(err)
	var item dao.PlaidItemDao
	err = ps.DB.Where("plaid_item_id=?", plaidItemID).Find(&item).Error
	suite.Require().NoError(err)

	webhookPayload := plaid.WebhookPayload{
		WebhookType: "ITEM",
		WebhookCode: "USER_ACCOUNT_REVOKED",
		ItemID:      item.PlaidItemID,
	}

	rec := suite.sendPlaidWebhook(h, webhookPayload)
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var revokedItem dao.PlaidItemDao
	err = ps.DB.Where("plaid_item_id = ?", item.PlaidItemID).First(&revokedItem).Error
	suite.Require().ErrorIs(err, gorm.ErrRecordNotFound, "Failed to delete plaid item after account revoked event")
	var revokedAccount dao.PlaidAccountDao
	err = ps.DB.Where("plaid_item_id = ?", item.PlaidItemID).First(&revokedAccount).Error
	suite.Require().ErrorIs(err, gorm.ErrRecordNotFound, "Failed to delete plaid account after account revoked event")
}

func (suite *IntegrationTestSuite) TestPlaidWebhookHandler_InvalidSignature_Unauthorized() {
	h := suite.newHandler()

	webhookPayload := plaid.WebhookPayload{
		WebhookType: "ITEM",
		WebhookCode: "PENDING_EXPIRATION",
		ItemID:      "test_item_123",
	}
	payloadBytes, err := json.Marshal(webhookPayload)
	suite.Require().NoError(err, "Failed to marshal webhook payload")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/plaid/webhooks", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Plaid-Verification", "invalid.jwt.token")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/plaid/webhooks")

	err = h.PlaidWebhookHandler(c)
	suite.Require().NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusUnauthorized, rec.Code, "Expected status code 401 Unauthorized")
}

func (suite *IntegrationTestSuite) TestPlaidWebhookHandler_InvalidPayload_BadRequest() {
	h := suite.newHandler()
	invalidPayload := []byte(`{"invalid": "json" "missing_comma": true}`)
	rec := suite.sendPlaidWebhook(h, invalidPayload)
	suite.Require().Equal(http.StatusBadRequest, rec.Code, "Expected status code 400 Bad Request")
}

func (suite *IntegrationTestSuite) TestPlaidWebhookHandler_PendingExpiration_Success() {
	h := suite.newHandler()
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	// All hardcoded values are set from mockoon
	plaidItemID := "DWVAAPWq4RHGlEaNyGKRTAnPLaEmo8Cvq7nc0"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0c0"
	err := ps.InsertItem(userRecord.Id, plaidItemID, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userRecord.Id, plaidItemID, unencryptedAccessToken)
	suite.Require().NoError(err)
	var item dao.PlaidItemDao
	err = ps.DB.Where("plaid_item_id=?", plaidItemID).Find(&item).Error
	suite.Require().NoError(err)

	webhookPayload := plaid.WebhookPayload{
		WebhookType: "ITEM",
		WebhookCode: "PENDING_EXPIRATION",
		ItemID:      item.PlaidItemID,
	}

	rec := suite.sendPlaidWebhook(h, webhookPayload)
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var itemNowPendingExpiration dao.PlaidItemDao
	err = suite.TestDB.Where("plaid_item_id = ?", item.PlaidItemID).First(&itemNowPendingExpiration).Error
	suite.Require().NoError(err, "Failed to fetch updated plaid item")
	suite.Require().True(itemNowPendingExpiration.IsPendingDisconnect, "Item should be marked as pending disconnect")
}

func (suite *IntegrationTestSuite) TestPlaidWebhookHandler_PendingDisconnect_Success() {
	h := suite.newHandler()
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	item := suite.createPlaidItemWithCheckingAndSavingsAccounts(ps, userRecord)

	webhookPayload := plaid.WebhookPayload{
		WebhookType: "ITEM",
		WebhookCode: "PENDING_DISCONNECT",
		ItemID:      item.PlaidItemID,
	}

	rec := suite.sendPlaidWebhook(h, webhookPayload)
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var itemNowPendingDisconnect dao.PlaidItemDao
	err := suite.TestDB.Where("plaid_item_id = ?", item.PlaidItemID).First(&itemNowPendingDisconnect).Error
	suite.Require().NoError(err, "Failed to fetch updated plaid item")
	suite.Require().True(itemNowPendingDisconnect.IsPendingDisconnect, "Item should be marked as pending disconnect")
}

func (suite *IntegrationTestSuite) TestPlaidWebhookHandler_LoginRepairedClearsPendingDisconnect_Success() {
	h := suite.newHandler()
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	item := suite.createPlaidItemWithCheckingAndSavingsAccounts(ps, userRecord)

	err := dao.PlaidItemDao{}.SetIsPendingDisconnect(item.PlaidItemID, true)
	suite.Require().NoError(err)

	err = dao.PlaidItemDao{}.SetItemError(item.PlaidItemID, "ITEM_LOGIN_REQUIRED")
	suite.Require().NoError(err)

	webhookPayload := plaid.WebhookPayload{
		WebhookType: "ITEM",
		WebhookCode: "LOGIN_REPAIRED",
		ItemID:      item.PlaidItemID,
	}

	rec := suite.sendPlaidWebhook(h, webhookPayload)
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var itemNowRepaired dao.PlaidItemDao
	err = suite.TestDB.Where("plaid_item_id = ?", item.PlaidItemID).First(&itemNowRepaired).Error
	suite.Require().NoError(err, "Failed to fetch updated plaid item")
	suite.Require().Nil(itemNowRepaired.ItemError, "Item error should be cleared")
	suite.Require().False(itemNowRepaired.IsPendingDisconnect, "Item should no longer be pending disconnect")
}

func (suite *IntegrationTestSuite) TestPlaidWebhookHandler_AutomaticallyVerified() {
	h := suite.newHandler()
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)
	token, err := security.GenerateOnboardedJwt(userRecord.Id, userPublicKey.PublicKey, nil)
	suite.Require().NoError(err, "Failed to generate JWT token")

	plaidItemID := suite.callPlaidPublicTokenExchangeForAutomatedMicroDeposits(token)

	// We now need to update the accessToken so that mockoon returns a different response:
	encryptedAccessToken, err := utils.EncryptKmsBinary("access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379caa1")
	suite.Require().NoError(err, "Failed to encrypt new access token")

	result := suite.TestDB.Model(&dao.PlaidItemDao{}).Where("plaid_item_id = ?", plaidItemID).Update("kms_encrypted_access_token", encryptedAccessToken)
	suite.Require().NoError(result.Error, "Failed to update new access token")
	suite.Require().Equal(int64(1), result.RowsAffected, "Failed to update new access token")

	webhookPayload := plaid.WebhookPayload{
		WebhookType: "AUTH",
		WebhookCode: "AUTOMATICALLY_VERIFIED",
		AccountID:   utils.Pointer("vzeNDwK7KQIm4yEog683uElbp9GRLEFXGKaa0"), // mockoon
		ItemID:      plaidItemID,
	}

	rec := suite.sendPlaidWebhook(h, webhookPayload)
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var account dao.PlaidAccountDao
	err = suite.TestDB.Model(dao.PlaidAccountDao{}).Where("plaid_item_id=?", plaidItemID).First(&account).Error
	suite.Require().NoError(err, "Failed to get an account")

	suite.Require().Equal("AUTOMATED_MICRODEPOSITS", string(*account.AuthMethod), "Account should be AUTOMATED_MICRODEPOSITS")
	suite.Require().Equal("Baku Endif", *account.PrimaryOwnerName, "Account should be have primary owner name set for automated micro-deposits flow.")
	suite.Require().Equal(int64(10000), *account.AvailableBalanceCents, "Account should not yet have available balance.")
	suite.Require().NotNil(account.BalanceRefreshedAt, "Automated micro-deposit linked accounts should have their balances marked as up-to-date.")
	suite.Require().Equal("automatically_verified", *account.VerificationStatus, "Account should now be automatically verified.")
}
