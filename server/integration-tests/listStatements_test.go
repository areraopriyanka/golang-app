package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/ledger"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestListStatements() {
	defer SetupMockForLedger(suite).Close()

	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})

	payload := ledger.ListStatementRequest{
		PageNumber: 1,
		PageSize:   10,
	}

	jsonPayloadBytes, err := json.Marshal(payload)
	suite.Require().NoError(err, "Failed to marshal test payload")

	jsonPayload := string(jsonPayloadBytes)

	payloadId := uuid.New().String()
	payloadRecord := dao.SignablePayloadDao{
		Id:      payloadId,
		Payload: jsonPayload,
		UserId:  &userRecord.Id,
	}
	err = suite.TestDB.Create(&payloadRecord).Error
	suite.Require().NoError(err, "Failed to insert test payload")

	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)

	listStatementsRequest := handler.ListStatementsRequest{
		LedgerApiRequest: request.LedgerApiRequest{
			Signature: "test_signature",
			Mfp:       "test_mfp",
			PayloadId: payloadId,
		},
		AccountId: "21281104",
	}
	requestBody, _ := json.Marshal(listStatementsRequest)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/list-statements", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, userPublicKey.PublicKey, c)

	err = handler.ListStatements(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.ListStatementResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().NotEmpty(responseBody.Statements)
	suite.Require().Equal(2, len(responseBody.Statements))
}
