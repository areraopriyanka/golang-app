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
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestValidateCvv() {
	defer SetupMockForLedger(suite).Close()
	user := suite.createUser()
	encryptedCvv := "fbbVg8NQFDwa5rtkjR+Wa4wqiZwUWoxaUVLO2fTFgeYvQ9RGBfyl6H0rpCvtEoEbhbVFTuJGX8hdgjfptOynnrKD5DUFpK0RDi+URuKIunzbiQ9Gq6iqs54S2LaWH0ZVwsvPb4HiP1mOX88oV7xZc7rzWcG1nOBErogItPTijlG/UMvVWU0SYMd+fjBNqMcdMeFe3GLch3KMFpAC4yC+sTSyLj9W8C0vOhF2bA3voxX2Vb5Xv27oR8Xn6jKoR+/ZzK0Ch3aF06P+Xxto/ZGZyvdxsN+s2ldEu+vHbK6QwXC30/U+J8TFOQ0rIbZny3LOtaBK0sbflAdLNlQwaFmCfQ=="
	payloadId := suite.createValidateCvvSignablePayload(encryptedCvv, user.Id, false)
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)

	validateCvvRequest := request.LedgerApiRequest{
		Signature: "test_signature",
		Mfp:       "test_mfp",
		PayloadId: payloadId,
	}
	requestBody, _ := json.Marshal(validateCvvRequest)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/validate-cvv", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/validate-cvv")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)

	err := handler.ValidateCvv(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.ValidateCvvResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().True(responseBody.IsValidCvv, "Expected IsValidCvv value")
}

func (suite *IntegrationTestSuite) createValidateCvvSignablePayload(Cvv string, userId string, isUnEncypted bool) string {
	userClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, nil)
	payloadData, err := userClient.BuildValidateCvvRequest("6f586be7bf1c44b8b4ea11b2e2510e25", Cvv, isUnEncypted)
	suite.Require().NoError(err, "Failed to generate validate cvv payload")
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
