package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestBuildActivateCardPayload() {
	defer SetupMockForLedger(suite).Close()
	// Create user record
	sessionId := suite.SetupTestData()
	e := handler.NewEcho()

	request := handler.BuildActivateCardRequest{
		CVV: "006",
		Pin: "1234",
	}
	requestBody, err := json.Marshal(request)
	suite.Require().NoError(err, "Failed to marshall request")

	req := httptest.NewRequest(http.MethodPost, "/cards/activate/build", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/activate/build")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "publicKey", c)

	err = handler.BuildActivateCardPayload(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var responseBody handler.BuildActivateCardResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")
	suite.Require().NotEmpty(responseBody.ValidateCvv.PayloadId, "ValidateCvv.PayloadId should not be empty")
	suite.Require().NotEmpty(responseBody.ChangePin.PayloadId, "SetCardPin.PayloadId should not be empty")
	suite.Require().NotEmpty(responseBody.UpdateStatus.PayloadId, "UpdateStatus.PayloadId should not be empty")

	// Check if payload data inserted in DB
	var validateCvvPayloadRecord dao.SignablePayloadDao
	err = suite.TestDB.Model(&dao.SignablePayloadDao{}).Where("id = ?", responseBody.ValidateCvv.PayloadId).First(&validateCvvPayloadRecord).Error
	suite.Require().NoError(err, "Failed to fetch ValidateCvv payload data")
	suite.Require().NotEmpty(validateCvvPayloadRecord.Payload, "ValidateCvv should not be empty")

	var changePinPayloadRecord dao.SignablePayloadDao
	err = suite.TestDB.Model(&dao.SignablePayloadDao{}).Where("id = ?", responseBody.ChangePin.PayloadId).First(&changePinPayloadRecord).Error
	suite.Require().NoError(err, "Failed to fetch SetCardPin payload data")
	suite.Require().NotEmpty(changePinPayloadRecord.Payload, "SetCardPin should not be empty")

	var updateStatusPayloadRecord dao.SignablePayloadDao
	err = suite.TestDB.Model(&dao.SignablePayloadDao{}).Where("id = ?", responseBody.UpdateStatus.PayloadId).First(&updateStatusPayloadRecord).Error
	suite.Require().NoError(err, "Failed to fetch UpdateStatus payload data")
	suite.Require().NotEmpty(updateStatusPayloadRecord.Payload, "UpdateStatus should not be empty")
}
