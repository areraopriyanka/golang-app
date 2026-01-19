package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/crypto"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/request"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestWebhookHandlerWithAchPull() {
	defer SetupMockForLedger(suite).Close()
	mockSardineServer := setupMockForSardine(suite)
	defer mockSardineServer.Close()

	config.Config.Sardine.ApiBase = mockSardineServer.URL

	h := suite.newHandler()

	publicKey, privateKey, err := crypto.CreateKeys()
	suite.Require().NoError(err, "Failed to generate keys for signing")

	config.Config.Webhook.PublicKey = publicKey

	sessionId := suite.SetupTestData()
	userPublicKey := suite.createUserPublicKeyRecord(sessionId)

	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH000006009011111",
		CardId:        "6f586be7bf1c44b8b4ea11b2e2510e25",
		UserId:        sessionId,
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
	}
	err = suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	rawPayload := map[string]interface{}{
		"channel": "ACH",
		"credit":  true,
		"creditorAccount": map[string]interface{}{
			"accountNumber":       "123456789012345",
			"customerAccountName": "Example Account",
			"customerAccountType": "CHECKING",
			"holderId":            "1802af7fa0a9318f108376772b394e91",
			"holderIdType":        "SSN",
			"holderName":          "Test User",
			"institutionId":       "124303298",
		},
		"customerID": "100000000033024",
		"debtorAccount": map[string]interface{}{
			"accountNumber": "987546218371925",
			"holderName":    "DONOR_FIRSTNAME",
			"institutionId": "011002550",
		},
		"from":               "3a65dce2329722437a762664b5acbf85",
		"instructedAmount":   1300,
		"instructedCurrency": "USD",
		"ledgerTxnNumber":    "QA00000000027046",
		"postedDate":         "2025-07-21T22:17:19.978Z",
		"referenceID":        "ledger.ach.transfer_ach_pull_1753136239654675000",
		"source":             "API",
		"status":             "PENDING",
		"subTransactionType": "ACH_PULL",
		"transactionNumber":  "QA00000000027046",
		"transactionType":    "ACH_PULL",
	}

	payloadBytes, err := json.Marshal(rawPayload)
	suite.Require().NoError(err, "Failed to marshal raw payload")

	signature, err := crypto.SignECDSA(payloadBytes, privateKey)
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	ledgerEventPayload := request.LedgerEventPayload{
		Source:    "PL",
		EventId:   "EVT45418",
		EventName: "Transaction.NEW",
		Signature: signature,
		Payload:   payloadBytes,
	}
	requestBody, err := json.Marshal(ledgerEventPayload)
	suite.Require().NoError(err, "Failed to marshall request")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/ledger/events", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/ledger/events")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, userPublicKey.PublicKey, c)

	err = h.LedgerWebhookHandler(customContext)

	suite.WaitForJobsDone(1)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var webhookRecord dao.LedgerTransactionEventDao
	err = suite.TestDB.Model(&dao.LedgerTransactionEventDao{}).Where("user_id = ?", sessionId).First(&webhookRecord).Error
	suite.Require().NoError(err, "Failed to fetch updated user data")

	suite.Require().Equal(webhookRecord.UserId, sessionId, "UserId was saved")
	suite.Require().NotEmpty(webhookRecord.AccountNumber, "Account number is present on record")
	suite.Require().NotEmpty(webhookRecord.AccountRoutingNumber, "Account number is present on record")
	suite.Require().NotEmpty(webhookRecord.Channel, "Channel is present on record")
	suite.Require().NotEmpty(webhookRecord.EventId, "Account number is present on record")
	suite.Require().Equal(webhookRecord.TransactionType, "ACH_PULL", "Transaction type is ACH PULL")
	suite.Require().Equal(webhookRecord.InstructedCurrency, "USD", "Transaction currency is save as USD")
	suite.Require().NotEmpty(webhookRecord.InstructedAmount, "Transaction amount is present on record")
	suite.Require().False(webhookRecord.IsOutward, "Record indicates incoming funds")
	suite.Require().NotEmpty(webhookRecord.ExternalBankAccountNumber, "Bank account number is present")
	suite.Require().NotEmpty(webhookRecord.ExternalBankAccountName, "Bank account name is present")
	suite.Require().NotEmpty(webhookRecord.ExternalBankAccountRoutingNumber, "Bank account routing number is present")
	suite.Require().NotEmpty(webhookRecord.RawPayload, "Raw payload is present for ach pull")
}

