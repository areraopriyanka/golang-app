package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/logging"
	"process-api/pkg/plaid"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Initial test scaffolding created with Claud Code 1.0.41
func (suite *IntegrationTestSuite) TestAchAccounts() {
	defer SetupMockForLedger(suite).Close()
	h := suite.newHandler()
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)
	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH000006009011111",
		CardId:        "6f586be7bf1c44b8b4ea11b2e2510e25",
		UserId:        user.Id,
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
	}
	err := suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	e := handler.NewEcho()
	e.GET("/ach/accounts", func(c echo.Context) error {
		cc := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)
		return h.AchAccounts(cc)
	})

	req := httptest.NewRequest(http.MethodGet, "/ach/accounts", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody handler.AchAccountsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().NotNil(responseBody.DreamFiAccounts, "DreamFiAccounts should not be nil")
	suite.Require().NotNil(responseBody.ExternalAccounts, "ExternalAccounts should not be nil")

	suite.Require().Len(responseBody.DreamFiAccounts, 1, "must return 1 DreamFi account")
	suite.Require().Len(responseBody.ExternalAccounts, 0, "must return 0 external accounts")

	dreamFiAccount := responseBody.DreamFiAccounts[0]
	suite.Require().Equal("217140", dreamFiAccount.ID, "DreamFi account ID should match")
	suite.Require().NotNil(dreamFiAccount.Institution, "Institution should not be nil")
	suite.Require().Equal("XD Legder", *dreamFiAccount.Institution, "Institution should be 'XD Legder'")
	suite.Require().Equal("Default DreamFi Account", dreamFiAccount.Name, "Account name should match")
	suite.Require().Equal("CHECKING", dreamFiAccount.Subtype, "Account subtype should be CHECKING")
	suite.Require().NotNil(dreamFiAccount.Mask, "Mask should not be nil")
	suite.Require().NotNil(dreamFiAccount.AvailableBalanceCents, "AvailableBalanceCents should not be nil")
	suite.Require().Equal(int64(10000), *dreamFiAccount.AvailableBalanceCents, "AvailableBalanceCents should be $100.00")
}

func (suite *IntegrationTestSuite) TestAchAccountsWithExternalAccounts() {
	defer SetupMockForLedger(suite).Close()
	h := suite.newHandler()
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)
	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH000006009011111",
		CardId:        "6f586be7bf1c44b8b4ea11b2e2510e25",
		UserId:        user.Id,
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
	}
	err := suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	plaidItemId := "j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB4a0"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b0"
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	err = ps.InsertItem(user.Id, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)

	checkingAccount := dao.PlaidAccountDao{
		PlaidAccountID:        "test-plaid-account-checking-0",
		UserID:                user.Id,
		PlaidItemID:           plaidItemId,
		Name:                  "Test Checking Account",
		Subtype:               dao.CheckingSubtype,
		Mask:                  utils.Pointer("1234"),
		InstitutionID:         utils.Pointer("ins_test"),
		InstitutionName:       utils.Pointer("Test Bank"),
		AvailableBalanceCents: utils.Pointer[int64](10000),
	}
	err = suite.TestDB.Create(&checkingAccount).Error
	suite.Require().NoError(err, "Failed to create checking account")

	savingsAccount := dao.PlaidAccountDao{
		PlaidAccountID:        "test-plaid-account-savings-0",
		UserID:                user.Id,
		PlaidItemID:           plaidItemId,
		Name:                  "Test Savings Account",
		Subtype:               dao.SavingsSubtype,
		Mask:                  utils.Pointer("5678"),
		InstitutionID:         utils.Pointer("ins_test"),
		InstitutionName:       utils.Pointer("Test Bank"),
		AvailableBalanceCents: utils.Pointer[int64](25000),
	}
	err = suite.TestDB.Create(&savingsAccount).Error
	suite.Require().NoError(err, "Failed to create savings account")

	e := handler.NewEcho()
	e.GET("/ach/accounts", func(c echo.Context) error {
		cc := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)
		return h.AchAccounts(cc)
	})

	req := httptest.NewRequest(http.MethodGet, "/ach/accounts", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody handler.AchAccountsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().NotNil(responseBody.DreamFiAccounts, "DreamFiAccounts should not be nil")
	suite.Require().NotNil(responseBody.ExternalAccounts, "ExternalAccounts should not be nil")

	suite.Require().Len(responseBody.DreamFiAccounts, 1, "must return 1 DreamFi account")
	suite.Require().Len(responseBody.ExternalAccounts, 2, "must return 2 external accounts")

	var checkingExt, savingsExt *handler.AchAccount
	for i := range responseBody.ExternalAccounts {
		switch responseBody.ExternalAccounts[i].Subtype {
		case "checking":
			checkingExt = &responseBody.ExternalAccounts[i]
		case "savings":
			savingsExt = &responseBody.ExternalAccounts[i]
		}
	}

	suite.Require().NotNil(checkingExt, "checking account should be found")
	suite.Require().Equal(checkingAccount.ID, checkingExt.ID)
	suite.Require().Equal("Test Checking Account", checkingExt.Name)
	suite.Require().Equal("checking", checkingExt.Subtype)
	suite.Require().NotNil(checkingExt.Institution)
	suite.Require().Equal("Test Bank", *checkingExt.Institution)
	suite.Require().NotNil(checkingExt.Mask)
	suite.Require().Equal("1234", *checkingExt.Mask)
	suite.Require().NotNil(checkingExt.AvailableBalanceCents)
	suite.Require().Equal(int64(10000), *checkingExt.AvailableBalanceCents)

	suite.Require().NotNil(savingsExt, "savings account should be found")
	suite.Require().Equal(savingsAccount.ID, savingsExt.ID)
	suite.Require().Equal("Test Savings Account", savingsExt.Name)
	suite.Require().Equal("savings", savingsExt.Subtype)
	suite.Require().NotNil(savingsExt.Institution)
	suite.Require().Equal("Test Bank", *savingsExt.Institution)
	suite.Require().NotNil(savingsExt.Mask)
	suite.Require().Equal("5678", *savingsExt.Mask)
	suite.Require().NotNil(savingsExt.AvailableBalanceCents)
	suite.Require().Equal(int64(25000), *savingsExt.AvailableBalanceCents)
}

