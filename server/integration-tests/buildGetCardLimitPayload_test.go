package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestBuildGetCardLimitPayload() {
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	// Create cardHolderRecord
	cardHolder := dao.UserAccountCardDao{
		CardHolderId:  "CH00000600900090978",
		CardId:        "5434a20e0a9b4d1bbf8fea3f543f9a05",
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
		UserId:        userRecord.Id,
	}
	err := suite.TestDB.Create(&cardHolder).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")
	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/dashboard/card-limit/build", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/dashboard/card-limit/build")

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, "publicKey", c)

	err = handler.BuildGetCardLimitPayload(customContext)
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