func (suite *IntegrationTestSuite) TestWebhookHandlerWithAchPush() {
	defer SetupMockForLedger(suite).Close()
	mockSardineServer := setupMockForSardine(suite)
	defer mockSardineServer.Close()

	config.Config.Sardine.ApiBase = mockSardineServer.URL

	h := suite.newHandler()

	publicKey, privateKey, err := crypto.CreateKeys()
	suite.Require().NoError(err, "Failed to generate keys for signing")

	config.Config.Webhook.PublicKey = publicKey

	sessionId := suite.SetupTestData()
	userPublicKey := suite.createUserPublicKeyRecord(sessionId)

	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH000006009011111",
		CardId:        "6f586be7bf1c44b8b4ea11b2e2510e25",
		UserId:        sessionId,
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
	}
	err = suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	rawPayload := map[string]interface{}{
		"avalBalance": 3690,
		"channel":     "ACH",
		"credit":      false,
		"creditorAccount": map[string]interface{}{
			"accountNumber": "987546218371925",
			"holderName":    "BENEFACTOR_FIRSTNAME",
			"institutionId": "011002550",
		},
		"customerID": "100000000033024",
		"debtorAccount": map[string]interface{}{
			"accountNumber":       "123456789012345",
			"customerAccountName": "Default DreamFi Account",
			"customerAccountType": "CHECKING",
			"holderId":            "1802af7fa0a9318f108376772b394e91",
			"holderIdType":        "SSN",
			"holderName":          "Test User",
			"institutionId":       "124303298",
		},
		"from":               "3a65dce2329722437a762664b5acbf85",
		"instructedAmount":   110,
		"instructedCurrency": "USD",
		"ledgerTxnNumber":    "QA00000000027059",
		"postedDate":         "2025-07-22T18:33:37.943Z",
		"referenceID":        "ledger.ach.transfer_ach_out_1753209217598398000",
		"source":             "API",
		"status":             "COMPLETED",
		"subTransactionType": "ACH_OUT",
		"transactionNumber":  "QA00000000027059",
		"transactionType":    "ACH_OUT",
	}

	payloadBytes, err := json.Marshal(rawPayload)
	suite.Require().NoError(err, "Failed to marshal raw payload")

	signature, err := crypto.SignECDSA(payloadBytes, privateKey)
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	ledgerEventPayload := request.LedgerEventPayload{
		Source:    "PL",
		EventId:   "EVT45418",
		EventName: "Transaction.NEW",
		Signature: signature,
		Payload:   payloadBytes,
	}
	requestBody, err := json.Marshal(ledgerEventPayload)
	suite.Require().NoError(err, "Failed to marshall request")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/ledger/events", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/ledger/events")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, userPublicKey.PublicKey, c)

	err = h.LedgerWebhookHandler(customContext)

	suite.WaitForJobsDone(1)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var webhookRecord dao.LedgerTransactionEventDao
	err = suite.TestDB.Model(&dao.LedgerTransactionEventDao{}).Where("user_id = ?", sessionId).First(&webhookRecord).Error
	suite.Require().NoError(err, "Failed to fetch updated user data")

	suite.Require().Equal(webhookRecord.UserId, sessionId, "UserId was saved")
	suite.Require().NotEmpty(webhookRecord.AccountNumber, "Account number is present on record")
	suite.Require().NotEmpty(webhookRecord.AccountRoutingNumber, "Account number is present on record")
	suite.Require().NotEmpty(webhookRecord.Channel, "Channel is present on record")
	suite.Require().NotEmpty(webhookRecord.EventId, "Account number is present on record")
	suite.Require().Equal(webhookRecord.TransactionType, "ACH_OUT", "Transaction type is ACH OUT")
	suite.Require().Equal(webhookRecord.InstructedCurrency, "USD", "Transaction currency is save as USD")
	suite.Require().NotEmpty(webhookRecord.InstructedAmount, "Transaction amount is present on record")
	suite.Require().True(webhookRecord.IsOutward, "Record indicates outgoing funds")
	suite.Require().NotEmpty(webhookRecord.ExternalBankAccountNumber, "Bank account number is present")
	suite.Require().NotEmpty(webhookRecord.ExternalBankAccountName, "Bank account name is present")
	suite.Require().NotEmpty(webhookRecord.ExternalBankAccountRoutingNumber, "Bank account routing number is present")
	suite.Require().NotEmpty(webhookRecord.RawPayload, "Raw payload is present for ach push")
}

