package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/sardine"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"github.com/google/uuid"
)

func setupMockForSardine(suite *IntegrationTestSuite) *httptest.Server {
	// Sets up mock sardine to specify response to sardine API
	mockLedgerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			suite.Require().NoError(err, "Failed to read request body")
		}

		var request sardine.PostCustomerInformationJSONRequestBody
		err = json.Unmarshal(payload, &request)
		if err != nil {
			suite.Require().NoError(err, "Failed to unmarshal request body")
		}

		if request.Customer.LastName == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPreconditionFailed)
			_, err = w.Write([]byte("{}"))
			suite.Require().NoError(err, "Failed to write empty response body")
			return
		}

		var responseBody string
		switch *request.Customer.LastName {
		case "lowsardine":
			// set level to low
			responseBody = `{"sessionKey":"11f3780f-d07b-45d4-8d4a-30b7ef86436b","level":"low","status":"Success","customer":{"score":0,"level":"low","signals":[{"key":"addressRiskLevel","value":"low","reasonCodes":["AL4","AL5","AL9"]},{"key":"emailDomainLevel","value":"low"},{"key":"emailLevel","value":"low"},{"key":"pepLevel","value":"low"},{"key":"phoneLevel","value":"low"},{"key":"sanctionLevel","value":"low"}],"address":{"validity":"invalid"}},"checkpoints":{"customer":{"riskLevel":{"value":"low","ruleIds":null}},"onboarding":{"riskLevel":{"value":"low","ruleIds":null}}},"rules":[{"id":957019,"isLive":false,"isAllowlisted":false,"name":"allow list JohnAAA SmithZZZ"}],"checkpointData":[{"name":"customer","type":"weighted_max"},{"name":"onboarding","type":"weighted_max"}]}`
		case "highsardine":
			// set level to high
			responseBody = `{"sessionKey":"a335e8ab-492c-47d3-873b-9a25bf364483","level":"high","status":"Success","customer":{"score":0,"level":"high","signals":[{"key":"addressRiskLevel","value":"low","reasonCodes":["AL4","AL5","AL9"]},{"key":"emailDomainLevel","value":"low"},{"key":"emailLevel","value":"low"},{"key":"kycLevel","value":"high"},{"key":"pepLevel","value":"low"},{"key":"phoneLevel","value":"low"},{"key":"sanctionLevel","value":"low"}],"address":{"validity":"invalid"}},"checkpoints":{"customer":{"kycLevel":{"value":"high","ruleIds":[957019]},"riskLevel":{"value":"high","ruleIds":[957019]}},"onboarding":{"riskLevel":{"value":"low","ruleIds":null}}},"rules":[{"id":957019,"isLive":true,"isAllowlisted":false,"name":"allow list JohnAAA SmithZZZ"}],"checkpointData":[{"name":"customer","type":"weighted_max"},{"name":"onboarding","type":"weighted_max"}]}`
		}
		var responseData map[string]interface{}
		err = json.Unmarshal([]byte(responseBody), &responseData)
		if err != nil {
			suite.Require().NoError(err, "Failed to unmarshal response body")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(responseData)
		if err != nil {
			suite.Require().NoError(err, "Response write should not return an error")
		}
	}))

	return mockLedgerServer
}

func (suite *IntegrationTestSuite) runKycTest(lastName, expectedKycStatus string) dao.MasterUserRecordDao {
	e := handler.NewEcho()
	h := suite.newHandler()
	mockSardineServer := setupMockForSardine(suite)
	defer mockSardineServer.Close()

	mockLedgerServer := SetupMockForLedger(suite)
	defer mockLedgerServer.Close()

	// Use the mock server URL instead of the actual Sardine API URL
	config.Config.Sardine.ApiBase = mockSardineServer.URL
	// Use mock ledger server
	config.Config.Ledger.Endpoint = mockLedgerServer.URL + "/jsonrpc"

	userStatus := constant.ADDRESS_CONFIRMED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{LastName: &lastName, UserStatus: &userStatus})

	request := request.SardineKycRequest{
		SSN:               "607789891",
		SardineSessionKey: "11f3780f-d07b-45d4-8d4a-30b7ef86436b",
	}
	requestBody, _ := json.Marshal(request)

	req := httptest.NewRequest(http.MethodPut, "/onboarding/customer/kyc", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err := h.SubmitUserDataToSardine(customContext)

	suite.Require().NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code)

	var responseBody response.MaybeRiverJobResponse[response.SardinePayload]
	var user dao.MasterUserRecordDao
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")
	suite.Require().NotEmpty(responseBody.JobId, "JobId should not be empty")
	suite.Require().NotEmpty(responseBody.State, "State should not be empty")
	suite.Require().NotNil(responseBody.Payload, "Payload should not be empty")
	suite.Require().NotNil(responseBody.Payload.Result, "Payload result should not be empty")
	suite.Require().NotEmpty(responseBody.Payload.Result.UserId, "UserId should not be empty")
	suite.Require().NotEmpty(responseBody.Payload.Result.KycStatus, "KycStatus should not be empty")
	suite.Require().Empty(responseBody.Payload.Error, "Error should be empty")
	suite.Require().Equal(expectedKycStatus, responseBody.Payload.Result.KycStatus)

	result := suite.TestDB.Where("id=?", testUser.Id).Find(&user)
	suite.Require().NoError(result.Error, "failed to fetch user record")
	return user
}

