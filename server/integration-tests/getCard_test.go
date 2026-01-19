package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestGetCardDetails() {
	defer SetupMockForLedger(suite).Close()
	sessionId := suite.SetupTestData()
	userPublicKey := suite.createUserPublicKeyRecord(sessionId)
	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/cards", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, userPublicKey.PublicKey, c)

	err := handler.GetCardDetails(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.GetCardResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().NotEmpty(responseBody.Card, "Expected card data")
	suite.Require().NotEmpty(responseBody.Card.CardId, "CardId should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardStatus, "CardStatus should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardStatusRaw, "CardStatusRaw should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardMaskNumber, "CardMaskNumber should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardExpiryDate, "CardExpiryDate should not be empty")
	suite.Require().False(responseBody.Card.IsReIssue, "isReIssue should be false")
	suite.Require().False(responseBody.Card.IsReplace, "isReplace should be false")
	suite.Require().False(responseBody.Card.IsReplaceLocked, "isReplaceLocked should be false")
	suite.Require().NotEmpty(responseBody.Card.ExternalCardId, "ExternalCardId should not be empty")
}

func (suite *IntegrationTestSuite) TestGetCardDetailsForClosedAccount() {
	defer SetupMockForLedger(suite).Close()

	user := suite.createUser()
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)

	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:   "CH0000060090",
		CardId:         "fcf2f39199174939fe437",
		AccountNumber:  "123456789012345",
		AccountStatus:  "CLOSED",
		UserId:         user.Id,
		CardMaskNumber: "************2457",
	}

	card := suite.TestDB.Select("card_holder_id", "card_id", "account_number", "account_status", "user_id", "card_mask_number").Create(&userAccountCard)
	suite.Require().NoError(card.Error, "Failed to insert card and cardHolder data")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/cards", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)

	err := handler.GetCardDetails(customContext)
	suite.Require().NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")
}

func (suite *IntegrationTestSuite) TestGetCardDetailsForReplacedCard() {
	defer SetupMockForLedger(suite).Close()

	user := suite.createUser()
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)

	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH0000060090",
		CardId:        "fcf2f39199174939fe437",
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
		UserId:        user.Id,
		IsReissue:     false,
		IsReplace:     true,
	}

	card := suite.TestDB.Select("card_holder_id", "card_id", "account_number", "account_status", "user_id", "is_reissue", "is_replace").Create(&userAccountCard)
	suite.Require().NoError(card.Error, "Failed to insert card and cardHolder data")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/cards", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)

	err := handler.GetCardDetails(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.GetCardResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().NotEmpty(responseBody.Card, "Expected card data")
	suite.Require().NotEmpty(responseBody.Card.CardId, "CardId should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardStatus, "CardStatus should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardStatusRaw, "CardStatusRaw should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardMaskNumber, "CardMaskNumber should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardExpiryDate, "CardExpiryDate should not be empty")
	suite.Require().False(responseBody.Card.IsReIssue, "isReIssue should be false")
	suite.Require().True(responseBody.Card.IsReplace, "isReplace should be true since old record is marked as replaced")
	suite.Require().False(responseBody.Card.IsReplaceLocked, "isReplaceLocked should be false")
	suite.Require().NotEmpty(responseBody.Card.ExternalCardId, "ExternalCardId should not be empty")
}

func (suite *IntegrationTestSuite) TestGetCardDetailsForReIssuedCard() {
	defer SetupMockForLedger(suite).Close()

	user := suite.createUser()
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)

	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH0000060090",
		CardId:        "fcf2f39199174939fe437",
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
		UserId:        user.Id,
		IsReissue:     true,
	}

	card := suite.TestDB.Select("card_holder_id", "card_id", "account_number", "account_status", "user_id", "is_reissue").Create(&userAccountCard)
	suite.Require().NoError(card.Error, "Failed to insert card and cardHolder data")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/cards", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)

	err := handler.GetCardDetails(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.GetCardResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().NotEmpty(responseBody.Card, "Expected card data")
	suite.Require().NotEmpty(responseBody.Card.CardId, "CardId should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardStatus, "CardStatus should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardStatusRaw, "CardStatusRaw should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardMaskNumber, "CardMaskNumber should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardExpiryDate, "CardExpiryDate should not be empty")
	suite.Require().True(responseBody.Card.IsReIssue, "isReIssue should be true since old record is marked as reIssue")
	suite.Require().False(responseBody.Card.IsReplace, "isReplace should be false")
	suite.Require().False(responseBody.Card.IsReplaceLocked, "isReplaceLocked should be false")
	suite.Require().NotEmpty(responseBody.Card.ExternalCardId, "ExternalCardId should not be empty")
}

func (suite *IntegrationTestSuite) TestGetCardDetailsForReIssuedAndPreviousCardFrozen() {
	defer SetupMockForLedger(suite).Close()

	user := suite.createUser()
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)

	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:   "CH0000060090",
		CardId:         "fcf2f39199174939fe437",
		AccountNumber:  "123456789012345",
		AccountStatus:  "ACTIVE",
		UserId:         user.Id,
		IsReissue:      true,
		PreviousCardId: utils.Pointer("frozen"),
	}

	card := suite.TestDB.Select("card_holder_id", "card_id", "account_number", "account_status", "user_id", "is_reissue", "previous_card_id").Create(&userAccountCard)
	suite.Require().NoError(card.Error, "Failed to insert card and cardHolder data")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/cards", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)

	err := handler.GetCardDetails(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.GetCardResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().NotEmpty(responseBody.Card, "Expected card data")
	suite.Require().NotEmpty(responseBody.Card.CardId, "CardId should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardStatus, "CardStatus should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardStatusRaw, "CardStatusRaw should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardMaskNumber, "CardMaskNumber should not be empty")
	suite.Require().NotEmpty(responseBody.Card.CardExpiryDate, "CardExpiryDate should not be empty")
	suite.Require().True(responseBody.Card.IsReIssue, "isReIssue should be true since old record is marked as reIssue")
	suite.Require().True(responseBody.Card.IsPreviousCardFrozen, "isPreviousCardFrozen should be true since previous card is getdetails returns frozen record from ledger")
	suite.Require().False(responseBody.Card.IsReplace, "isReplace should be false")
	suite.Require().False(responseBody.Card.IsReplaceLocked, "isReplaceLocked should be false")
	suite.Require().NotEmpty(responseBody.Card.ExternalCardId, "ExternalCardId should not be empty")
}

func (suite *IntegrationTestSuite) createUser() dao.MasterUserRecordDao {
	sessionId := uuid.New().String()

	config.Config.Aws.KmsEncryptionKeyId = "test-kms-encryption-key-id"
	encryptedPassword, err := utils.EncryptKmsBinary("@8Kf0exhwDN6$sx@$3nazrABuaVBQxsI")
	suite.Require().NoError(err, "Failed to encrypt example ledgerPassword")

	userRecord := dao.MasterUserRecordDao{
		Id:                         sessionId,
		FirstName:                  "Foo",
		LastName:                   "Bar",
		Email:                      "test.user@gmail.com",
		KmsEncryptedLedgerPassword: []byte(encryptedPassword),
		UserStatus:                 constant.ACTIVE,
		LedgerCustomerNumber:       "100000000006001",
	}

	err = suite.TestDB.Select("id", "email", "first_name", "last_name", "kms_encrypted_ledger_password", "user_status", "ledger_customer_number").Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")
	return userRecord
}
