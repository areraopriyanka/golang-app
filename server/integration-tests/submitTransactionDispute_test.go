package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"
)

func (suite *IntegrationTestSuite) TestSubmitTransactionDispute() {
	defer SetupMockForLedger(suite).Close()

	user := suite.createTestUser(PartialMasterUserRecordDao{LedgerCustomerNumber: utils.Pointer("100000000034052")})

	request := dao.SubmitTransactionDisputeRequest{
		Reason:  "Duplicate transaction",
		Details: "I am not sure what this charge is for.",
	}
	requestBody, err := json.Marshal(request)
	suite.Require().NoError(err, "Failed to marshall request body")
	referenceId := "ledger.ach.transfer_ach_pull_1755001708162912900"
	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/account/customer/transaction/%s/disputes", referenceId), bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/customer/transaction/:referenceId/dispute")

	c.SetParamNames("referenceId")
	c.SetParamValues(referenceId)

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, "publicKey", c)

	err = handler.SubmitTransactionDispute(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusCreated, rec.Code, "Expected status code 201 Created")

	var transactionDisputeRecord dao.TransactionDisputeDao
	err = suite.TestDB.Model(&dao.TransactionDisputeDao{}).Where("transaction_identifier = ?", referenceId).First(&transactionDisputeRecord).Error
	suite.Require().NoError(err, "Failed to fetch transaction dispute record from database")
	suite.Require().Equal("pending", transactionDisputeRecord.Status, "Invalid status")
	suite.Require().NotEmpty(transactionDisputeRecord.TransactionIdentifier, "Process id should not be empty")
	suite.Require().Equal("ledger.ach.transfer_ach_pull_1755001708162912900", transactionDisputeRecord.TransactionIdentifier, "Process id must be valid")
	suite.Require().NotEmpty(transactionDisputeRecord.UserId, "User id should not be empty")
	suite.Require().Equal(user.Id, transactionDisputeRecord.UserId, "User id must be valid")
	suite.Require().Equal("Duplicate transaction", transactionDisputeRecord.Reason, "Invalid reason")
	suite.Equal("I am not sure what this charge is for.", transactionDisputeRecord.Details, "Invalid details")
}

func (suite *IntegrationTestSuite) TestSubmitTransactionDisputeForEmptyTransaction() {
	defer SetupMockForLedger(suite).Close()
	user := suite.createTestUser(PartialMasterUserRecordDao{})

	request := dao.SubmitTransactionDisputeRequest{
		Reason:  "Duplicate transaction",
		Details: "I am not sure what this charge is for.",
	}
	requestBody, err := json.Marshal(request)
	suite.Require().NoError(err, "Failed to marshall request body")

	referenceId := "emptyTransaction"
	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/account/customer/transaction/%s/disputes", referenceId), bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/customer/transaction/:referenceId/dispute")

	c.SetParamNames("referenceId")
	c.SetParamValues(referenceId)

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, "publicKey", c)

	err = handler.SubmitTransactionDispute(customContext)
	suite.NotNil(err, "Handler should return an error")

	errResp := err.(response.ErrorResponse)
	suite.Equal(http.StatusNotFound, errResp.StatusCode, "Expected status code 404 Not Found")
	suite.Equal("TRANSACTION_DOES_NOT_EXIST", errResp.ErrorCode, "Expected error code TRANSACTION_DOES_NOT_EXIST")
}

func (suite *IntegrationTestSuite) TestSubmitTransactionDisputeForInvalidTransactionType() {
	defer SetupMockForLedger(suite).Close()

	user := suite.createTestUser(PartialMasterUserRecordDao{LedgerCustomerNumber: utils.Pointer("100000000039103")})

	request := dao.SubmitTransactionDisputeRequest{
		Reason:  "Duplicate transaction",
		Details: "I am not sure what this charge is for.",
	}
	requestBody, err := json.Marshal(request)
	suite.Require().NoError(err, "Failed to marshall request body")
	referenceId := "voidTransaction"
	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/account/customer/transaction/%s/disputes", referenceId), bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/customer/transaction/:referenceId/dispute")

	c.SetParamNames("referenceId")
	c.SetParamValues(referenceId)

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, "publicKey", c)

	err = handler.SubmitTransactionDispute(customContext)
	suite.NotNil(err, "Handler should return an error")

	errResp := err.(response.ErrorResponse)
	suite.Equal(http.StatusUnprocessableEntity, errResp.StatusCode, "Expected status code 422 Unprocessable Entity")
	suite.Equal("DISPUTE_REQUEST_INVALID_TRANSACTION_TYPE", errResp.ErrorCode, "Expected error code DISPUTE_REQUEST_INVALID_TRANSACTION_TYPE")
}

func (suite *IntegrationTestSuite) TestSubmitTransactionDisputeForTransactionctionThatDoesNotBelongToUser() {
	defer SetupMockForLedger(suite).Close()

	user := suite.createTestUser(PartialMasterUserRecordDao{})

	request := dao.SubmitTransactionDisputeRequest{
		Reason:  "Duplicate transaction",
		Details: "I am not sure what this charge is for.",
	}
	requestBody, err := json.Marshal(request)
	suite.Require().NoError(err, "Failed to marshall request body")
	referenceId := "ledger.ach.transfer_ach_pull_1755001708162912900"
	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/account/customer/transaction/%s/disputes", referenceId), bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/customer/transaction/:referenceId/dispute")

	c.SetParamNames("referenceId")
	c.SetParamValues(referenceId)

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, "publicKey", c)

	err = handler.SubmitTransactionDispute(customContext)
	suite.Require().Error(err, "Handler should return an error")

	var responseErr response.ErrorResponse
	suite.Require().ErrorAs(err, &responseErr, "handler should return an ErrorResponse")
	suite.Require().Equal(500, responseErr.StatusCode, "handler should return 500")
}