func (suite *IntegrationTestSuite) TestKycPassForInValidUserState() {
	e := handler.NewEcho()
	h := suite.newHandler()
	userStatus := constant.AGE_VERIFICATION_PASSED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	mockSardineServer := setupMockForSardine(suite)
	defer mockSardineServer.Close()

	mockLedgerServer := SetupMockForLedger(suite)
	defer mockLedgerServer.Close()

	// Use the mock server URL instead of the actual Sardine API URL
	config.Config.Sardine.ApiBase = mockSardineServer.URL
	// Use mock ledger server
	config.Config.Ledger.Endpoint = mockLedgerServer.URL + "/jsonrpc"

	request := request.SardineKycRequest{
		SSN:               "607789891",
		SardineSessionKey: "11f3780f-d07b-45d4-8d4a-30b7ef86436b",
	}
	requestBody, _ := json.Marshal(request)

	req := httptest.NewRequest(http.MethodPut, "/onboarding/customer/kyc", bytes.NewReader(requestBody))

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err := h.SubmitUserDataToSardine(customContext)

	suite.Require().NotNil(err, "Handler should return an error")
	errResp := err.(*response.ErrorResponse)
	suite.Equal(http.StatusPreconditionFailed, errResp.StatusCode, "Expected status code 412 StatusPreconditionFailed")
	suite.Equal(constant.INVALID_USER_STATE, errResp.ErrorCode, "Expected error code INVALID_USER_STATE")
}

func (suite *IntegrationTestSuite) TestGetSardineJobStatus() {
	h := suite.newHandler()
	e := handler.NewEcho()

	lastName := "lowsardine"
	userStatus := constant.ADDRESS_CONFIRMED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{LastName: &lastName, UserStatus: &userStatus})

	mockSardineServer := setupMockForSardine(suite)
	defer mockSardineServer.Close()

	mockLedgerServer := SetupMockForLedger(suite)
	defer mockLedgerServer.Close()

	// Use the mock server URL instead of the actual Sardine API URL
	config.Config.Sardine.ApiBase = mockSardineServer.URL
	// Use mock ledger server
	config.Config.Ledger.Endpoint = mockLedgerServer.URL + "/jsonrpc"
	encryptedSSN, err := utils.EncryptKms("607789891")
	suite.Require().NoError(err, "Failed to encrypt ssn")

	ctx := context.Background()

	sessionKey := uuid.New().String()
	jobInsertResult, err := h.RiverClient.Insert(ctx, handler.SardineJobArgs{
		UserId:     testUser.Id,
		SessionKey: sessionKey,
		SSN:        encryptedSSN,
	}, nil)

	suite.Require().NoError(err, "Failed to insert sort job")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/onboarding/customer/kyc/%d", jobInsertResult.Job.ID), nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/onboarding/customer/kyc/:jobId")
	c.SetParamNames("jobId")
	c.SetParamValues(fmt.Sprintf("%d", jobInsertResult.Job.ID))

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err = h.GetSardineJobStatus(customContext)

	suite.WaitForJobsDone(1)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var statusResponseBody response.MaybeRiverJobResponse[response.SardinePayload]

	err = json.Unmarshal(rec.Body.Bytes(), &statusResponseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	allowedStates := []string{
		"available", "cancelled", "completed", "discarded", "pending", "retryable", "running", "scheduled",
	}
	suite.Contains(allowedStates, statusResponseBody.State, "Job state should be of one of River's allowed states.")
}

func (suite *IntegrationTestSuite) TestKycPass() {
	user := suite.runKycTest("lowsardine", constant.KYC_PASS)
	suite.Require().Equal("KYC_PASS", user.UserStatus, "user's status should be updated to KYC_PASS")
	// If kyc is passed then check ledgerCustomerNumber is updated
	suite.NotEmpty(user.LedgerCustomerNumber, "ledgerCustomerNumber should be updated")
}

func (suite *IntegrationTestSuite) TestKycFail() {
	user := suite.runKycTest("highsardine", constant.KYC_FAIL)
	suite.Require().Equal("KYC_FAIL", user.UserStatus, "user's status should be updated to KYC_FAIL")
}
