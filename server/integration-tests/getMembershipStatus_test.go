package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestGetMembershipStatus() {
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	suite.createMembershipRecord(userRecord.Id)
	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/account/membership/status", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	customContext := security.GenerateLoggedInRegisteredUserContext(userRecord.Id, userPublicKey.PublicKey, c)

	err := handler.GetMemberShipStatus(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.GetMemberShipStatus
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().NotEmpty(responseBody.Status, "Status should not be empty")
	suite.Require().Equal(responseBody.Status, "subscribed", "Mebership status must be subscribed")
}
