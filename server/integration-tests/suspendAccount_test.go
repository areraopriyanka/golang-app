package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/plaid"
	"process-api/pkg/security"

	"github.com/riverqueue/river"
)

func (suite *IntegrationTestSuite) TestSuspendAccount() {
	defer SetupMockForLedger(suite).Close()

	sessionId := suite.SetupTestData()
	userPublicKey := suite.createUserPublicKeyRecord(sessionId)

	closeAccountRequest := handler.CloseAccountRequest{
		AccountClosureReason: "Not using it",
	}
	requestBody, err := json.Marshal(closeAccountRequest)
	suite.Require().NoError(err, "Failed to marshall request")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPut, "/account/close", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/close")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, userPublicKey.PublicKey, c)

	err = handler.SuspendAccount(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.UpdateAccountStatusResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().NotEmpty(responseBody.UpdatedAccountStatus, "UpdatedAccountStatus should not be empty")
	suite.Require().Equal("temporary_inactive", responseBody.UpdatedAccountStatus, "Updated account status should be temporary_inactive")

	var userAccountCard dao.UserAccountCardDao
	err = suite.TestDB.Model(&dao.UserAccountCardDao{}).Where("user_id = ?", sessionId).First(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to get userAccountCard record from db")
	suite.Equal(userAccountCard.AccountStatus, ledger.SUSPENDED, "Account Status is updated to SUSPENDED")
	suite.Equal(userAccountCard.AccountClosureReason, "Not using it", "Account Closure Reason is saved in record")
}

func (suite *IntegrationTestSuite) TestCloseAccountWorkMethod() {
	defer SetupMockForLedger(suite).Close()
	h := suite.newHandler()
	sessionId := suite.SetupTestData()

	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}

	plaidItemID := "DWVAAPWq4RHGlEaNyGKRTAnPLaEmo8Cvq7nc0"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0c0"

	err := ps.InsertItem(sessionId, plaidItemID, unencryptedAccessToken)
	suite.Require().NoError(err)

	err = ps.InitialAccountsGetRequest(sessionId, plaidItemID, unencryptedAccessToken)
	suite.Require().NoError(err)

	var item dao.PlaidItemDao
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_item_id=?", plaidItemID).Find(&item).Error
	suite.Require().NoError(err)

	args := handler.CloseSuspendedAccountArgs{UserId: sessionId}
	job := &river.Job[handler.CloseSuspendedAccountArgs]{Args: args}

	worker := handler.CloseSuspendedAccountWorker{Plaid: h.Plaid}
	err = worker.Work(context.Background(), job)
	suite.Require().NoError(err)

	userAccountCard, _ := dao.UserAccountCardDao{}.FindOneByAccountNumber(suite.TestDB, "123456789012345")
	suite.Require().Equal("CLOSED", userAccountCard.AccountStatus, "Account must be CLOSED")
	suite.Require().Equal("Account closed due to completion of the 60-day suspension period", userAccountCard.AccountClosureReason, "Invalid account closure reason")

	// Ensure all external bank data is deleted from the plaid_items and plaid_accounts tables.
	var plaidItems []dao.PlaidItemDao
	err = db.DB.Where("plaid_item_id=?", plaidItemID).Find(&plaidItems).Error
	suite.Require().NoError(err, "failed to get plaid item record")
	suite.Require().Len(plaidItems, 0, "all plaid items associated with the deleted account must be deleted")

	var accounts []dao.PlaidAccountDao
	err = db.DB.Where("user_id = ?", sessionId).Find(&accounts).Error
	suite.Require().NoError(err, "failed to get account records")
	suite.Len(accounts, 0, "all accounts associated with the deleted account's item must also be deleted")
}
