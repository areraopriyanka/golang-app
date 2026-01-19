package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/ledger"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestTransactionAchPush() {
	defer SetupMockForLedger(suite).Close()
	h := suite.newHandler()

	config.Config.Aws.KmsEncryptionKeyId = "test-kms-encryption-key-id"
	encryptedPassword, err := utils.EncryptKmsBinary("@8Kf0exhwDN6$sx@$3nazrABuaVBQxsI")
	suite.Require().NoError(err, "Failed to encrypt example ledgerPassword")

	sessionId := uuid.New().String()
	userRecord := dao.MasterUserRecordDao{
		Id:                         sessionId,
		FirstName:                  "DEBTOR",
		LastName:                   "Bar",
		Email:                      "email@example.com",
		KmsEncryptedLedgerPassword: []byte(encryptedPassword),
		LedgerCustomerNumber:       "100000000006001",
		Password:                   []byte("password"),
		UserStatus:                 constant.ACTIVE,
	}

	err = suite.TestDB.Select("id", "first_name", "last_name", "email", "kms_encrypted_ledger_password", "ledger_customer_number", "password", "user_status").Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	reason := "Settlements"
	payloadData := ledger.BuildOutboundAchCreditRequest(
		&userRecord,
		"50040002699049",
		"000000000",
		"50000",
		"CREDITOR",
		"234567890123456",
		"012345678",
		"CHECKING",
		&reason,
		nil,
	)

	payload, err := dao.CreateSignablePayloadForUser(userRecord.Id, payloadData)
	suite.Require().Nil(err, "Failed to create signable payload for user")

	transactionRequest := handler.TransactionAchPushRequest{
		Signature: "example_signature",
		PayloadId: payload.PayloadId,
	}
	requestBody, err := json.Marshal(transactionRequest)
	suite.Require().NoError(err, "Failed to marshall transaction request")

	encryptedApiKey, err := utils.EncryptKmsBinary("c077ad8b3d6f40c9896f5fb475f738d6")
	suite.Require().NoError(err, "Failed to encrypt example apiKey")
	suite.T().Logf("Encrypted API Key: %v", encryptedApiKey)

	userPublicKey := dao.UserPublicKey{
		UserId:             sessionId,
		KmsEncryptedApiKey: []byte(encryptedApiKey),
		KeyId:              "exampleKeyId",
		PublicKey:          "examplePublicKey",
	}

	err = suite.TestDB.Create(&userPublicKey).Error
	suite.Require().NoError(err, "Failed to insert user public key record")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/ach/push", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/ach/push")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "examplePublicKey", c)

	err = h.TransactionAchPush(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody handler.TransactionAchPushResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().EqualValues(500_00, responseBody.Amount, "Expected Amount to match")
}
