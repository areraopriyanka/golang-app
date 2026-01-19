package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/ledger"
	"process-api/pkg/security"
	"process-api/pkg/utils"
)

func (suite *IntegrationTestSuite) TestBuildCardPinPayload() {
	config.Config.Aws.KmsEncryptionKeyId = "test-kms-encryption-key-id"
	config.Config.Ledger.CardsPublicKey = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApSQgtJ7XXaiYhcTgjuhJ
VWJDdXykujwP4yRCpcEKQBTtdVPkHquPaeDhozCtATO/XoZFJ7d1WtR8Ia//8jZg
FxsFDL7hR8ISR+vVB3sb8cOJvRTCXlMo7OPygi/RpGW8stQaBOSpf6fu4dp8A5Ty
6agiQ1zptDMB2ZHcSbBM36L1ri2UssvPxqn97UTb/3dqwleh0kCITybx8ZspXvji
65YT8JdTPLNEOby+5K8JRlyGJyDmfpjB6SAenBOpOp1JpBlxN8L4+SGINEsYq5//
MAocOQ27Ae2/cV4007V59D1y+FHRrkT++K3freRxIUDmGrzmjfUmMDqB75oD+fYV
4QIDAQAB
-----END PUBLIC KEY-----`

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	encryptedApiKey, err := utils.EncryptKmsBinary("c077ad8b3d6f40c9896f5fb475f738d6")
	suite.Require().NoError(err, "Failed to encrypt example apiKey")

	userPublicKey := dao.UserPublicKey{
		UserId:             userRecord.Id,
		KmsEncryptedApiKey: []byte(encryptedApiKey),
		KeyId:              "exampleKeyId",
		PublicKey:          "examplePublicKey",
	}

	err = suite.TestDB.Create(&userPublicKey).Error
	suite.Require().NoError(err, "Failed to insert user public key record")

	userAccountCard := dao.UserAccountCardDao{
		CardId:        "exampleCardId",
		UserId:        userRecord.Id,
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
	}
	err = suite.TestDB.Select("card_id", "user_id", "account_number", "account_status").Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert test card")

	buildPinPayloadRequest := handler.BuildCardPinPayloadRequest{
		NewPin: "1234",
	}
	requestBody, _ := json.Marshal(buildPinPayloadRequest)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/pin/build", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("cards/pin/build")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, "examplePublicKey", c)

	err = handler.BuildCardPinPayload(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var responseBody struct {
		PayloadId string `json:"payloadId"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")

	var payloadRecord dao.SignablePayloadDao
	err = suite.TestDB.Model(&dao.SignablePayloadDao{}).Where("id = ?", responseBody.PayloadId).First(&payloadRecord).Error
	suite.Require().NoError(err, "Failed to fetch updated user data")

	suite.Require().Equal(responseBody.PayloadId, payloadRecord.Id, "Payload ID should match")
	suite.Require().NotEmpty(payloadRecord.Payload, "Payload should not be empty")

	var pinChangePayload ledger.ChangePinRequest
	err = json.Unmarshal([]byte(payloadRecord.Payload), &pinChangePayload)
	suite.Require().NoError(err, "Failed to unmarshal payload")
	suite.Require().Equal("PIN_CHANGE", pinChangePayload.TransactionType, "TransactionType should be PIN_CHANGE")
	suite.Require().Equal(userRecord.LedgerCustomerNumber, pinChangePayload.CustomerId, "CustomerID should match")
	suite.Require().Equal(userAccountCard.AccountNumber, pinChangePayload.AccountNumber, "AccountNumber should match")
	suite.Require().Equal("PL", pinChangePayload.Product, "Product should be DEFAULT")
	suite.Require().Equal("VISA_DPS", pinChangePayload.Channel, "Channel should be VISA_DPS")
	suite.Require().Equal("DREAMFI_MVP", pinChangePayload.Program, "Program should be DEFAULT")
}
