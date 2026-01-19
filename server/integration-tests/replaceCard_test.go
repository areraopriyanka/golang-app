package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/clock"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/ledger"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"time"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) createReplaceOrReissueCardSignablePayload(customerNo, accountNumber, statusAction, userId string) string {
	reference := uuid.New().String()

	payloadData := ledger.ReplaceOrReissueCardRequest{
		Reference:       reference,
		TransactionType: "REPLACE_CARD",
		CustomerId:      customerNo,
		AccountNumber:   accountNumber,
		CardId:          "6f586be7bf1c44b8b4ea11b2e2510e25",
		Product:         "PL",
		Channel:         "VISA_DPS",
		Program:         "DREAMFI_MVP",
		StatusAction:    statusAction,
	}
	jsonPayloadBytes, err := json.Marshal(payloadData)
	suite.Require().NoError(err, "Failed to marshal test payload")

	jsonPayload := string(jsonPayloadBytes)

	payloadId := uuid.New().String()
	payloadRecord := dao.SignablePayloadDao{
		Id:      payloadId,
		Payload: jsonPayload,
		UserId:  &userId,
	}
	err = suite.TestDB.Create(&payloadRecord).Error
	suite.Require().NoError(err, "Failed to insert test payload")
	return payloadId
}

func (suite *IntegrationTestSuite) TestReplaceCard() {
	defer SetupMockForLedger(suite).Close()

	testUser := suite.createTestUser(PartialMasterUserRecordDao{})

	// Create cardHolderRecord
	cardHolder := dao.UserAccountCardDao{
		CardHolderId:  "CH00000600900090978",
		CardId:        "report_lost_stolen",
		UserId:        testUser.Id,
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
	}

	err := suite.TestDB.Create(&cardHolder).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	payloadId := suite.createReplaceOrReissueCardSignablePayload(testUser.LedgerCustomerNumber, cardHolder.AccountNumber, ledger.REPLACE_CARD, testUser.Id)
	userPublicKey := suite.createUserPublicKeyRecord(testUser.Id)

	replaceCardRequest := request.ReplaceCardRequest{
		LedgerApiRequest: request.LedgerApiRequest{
			Signature: "test_signature",
			Mfp:       "test_mfp",
			PayloadId: payloadId,
		},
	}
	requestBody, _ := json.Marshal(replaceCardRequest)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/replace", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/replace")

	customContext := security.GenerateLoggedInRegisteredUserContext(testUser.Id, userPublicKey.PublicKey, c)

	err = handler.ReplaceCard(customContext)
	suite.Require().NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var cardHolderRecord dao.UserAccountCardDao
	err = suite.TestDB.Model(&dao.UserAccountCardDao{}).Where("card_holder_id = ?", cardHolder.CardHolderId).First(&cardHolderRecord).Error
	suite.Require().NoError(err, "Failed to retrieve cardHolder record")
	suite.Require().NotEmpty(cardHolderRecord.CardId, "Card ID should not be empty")
	suite.Require().NotEqual(cardHolderRecord.CardId, cardHolder.CardId, "Card ID should be updated")
	suite.Require().True(cardHolderRecord.IsReplace, "Card should be updated to indicate it is replaced if lost or stolen")
	suite.Require().False(cardHolderRecord.IsReissue, "Card should not be updated to indicate it is reissued if lost or stolen")
}

