package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestBuildValidateCvvPayload() {
	defer SetupMockForLedger(suite).Close()
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	// Create userAccountCard Record
	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH000006009011111",
		CardId:        "6f586be7bf1c44b8b4ea11b2e2510e25",
		AccountNumber: "190098989898",
		AccountStatus: "ACTIVE",
		UserId:        userRecord.Id,
	}

	err := suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")
	e := handler.NewEcho()

	request := request.BuildValidateCvvRequest{
		CVV: "006",
	}
	requestBody, err := json.Marshal(request)
	suite.Require().NoError(err, "Failed to marshall request")
	req := httptest.NewRequest(http.MethodGet, "/cards/validate-cvv/build", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/validate-cvv/build")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, "publicKey", c)

	err = handler.BuildValidateCvvPayload(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var responseBody response.BuildPayloadResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")
	suite.Require().NotEmpty(responseBody.PayloadId, "PayloadId should not be empty")

	// Check if payload data inserted in DB
	var payloadRecord dao.SignablePayloadDao
	err = suite.TestDB.Model(&dao.SignablePayloadDao{}).Where("id = ?", responseBody.PayloadId).First(&payloadRecord).Error
	suite.Require().NoError(err, "Failed to fetch updated payload data")

	suite.Require().NotEmpty(payloadRecord.Payload, "Payload should not be empty")
}
