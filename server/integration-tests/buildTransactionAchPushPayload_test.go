package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model"
	"process-api/pkg/model/response"
	"process-api/pkg/plaid"
	"process-api/pkg/security"
)

type buildTransactionAchPushContext struct {
	handler           *handler.Handler
	user              dao.MasterUserRecordDao
	token             string
	externalAccountID string
}

func (suite *IntegrationTestSuite) TestBuildTransactionAchPushPayload() {
	defer SetupMockForLedger(suite).Close()
	ctx := suite.newBuildTransactionAchPushContext()
	note := "Settlement"
	buildPayloadRequest := handler.BuildTransactionAchPushPayloadRequest{
		AmountCents:       model.TransferCents(5000),
		Note:              &note,
		DreamFiAccountID:  "217140",
		ExternalAccountID: ctx.externalAccountID,
	}

	suite.AssertSuccessBuildTransactionAchPushPayload(ctx, buildPayloadRequest)
}

func (suite *IntegrationTestSuite) TestBuildTransactionAchPushPayloadWithoutNote() {
	defer SetupMockForLedger(suite).Close()
	ctx := suite.newBuildTransactionAchPushContext()
	buildPayloadRequest := handler.BuildTransactionAchPushPayloadRequest{
		AmountCents:       model.TransferCents(5000),
		DreamFiAccountID:  "217140",
		ExternalAccountID: ctx.externalAccountID,
	}
	suite.AssertSuccessBuildTransactionAchPushPayload(ctx, buildPayloadRequest)
}

func (suite *IntegrationTestSuite) newBuildTransactionAchPushContext() buildTransactionAchPushContext {
	h := suite.newHandler()
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)
	_ = suite.createUserAccountCard(userRecord.Id)

	token, err := security.GenerateOnboardedJwt(userRecord.Id, userPublicKey.PublicKey, nil)
	suite.Require().NoError(err, "Failed to generate JWT token")

	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	// All hardcoded values are set from mockoon
	plaidItemId := "j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB4a0"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b0"
	err = ps.InsertItem(userRecord.Id, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userRecord.Id, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	plaidAccountID := "vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK980"
	var record dao.PlaidAccountDao
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_account_id=?", plaidAccountID).Find(&record).Error
	suite.Require().NoError(err)
	externalAccountID := record.ID
	return buildTransactionAchPushContext{
		handler:           h,
		user:              userRecord,
		token:             token,
		externalAccountID: externalAccountID,
	}
}

func (suite *IntegrationTestSuite) callBuildTransactionAchPush(ctx buildTransactionAchPushContext, buildPayloadRequest handler.BuildTransactionAchPushPayloadRequest) *httptest.ResponseRecorder {
	requestBody, err := json.Marshal(buildPayloadRequest)
	suite.Require().NoError(err, "could not marhsal payload")
	e := handler.NewEcho()
	mw := e.Group("", security.LoggedInRegisteredUserMiddleware)
	mw.POST("/accounts/ach/push/build", ctx.handler.BuildTransactionAchPushPayload)
	req := httptest.NewRequest(http.MethodPost, "/accounts/ach/push/build", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ctx.token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func (suite *IntegrationTestSuite) AssertSuccessBuildTransactionAchPushPayload(ctx buildTransactionAchPushContext, buildPayloadRequest handler.BuildTransactionAchPushPayloadRequest) {
	rec := suite.callBuildTransactionAchPush(ctx, buildPayloadRequest)
	suite.Require().Equal(http.StatusOK, rec.Code)

	var responseBody struct {
		PayloadId string `json:"payloadId"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")

	var payloadRecord dao.SignablePayloadDao
	err = suite.TestDB.Model(&dao.SignablePayloadDao{}).Where("id = ?", responseBody.PayloadId).First(&payloadRecord).Error
	suite.Require().NoError(err, "Failed to fetch updated user data")

	suite.Require().Equal(responseBody.PayloadId, payloadRecord.Id, "Payload ID should match")
	suite.Require().NotEmpty(payloadRecord.Payload, "Payload should not be empty")

	var payload ledger.OutboundAchCreditRequest
	err = json.Unmarshal([]byte(payloadRecord.Payload), &payload)
	suite.Require().NoError(err, "Failed to unmarshal payload")
	suite.Require().Equal("ACH", payload.Channel, "TransactionType should be ACH")
	suite.Require().Equal("ACH_OUT", payload.TransactionType, "TransactionType should be ACH_OUT")
	suite.Require().Equal("5000", payload.TransactionAmount.Amount, "Amount should match")
	suite.Require().Equal(ctx.user.FirstName, payload.Debtor.FirstName, "Creditor Name should match")
	userAccountCard, err := dao.UserAccountCardDao{}.FindOneActiveByUserId(suite.TestDB, ctx.user.Id)
	suite.Require().NotNil(userAccountCard)
	suite.Require().NoError(err)
	// "9900009606" and "011401533" come from mockoon; we do not store the debtor's account number in the middleware
	suite.Require().Equal("9900009606", payload.CreditorAccount.Identification, "Creditor Account number should match")
	suite.Require().Equal("011401533", payload.CreditorAccount.Institution.Identification, "Creditor routing number should match")
	suite.Require().Equal(userAccountCard.AccountNumber, payload.DebtorAccount.Identification, "Creditor Account Name should match")
}

func (suite *IntegrationTestSuite) TestBuildTransactionAchPushPayloadWithMaxAmount() {
	defer SetupMockForLedger(suite).Close()
	ctx := suite.newBuildTransactionAchPushContext()
	buildPayloadRequest := handler.BuildTransactionAchPushPayloadRequest{
		AmountCents:       model.TransferCents(10000001),
		DreamFiAccountID:  "217140",
		ExternalAccountID: "vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK980",
	}
	rec := suite.callBuildTransactionAchPush(ctx, buildPayloadRequest)
	suite.Require().Equal(http.StatusBadRequest, rec.Code, "AmountCents: Should be more than 0 dollar and less than 100000 dollars")
}

func (suite *IntegrationTestSuite) TestBuildTransactionAchPushPayloadNSF() {
	defer SetupMockForLedger(suite).Close()
	ctx := suite.newBuildTransactionAchPushContext()
	buildPayloadRequest := handler.BuildTransactionAchPushPayloadRequest{
		AmountCents:       model.TransferCents(100000),
		DreamFiAccountID:  "217140",
		ExternalAccountID: "vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK980",
	}
	rec := suite.callBuildTransactionAchPush(ctx, buildPayloadRequest)
	suite.Require().Equal(http.StatusUnprocessableEntity, rec.Code, "Should be a 422 error code")
	var errResponse response.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &errResponse)
	suite.Require().NoError(err, "Failed to parse response body")
	suite.Equal(constant.INSUFFICIENT_FUNDS, errResponse.ErrorCode, "Expected error code INSUFFICIENT_FUNDS")
}