func (suite *IntegrationTestSuite) TestReissueCard() {
	defer SetupMockForLedger(suite).Close()

	testUser := suite.createTestUser(PartialMasterUserRecordDao{})

	// Create cardHolderRecord
	cardHolder := dao.UserAccountCardDao{
		CardHolderId:  "CH00000600900090978",
		CardId:        "report_lost_stolen",
		UserId:        testUser.Id,
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
	}

	err := suite.TestDB.Create(&cardHolder).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	payloadId := suite.createReplaceOrReissueCardSignablePayload(testUser.LedgerCustomerNumber, cardHolder.AccountNumber, ledger.REISSUE, testUser.Id)
	userPublicKey := suite.createUserPublicKeyRecord(testUser.Id)

	replaceCardRequest := request.ReplaceCardRequest{
		LedgerApiRequest: request.LedgerApiRequest{
			Signature: "test_signature",
			Mfp:       "test_mfp",
			PayloadId: payloadId,
		},
	}
	requestBody, _ := json.Marshal(replaceCardRequest)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/replace", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/replace")

	customContext := security.GenerateLoggedInRegisteredUserContext(testUser.Id, userPublicKey.PublicKey, c)

	err = handler.ReplaceCard(customContext)
	suite.Require().NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var cardHolderRecord dao.UserAccountCardDao
	err = suite.TestDB.Model(&dao.UserAccountCardDao{}).Where("card_holder_id = ?", cardHolder.CardHolderId).First(&cardHolderRecord).Error
	suite.Require().NoError(err, "Failed to retrieve cardHolder record")
	suite.Require().NotEmpty(cardHolderRecord.CardId, "Card ID should not be empty")
	suite.Require().NotEqual(cardHolderRecord.CardId, cardHolder.CardId, "Card ID should be updated")
	suite.Require().True(cardHolderRecord.IsReissue, "Card should be updated to indicate it is reissued if not lost or stolen")
	suite.Require().False(cardHolderRecord.IsReplace, "Card should not be updated to indicate it is replaced if not lost or stolen")
}

func (suite *IntegrationTestSuite) TestReplaceCardWithActiveReplaceLock() {
	defer SetupMockForLedger(suite).Close()

	testUser := suite.createTestUser(PartialMasterUserRecordDao{})

	// Create cardHolderRecord
	replaceLockExpiresAt := clock.Now().Add(24 * time.Hour)
	cardHolder := dao.UserAccountCardDao{
		CardHolderId:         "CH00000600900090978",
		CardId:               "report_lost_stolen",
		UserId:               testUser.Id,
		AccountNumber:        "123456789012345",
		AccountStatus:        "ACTIVE",
		ReplaceLockExpiresAt: &replaceLockExpiresAt,
	}

	err := suite.TestDB.Create(&cardHolder).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	payloadId := suite.createReplaceOrReissueCardSignablePayload(testUser.LedgerCustomerNumber, cardHolder.AccountNumber, ledger.REPLACE_CARD, testUser.Id)
	userPublicKey := suite.createUserPublicKeyRecord(testUser.Id)

	replaceCardRequest := request.ReplaceCardRequest{
		LedgerApiRequest: request.LedgerApiRequest{
			Signature: "test_signature",
			Mfp:       "test_mfp",
			PayloadId: payloadId,
		},
	}
	requestBody, _ := json.Marshal(replaceCardRequest)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/replace", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/replace")

	customContext := security.GenerateLoggedInRegisteredUserContext(testUser.Id, userPublicKey.PublicKey, c)

	err = handler.ReplaceCard(customContext)
	suite.Require().NotNil(err, "Handler should return an error for active replace lock")
	errResp := err.(response.ErrorResponse)
	suite.Require().Equal(http.StatusConflict, errResp.StatusCode, "Expected status code 409 Conflict")

	var cardHolderRecord dao.UserAccountCardDao
	err = suite.TestDB.Model(&dao.UserAccountCardDao{}).Where("card_holder_id = ?", cardHolder.CardHolderId).First(&cardHolderRecord).Error
	suite.Require().NoError(err, "Failed to retrieve cardHolder record")
	suite.Require().NotEmpty(cardHolderRecord.CardId, "Card ID should not be empty")
	suite.Require().Equal(cardHolderRecord.CardId, cardHolder.CardId, "Card ID should remain the same")
	suite.Require().False(cardHolderRecord.IsReplace, "Card should not be updated to indicate it is replaced if lost or stolen")
	suite.Require().False(cardHolderRecord.IsReissue, "Card should not be updated to indicate it is reissued if lost or stolen")
}

