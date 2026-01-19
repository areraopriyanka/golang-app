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

func (suite *IntegrationTestSuite) TestBuildReplaceCardPayload() {
	sessionId := suite.SetupTestData()

	replaceCardReason := request.ReplaceCardReason{
		Reason: "stolen",
	}

	requestBody, _ := json.Marshal(replaceCardReason)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/replace/build", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "publicKey", c)

	err := handler.BuildReplaceCardPayload(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var responseBody response.BuildPayloadResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")
	suite.Require().NotEmpty(responseBody.PayloadId, "PayloadId should not be empty")
	suite.Require().NotEmpty(responseBody.Payload, "Payload should not be empty")

	// Validate ReplaceCard payload insertion in DB
	suite.ValidatePayloadInDB(responseBody.PayloadId)
}

// setupTestData inserts test user and cardHolder records into the DB.
func (suite *IntegrationTestSuite) SetupTestData() string {
	user := suite.createTestUser(PartialMasterUserRecordDao{})

	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH0000060090",
		CardId:        "fcf2f39199174939fe437",
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
		UserId:        user.Id,
	}

	err := suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	return user.Id
}

func (suite *IntegrationTestSuite) ValidatePayloadInDB(payloadId string) {
	var payloadRecord dao.SignablePayloadDao
	err := suite.TestDB.Model(&dao.SignablePayloadDao{}).Where("id = ?", payloadId).First(&payloadRecord).Error
	suite.Require().NoError(err, "Failed to fetch updated payload data")
	suite.Require().NotEmpty(payloadRecord.Payload, "Payload should not be empty")
}
