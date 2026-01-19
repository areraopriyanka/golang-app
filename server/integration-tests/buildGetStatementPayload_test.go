package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"
)

func (suite *IntegrationTestSuite) TestBuildGetStatementPayload() {
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	config.Config.Aws.KmsEncryptionKeyId = "test-kms-encryption-key-id"

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

	e := handler.NewEcho()

	request := request.BuildGetStatementPayloadRequest{
		StatementId: "35123",
	}

	requestBody, err := json.Marshal(request)
	suite.Require().NoError(err, "Failed to marshall request")

	req := httptest.NewRequest(http.MethodPost, "/get-statement/build", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, "examplePublicKey", c)

	err = handler.BuildGetStatementPayload(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var responseBody response.BuildPayloadResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")
	suite.Require().NotEmpty(responseBody.PayloadId, "PayloadId should not be empty")

	var payloadRecord dao.SignablePayloadDao
	err = suite.TestDB.Model(&dao.SignablePayloadDao{}).Where("id = ?", responseBody.PayloadId).First(&payloadRecord).Error
	suite.Require().NoError(err, "Failed to fetch updated payload data")

	suite.Require().NotEmpty(payloadRecord.Payload, "Payload should not be empty")
}