func (suite *IntegrationTestSuite) TestWebhookHandlerWithDebitAuthorization() {
	defer SetupMockForLedger(suite).Close()
	mockSardineServer := setupMockForSardine(suite)
	defer mockSardineServer.Close()

	config.Config.Sardine.ApiBase = mockSardineServer.URL

	h := suite.newHandler()

	publicKey, privateKey, err := crypto.CreateKeys()
	suite.Require().NoError(err, "Failed to generate keys for signing")

	config.Config.Webhook.PublicKey = publicKey

	sessionId := suite.SetupTestData()
	userPublicKey := suite.createUserPublicKeyRecord(sessionId)

	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH000006009011111",
		CardId:        "6f586be7bf1c44b8b4ea11b2e2510e25",
		UserId:        sessionId,
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
	}
	err = suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	rawPayload := map[string]interface{}{
		"avalBalance": 2810,
		"channel":     "VISA_DPS",
		"credit":      false,
		"creditorAccount": map[string]interface{}{
			"accountNumber":       "500400036485912",
			"customerAccountName": "Default DreamFi Account",
			"customerAccountType": "CHECKING",
			"holderId":            "1802af7fa0a9318f108376772b394e91",
			"holderIdType":        "SSN",
			"holderName":          "Test User",
			"institutionId":       "124303298",
		},
		"customerID": "100000000033024",
		"debtorAccount": map[string]interface{}{
			"accountNumber":       "123456789012345",
			"customerAccountName": "Default DreamFi Account",
			"customerAccountType": "CHECKING",
			"holderId":            "1802af7fa0a9318f108376772b394e91",
			"holderIdType":        "SSN",
			"holderName":          "Test User",
			"institutionId":       "124303298",
		},
		"instructedAmount":   880,
		"instructedCurrency": "USD",
		"ledgerTxnNumber":    "QA00000000027060",
		"mcc":                "5965",
		"postedDate":         "2025-07-22T18:42:00.57Z",
		"referenceID":        "8ab0ef02261f4aae946df50e3d1c564b",
		"source":             "VISA_DPS",
		"status":             "PENDING",
		"subTransactionType": "PRE_AUTH_DOM",
		"transactionNumber":  "QA00000000027060",
		"transactionType":    "PRE_AUTH",
	}

	payloadBytes, err := json.Marshal(rawPayload)
	suite.Require().NoError(err, "Failed to marshal raw payload")

	signature, err := crypto.SignECDSA(payloadBytes, privateKey)
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	ledgerEventPayload := request.LedgerEventPayload{
		Source:    "PL",
		EventId:   "EVT45418",
		EventName: "Transaction.NEW",
		Signature: signature,
		Payload:   payloadBytes,
	}
	requestBody, err := json.Marshal(ledgerEventPayload)
	suite.Require().NoError(err, "Failed to marshall request")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/ledger/events", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/ledger/events")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, userPublicKey.PublicKey, c)

	err = h.LedgerWebhookHandler(customContext)

	suite.WaitForJobsDone(1)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var webhookRecord dao.LedgerTransactionEventDao
	err = suite.TestDB.Model(&dao.LedgerTransactionEventDao{}).Where("user_id = ?", sessionId).First(&webhookRecord).Error
	suite.Require().NoError(err, "Failed to fetch updated user data")

	suite.Require().Equal(webhookRecord.UserId, sessionId, "UserId was saved")
	suite.Require().NotEmpty(webhookRecord.AccountNumber, "Account number is present on record")
	suite.Require().NotEmpty(webhookRecord.AccountRoutingNumber, "Account number is present on record")
	suite.Require().NotEmpty(webhookRecord.Channel, "Channel is present on record")
	suite.Require().NotEmpty(webhookRecord.EventId, "Account number is present on record")
	suite.Require().Equal(webhookRecord.TransactionType, "PRE_AUTH", "Transaction type is PRE AUTH")
	suite.Require().Equal(webhookRecord.InstructedCurrency, "USD", "Transaction currency is save as USD")
	suite.Require().NotEmpty(webhookRecord.InstructedAmount, "Transaction amount is present on record")
	suite.Require().True(webhookRecord.IsOutward, "Record indicates outgoing funds")
	suite.Require().Empty(webhookRecord.ExternalBankAccountNumber, "Bank account number is not present for debit auth")
	suite.Require().NotEmpty(webhookRecord.Mcc, "Mcc is present for debit auth")
	suite.Require().NotEmpty(webhookRecord.RawPayload, "Raw payload is present for debit auth")
}

func (suite *IntegrationTestSuite) TestStatementEmailNotificationWebhook() {
	defer SetupMockForLedger(suite).Close()

	suite.configEmail()

	h := suite.newHandler()

	publicKey, privateKey, err := crypto.CreateKeys()
	suite.Require().NoError(err, "Failed to generate keys for signing")

	config.Config.Webhook.PublicKey = publicKey

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)

	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH0000060090",
		CardId:        "fcf2f39199174939fe437",
		AccountId:     "8146030",
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
		UserId:        userRecord.Id,
	}

	err = suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	rawPayload := map[string]interface{}{
		"accountIDs": []string{"8146030"},
		"batch_id":   "5fc4d81b3bc659f293c41c69782a8df7",
		"pageNumber": 1,
		"timestamp":  "2025-10-07T13:34:06.550211131Z",
		"totalPages": 1,
	}

	payloadBytes, err := json.Marshal(rawPayload)
	suite.Require().NoError(err, "Failed to marshal raw payload")

	signature, err := crypto.SignECDSA(payloadBytes, privateKey)

	suite.Require().NoError(err, "Failed to sign payload for new statements event payload")

	ledgerStatementNotificationPayload := request.LedgerEventPayload{
		Source:    "PL",
		EventId:   "EVT45418",
		EventName: "MONTHLY STATEMENT GENERATION",
		Signature: signature,
		Payload:   payloadBytes,
	}

	requestBody, err := json.Marshal(ledgerStatementNotificationPayload)
	suite.Require().NoError(err, "Failed to marshall request")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/ledger/events", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/ledger/events")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, userPublicKey.PublicKey, c)

	err = h.LedgerWebhookHandler(customContext)

	suite.WaitForJobsDone(2)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")
}