func (suite *IntegrationTestSuite) TestAchAccountsUnauthorized() {
	defer SetupMockForLedger(suite).Close()
	h := suite.newHandler()

	e := handler.NewEcho()
	e.GET("/ach/accounts", func(c echo.Context) error {
		// No authentication context provided
		return h.AchAccounts(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/ach/accounts", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Require().Equal(http.StatusUnauthorized, rec.Code, "Expected 401 Unauthorized")
}

func (suite *IntegrationTestSuite) TestAchAccountsUserNotFound() {
	defer SetupMockForLedger(suite).Close()
	h := suite.newHandler()

	nonExistentUserId := uuid.New().String()
	publicKey := "test-public-key"

	e := handler.NewEcho()
	e.GET("/ach/accounts", func(c echo.Context) error {
		cc := security.GenerateLoggedInRegisteredUserContext(nonExistentUserId, publicKey, c)
		return h.AchAccounts(cc)
	})

	req := httptest.NewRequest(http.MethodGet, "/ach/accounts", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Require().Equal(http.StatusNotFound, rec.Code, "Expected 404 Not Found")
}

func (suite *IntegrationTestSuite) TestAchAccounts_WithExternalAccounts_WithErrors() {
	defer SetupMockForLedger(suite).Close()

	h := suite.newHandler()
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	_ = suite.createUserAccountCard(userRecord.Id)
	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)
	token, err := security.GenerateOnboardedJwt(userRecord.Id, userPublicKey.PublicKey, nil)
	suite.Require().NoError(err, "Failed to generate JWT token")
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	item := suite.createPlaidItemWithCheckingAndSavingsAccounts(ps, userRecord)
	err = dao.PlaidItemDao{}.SetItemError(item.PlaidItemID, "LOGIN_REQUIRED")
	suite.Require().NoError(err, "Failed to set plaid item error")

	e := handler.NewEcho()
	h.BuildRoutes(e, "", "test")
	req := httptest.NewRequest(http.MethodGet, "/account/ach/accounts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody handler.AchAccountsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")
	suite.Require().Len(responseBody.ExternalAccounts, 2, "must return 2 external accounts")

	accountA := responseBody.ExternalAccounts[0]
	accountB := responseBody.ExternalAccounts[1]

	suite.Require().True(accountA.NeedsPlaidLinkUpdate, "account should have indication it needs to be updated")
	suite.Require().True(accountB.NeedsPlaidLinkUpdate, "account should have indication it needs to be updated")
}
