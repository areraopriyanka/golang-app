package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/clock"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/logging"
	"process-api/pkg/plaid"
	"process-api/pkg/security"
	"sort"
	"strconv"
	"time"
)

type refreshBalancesContext struct {
	handler       *handler.Handler
	user          dao.MasterUserRecordDao
	userPublicKey dao.UserPublicKey
	token         string
	plaidItemId   string
}

const (
	doNotRefreshStaleAccounts = iota
	refreshStaleAccounts
	doNotRefreshStaleAccountsDueToItemErrors
)

func (suite *IntegrationTestSuite) TestRefreshBalancesWithStaleAccounts() {
	ctx := suite.newRefreshBalancesContext(refreshStaleAccounts)
	e := handler.NewEcho()
	ctx.handler.BuildRoutes(e, "", "test")

	req := httptest.NewRequest(http.MethodPost, "/account/balance/refresh", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ctx.token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Require().Equal(http.StatusCreated, rec.Code, "Expected status code 201 Created")
	var refreshResponseBody handler.BalanceRefreshResponse
	err := json.Unmarshal(rec.Body.Bytes(), &refreshResponseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")
	suite.Require().NotNil(refreshResponseBody.JobId)
	jobId := strconv.Itoa(int(*refreshResponseBody.JobId))

	suite.WaitForJobsDone(1)

	var statusResponseBody handler.BalanceRefreshStatusResponse
	req2 := httptest.NewRequest(http.MethodGet, "/account/balance/refresh/status/"+jobId, nil)
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+ctx.token)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	suite.Require().Equal(http.StatusOK, rec2.Code, "Expected status code 200 OK")

	err = json.Unmarshal(rec2.Body.Bytes(), &statusResponseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().False(statusResponseBody.IsRefreshing, "Job should be completed")
	suite.Require().Len(statusResponseBody.Balances, 2, "Should return 2 account balances")

	balances := statusResponseBody.Balances
	// Just to get a deterministic ordering
	sort.Slice(balances, func(i, j int) bool {
		return *balances[i].AvailableBalanceCents < *balances[j].AvailableBalanceCents
	})

	var accounts []dao.PlaidAccountDao
	err = db.DB.Model(dao.PlaidAccountDao{}).Where("plaid_item_id=?", ctx.plaidItemId).Order("plaid_account_id asc").Find(&accounts).Error
	suite.Require().NoError(err)
	suite.Require().Len(accounts, 2, "must return 2 accounts")
	for _, account := range accounts {
		recently := clock.Now().Add(-1 * time.Minute)
		suite.Require().NotNil(account.BalanceRefreshedAt, "Balance refreshed at should not be nil")
		suite.Require().True(account.BalanceRefreshedAt.After(recently), "Balance should have been refreshed recently")
	}
	suite.Require().Equal(accounts[0].ID, balances[0].ID, "Should return matching IDs")
	suite.Require().Equal(accounts[1].ID, balances[1].ID, "Should return matching IDs")
	suite.Require().Equal(accounts[0].AvailableBalanceCents, balances[0].AvailableBalanceCents, "Should return matching AvailableBalanceCents")
	suite.Require().Equal(accounts[1].AvailableBalanceCents, balances[1].AvailableBalanceCents, "Should return matching AvailableBalanceCents")
	suite.Require().Equal(int64(5000), *balances[0].AvailableBalanceCents, "Should return the correct account balance")
	suite.Require().Equal(int64(6000), *balances[1].AvailableBalanceCents, "Should return the correct account balance")
}

func (suite *IntegrationTestSuite) TestRefreshBalancesNoStaleAccounts() {
	ctx := suite.newRefreshBalancesContext(doNotRefreshStaleAccounts)
	e := handler.NewEcho()
	ctx.handler.BuildRoutes(e, "", "test")
	req := httptest.NewRequest(http.MethodPost, "/account/balance/refresh", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ctx.token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK - no job needed")

	var refreshResponseBody handler.BalanceRefreshResponse
	err := json.Unmarshal(rec.Body.Bytes(), &refreshResponseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")
	suite.Require().Nil(refreshResponseBody.JobId)
}

func (suite *IntegrationTestSuite) TestRefreshBalancesNoStaleAccountsDueToItemErrors() {
	ctx := suite.newRefreshBalancesContext(doNotRefreshStaleAccountsDueToItemErrors)
	e := handler.NewEcho()
	ctx.handler.BuildRoutes(e, "", "test")
	req := httptest.NewRequest(http.MethodPost, "/account/balance/refresh", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ctx.token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK - no job needed")

	var refreshResponseBody handler.BalanceRefreshResponse
	err := json.Unmarshal(rec.Body.Bytes(), &refreshResponseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")
	suite.Require().Nil(refreshResponseBody.JobId)
}

func (suite *IntegrationTestSuite) TestBalanceRefreshStatusNoJobs() {
	ctx := suite.newRefreshBalancesContext(doNotRefreshStaleAccounts)
	e := handler.NewEcho()
	ctx.handler.BuildRoutes(e, "", "test")
	req := httptest.NewRequest(http.MethodGet, "/account/balance/refresh/status/1", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ctx.token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Require().Equal(http.StatusNotFound, rec.Code, "Expected status code 404")
}

func (suite *IntegrationTestSuite) newRefreshBalancesContext(createStaleAccounts int) refreshBalancesContext {
	h := suite.newHandler()
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)
	_ = suite.createUserAccountCard(userRecord.Id)

	token, err := security.GenerateOnboardedJwt(userRecord.Id, userPublicKey.PublicKey, nil)
	suite.Require().NoError(err, "Failed to generate JWT token")

	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}

	// Values are hardcoded in mockoon
	plaidItemId := "DWVAAPWq4RHGlEaNyGKRTAnPLaEmo8Cvq7na6"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b2"
	switch createStaleAccounts {
	case refreshStaleAccounts:
		plaidItemId = "DWVAAPWq4RHGlEaNyGKRTAnPLaEmo8Cvq7nd0"
		unencryptedAccessToken = "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0d0"
	case doNotRefreshStaleAccountsDueToItemErrors:
		plaidItemId = "DWVAAPWq4RHGlEaNyGKRTAnPLaEmo8Cvq7nd1"
		unencryptedAccessToken = "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0d1"
	}
	err = ps.InsertItem(userRecord.Id, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userRecord.Id, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)

	var accounts []dao.PlaidAccountDao
	if createStaleAccounts == refreshStaleAccounts || createStaleAccounts == doNotRefreshStaleAccountsDueToItemErrors {
		now := clock.Now()
		balanceRefreshedAt := now.Add(-25 * time.Hour)
		err = ps.DB.Model(&dao.PlaidAccountDao{}).
			Where("plaid_item_id = ?", plaidItemId).
			Update("balance_refreshed_at", balanceRefreshedAt).Error
		suite.Require().NoError(err)
		err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_item_id=?", plaidItemId).Order("plaid_account_id asc").Find(&accounts).Error
		suite.Require().NoError(err)
	}

	if createStaleAccounts == doNotRefreshStaleAccountsDueToItemErrors {
		err = ps.DB.Model(&dao.PlaidItemDao{}).
			Where("plaid_item_id = ?", plaidItemId).
			Update("item_error", "ERROR").Error
		suite.Require().NoError(err)
	}

	return refreshBalancesContext{
		handler:       h,
		user:          userRecord,
		userPublicKey: userPublicKey,
		token:         token,
		plaidItemId:   plaidItemId,
	}
}
