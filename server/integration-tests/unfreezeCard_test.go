package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestUnfreezeCard() {
	defer SetupMockForLedger(suite).Close()

	sessionId := suite.SetupTestData()

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/unfreeze", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/unfreeze")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "publicKey", c)

	err := handler.UnfreezeCard(customContext)
	suite.Require().NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.UpdateCardStatusResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")
	suite.Require().NotEmpty(responseBody.UpdatedCardStatus, "updatedCardStatus should not be empty")
	suite.Require().Equal("active", responseBody.UpdatedCardStatus, "Invalid card status")
}