func (suite *IntegrationTestSuite) TestReplaceCardWithExpiredReplaceLock() {
	defer SetupMockForLedger(suite).Close()

	testUser := suite.createTestUser(PartialMasterUserRecordDao{})

	// Create cardHolderRecord
	replaceLockExpiresAt := clock.Now().Add(-24 * time.Hour)
	cardHolder := dao.UserAccountCardDao{
		CardHolderId:         "CH00000600900090978",
		CardId:               "report_lost_stolen",
		UserId:               testUser.Id,
		AccountNumber:        "123456789012345",
		AccountStatus:        "ACTIVE",
		ReplaceLockExpiresAt: &replaceLockExpiresAt,
	}

	err := suite.TestDB.Create(&cardHolder).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	payloadId := suite.createReplaceOrReissueCardSignablePayload(testUser.LedgerCustomerNumber, cardHolder.AccountNumber, ledger.REPLACE_CARD, testUser.Id)
	userPublicKey := suite.createUserPublicKeyRecord(testUser.Id)

	replaceCardRequest := request.ReplaceCardRequest{
		LedgerApiRequest: request.LedgerApiRequest{
			Signature: "test_signature",
			Mfp:       "test_mfp",
			PayloadId: payloadId,
		},
	}
	requestBody, _ := json.Marshal(replaceCardRequest)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/replace", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/replace")

	customContext := security.GenerateLoggedInRegisteredUserContext(testUser.Id, userPublicKey.PublicKey, c)

	err = handler.ReplaceCard(customContext)
	suite.Require().NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var cardHolderRecord dao.UserAccountCardDao
	err = suite.TestDB.Model(&dao.UserAccountCardDao{}).Where("card_holder_id = ?", cardHolder.CardHolderId).First(&cardHolderRecord).Error
	suite.Require().NoError(err, "Failed to retrieve cardHolder record")
	suite.Require().NotEmpty(cardHolderRecord.CardId, "Card ID should not be empty")
	suite.Require().NotEqual(cardHolderRecord.CardId, cardHolder.CardId, "Card ID should be updated")
	suite.Require().True(cardHolderRecord.IsReplace, "Card should be updated to indicate it is replaced if lost or stolen")
	suite.Require().False(cardHolderRecord.IsReissue, "Card should not be updated to indicate it is reissued if lost or stolen")
}

func (suite *IntegrationTestSuite) TestReplaceCardWithInvalidCardStatus() {
	defer SetupMockForLedger(suite).Close()

	testUser := suite.createTestUser(PartialMasterUserRecordDao{})

	// Create cardHolderRecord
	cardHolder := dao.UserAccountCardDao{
		CardHolderId:  "CH00000600900090978",
		CardId:        "5434a20e0a9b4d1bbf8fea3f543f9a05",
		UserId:        testUser.Id,
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
	}

	err := suite.TestDB.Create(&cardHolder).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	payloadId := suite.createReplaceOrReissueCardSignablePayload(testUser.LedgerCustomerNumber, cardHolder.AccountNumber, ledger.REPLACE_CARD, testUser.Id)
	userPublicKey := suite.createUserPublicKeyRecord(testUser.Id)

	replaceCardRequest := request.ReplaceCardRequest{
		LedgerApiRequest: request.LedgerApiRequest{
			Signature: "test_signature",
			Mfp:       "test_mfp",
			PayloadId: payloadId,
		},
	}
	requestBody, _ := json.Marshal(replaceCardRequest)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/replace", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/replace")

	customContext := security.GenerateLoggedInRegisteredUserContext(testUser.Id, userPublicKey.PublicKey, c)

	err = handler.ReplaceCard(customContext)
	suite.Require().Error(err, "Handler should  return an error")

	var responseErr response.ErrorResponse
	suite.Require().ErrorAs(err, &responseErr, "handler should return an ErrorResponse")
	suite.Require().Equal(409, responseErr.StatusCode, "handler should return 409")
}
