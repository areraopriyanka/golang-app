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
	"process-api/pkg/security"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestGetCardLimit() {
	defer SetupMockForLedger(suite).Close()
	user := suite.createUser()

	cardHolder := dao.UserAccountCardDao{
		CardHolderId:  "CH000006009011111",
		CardId:        "6f586be7bf1c44b8b4ea11b2e2510e25",
		UserId:        user.Id,
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
	}
	err := suite.TestDB.Create(&cardHolder).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	payloadId := suite.createGetCardLimitSignablePayload(user.LedgerCustomerNumber, cardHolder.AccountNumber, user.Id)
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)

	getCardLimitRequest := request.LedgerApiRequest{
		Signature: "test_signature",
		Mfp:       "test_mfp",
		PayloadId: payloadId,
	}
	requestBody, _ := json.Marshal(getCardLimitRequest)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/dashboard/card-limit", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/dashboard/card-limit")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)

	err = handler.GetCardLimit(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody ledger.GetCardLimitResult
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().NotEmpty(responseBody.Card, "Expected card data")
}

func (suite *IntegrationTestSuite) createGetCardLimitSignablePayload(customerNo, accountNumber, useId string) string {
	apiConfig := ledger.NewNetXDCardApiConfig(config.Config.Ledger)
	payloadData := apiConfig.BuildGetCardLimitRequest(customerNo, accountNumber, "6f586be7bf1c44b8b4ea11b2e2510e25")
	jsonPayloadBytes, err := json.Marshal(payloadData)
	suite.Require().NoError(err, "Failed to marshal test payload")

	jsonPayload := string(jsonPayloadBytes)

	payloadId := uuid.New().String()
	payloadRecord := dao.SignablePayloadDao{
		Id:      payloadId,
		Payload: jsonPayload,
		UserId:  &useId,
	}
	err = suite.TestDB.Create(&payloadRecord).Error
	suite.Require().NoError(err, "Failed to insert test payload")
	return payloadId
}
