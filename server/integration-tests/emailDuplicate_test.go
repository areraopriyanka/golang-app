package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/constant"
	"process-api/pkg/handler"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
)

func (suite *IntegrationTestSuite) TestEmailDuplicateHandlerWithDatabaseDuplicate() {
	userStatus := constant.USER_CREATED
	_ = suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	// hardcoded match with createTestUser
	email := "testuser@gmail.com"
	emailResponse := requestEmailDuplicate(email, suite)

	suite.Equal(true, emailResponse.IsEmailDuplicate, "should find duplicate email")
}

func (suite *IntegrationTestSuite) TestEmailDuplicateHandlerWithLedgerDuplicate() {
	// hardcoded match with SetupMockForLedger
	email := "existing-ledger@gmail.com"
	emailResponse := requestEmailDuplicate(email, suite)

	suite.Equal(true, emailResponse.IsEmailDuplicate, "should find duplicate email")
}

func (suite *IntegrationTestSuite) TestEmailDuplicateHandlerWithoutDuplicate() {
	userStatus := constant.USER_CREATED
	_ = suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	email := "foo@baz.com"
	emailResponse := requestEmailDuplicate(email, suite)

	suite.Equal(false, emailResponse.IsEmailDuplicate, "should not find duplicate email")
}

func requestEmailDuplicate(email string, suite *IntegrationTestSuite) response.EmailDuplicateResponse {
	defer SetupMockForLedger(suite).Close()

	updateRequest := request.EmailDuplicateRequest{
		Email: email,
	}
	requestBody, err := json.Marshal(updateRequest)
	suite.Require().NoError(err, "Failed to marshall request")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/onboarding/emailDuplicate", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/onboarding/emailDuplicate")

	err = handler.IsEmailDuplicate(c)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var emailResponse response.EmailDuplicateResponse
	err = json.Unmarshal(rec.Body.Bytes(), &emailResponse)
	suite.Require().NoError(err, "Failed to unmarshal login response")
	return emailResponse
}
