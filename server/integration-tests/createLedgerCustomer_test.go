package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/clock"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"time"
)

func (suite *IntegrationTestSuite) TestCompleteLedgerCustomerForValidState() {
	unfreeze := clock.FreezeNow()
	defer unfreeze()
	defer SetupMockForLedger(suite).Close()

	userStatus := constant.CARD_AGREEMENTS_REVIEWED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	createLedgerCustomerRequest := handler.CompleteLedgerCustomerRequest{
		PublicKey:         `-----BEGIN PUBLIC KEY----- MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEpv0cLMRQ94fqkpU2uIdy+bjqhkewFDm8WP4rfdO/mAflz/UDjyEKp5FxfQlvZuQgbTyEIK/YBgU2AvbgEbNp6A== -----END PUBLIC KEY-----`,
		SardineSessionKey: `00000000-0000-0000-00000000000000000`,
	}
	requestBody, _ := json.Marshal(createLedgerCustomerRequest)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPut, "/onboarding/customer/complete-ledger-customer", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err := handler.CompleteLedgerCustomer(customContext)
	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusCreated, rec.Code, "Expected status code 201")

	var responseBody handler.CompleteLedgerCustomerResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	var userRecord dao.MasterUserRecordDao
	result := suite.TestDB.Where("id=?", testUser.Id).Find(&userRecord)
	suite.Require().NoError(result.Error, "failed to re-query user record")
	suite.Equal("ACTIVE", userRecord.UserStatus, "status should be ACTIVE if successful")

	// Validate the token received in the response header
	tokenWithBearer := rec.Header().Get("Authorization")
	suite.NotEmpty(tokenWithBearer, "Authorization token should be present in response headers")
	claims := security.GetClaimsFromToken(tokenWithBearer)
	suite.Equal(claims.Subject, testUser.Id, "JWT claim userId should match userId from user record")
	suite.Equal(claims.UserState, constant.ACTIVE, "Token does not belongs to ACTIVE user")
	suite.WithinDuration(clock.Now().Add(30*time.Minute), claims.ExpiresAt.Time, 0, "Expiration should match expectation")

	var membershipRecord dao.UserMembershipDao
	result = suite.TestDB.Where("user_id=?", testUser.Id).First(&membershipRecord)
	suite.Require().NoError(result.Error, "failed to re-query user membership record")
	suite.Equal("subscribed", membershipRecord.MembershipStatus, "membership status should be subscribed")
}

func (suite *IntegrationTestSuite) TestCompleteLedgerCustomerWithoutPublicKey() {
	unfreeze := clock.FreezeNow()
	defer unfreeze()
	defer SetupMockForLedger(suite).Close()

	userStatus := constant.CARD_AGREEMENTS_REVIEWED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	createLedgerCustomerRequest := handler.CompleteLedgerCustomerRequest{
		PublicKey:         `-----BEGIN PUBLIC KEY----- MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEpv0cLMRQ94fqkpU2uIdy+bjqhkewFDm8WP4rfdO/mAflz/UDjyEKp5FxfQlvZuQgbTyEIK/YBgU2AvbgEbNp6A== -----END PUBLIC KEY-----`,
		SardineSessionKey: `00000000-0000-0000-00000000000000000`,
	}
	requestBody, _ := json.Marshal(createLedgerCustomerRequest)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPut, "/onboarding/customer/complete-ledger-customer", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err := handler.CompleteLedgerCustomer(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusCreated, rec.Code, "Expected status code 201")

	var responseBody handler.CompleteLedgerCustomerResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal("CH00000000009004", responseBody.CardHolderId)
	suite.Require().Equal("d3360652aec34493976fa0d24b9d098d", responseBody.CardId)

	// Validate the token received in the response header
	tokenWithBearer := rec.Header().Get("Authorization")
	suite.NotEmpty(tokenWithBearer, "Authorization token should be present in response headers")
	claims := security.GetClaimsFromToken(tokenWithBearer)
	suite.Equal(claims.Subject, testUser.Id, "JWT claim userId should match userId from user record")
	suite.WithinDuration(clock.Now().Add(30*time.Minute), claims.ExpiresAt.Time, 0, "Expiration should match expectation")
	suite.Equal(claims.UserState, constant.ACTIVE, "Token does not belongs to ACTIVE user")

	var userRecord dao.MasterUserRecordDao
	result := suite.TestDB.Where("id=?", testUser.Id).Find(&userRecord)
	suite.Require().NoError(result.Error, "failed to re-query user record")

	suite.Equal("ACTIVE", userRecord.UserStatus, "status should be ACTIVE if successful")
}

func (suite *IntegrationTestSuite) TestCompleteLedgerCustomerForInvalidState() {
	defer SetupMockForLedger(suite).Close()

	userStatus := constant.USER_CREATED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	createLedgerCustomerRequest := handler.CompleteLedgerCustomerRequest{
		PublicKey:         `-----BEGIN PUBLIC KEY----- MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEpv0cLMRQ94fqkpU2uIdy+bjqhkewFDm8WP4rfdO/mAflz/UDjyEKp5FxfQlvZuQgbTyEIK/YBgU2AvbgEbNp6A== -----END PUBLIC KEY-----`,
		SardineSessionKey: `00000000-0000-0000-00000000000000000`,
	}
	requestBody, _ := json.Marshal(createLedgerCustomerRequest)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPut, "/onboarding/customer/complete-ledger-customer", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err := handler.CompleteLedgerCustomer(customContext)

	suite.Require().NotNil(err, "Handler should return an error for invalid user state")
	errResp := err.(*response.ErrorResponse)
	suite.Equal(http.StatusPreconditionFailed, errResp.StatusCode, "Expected status code 412 StatusPreconditionFailed")
	suite.Equal(constant.INVALID_USER_STATE, errResp.ErrorCode, "Expected error code INVALID_USER_STATE")
}
