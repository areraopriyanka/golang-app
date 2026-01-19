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
	"process-api/pkg/model/response"
	"process-api/pkg/plaid"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"github.com/labstack/echo/v4"
	plaid_go "github.com/plaid/plaid-go/v34/plaid"
)

func (suite *IntegrationTestSuite) TestPlaidExchangePublicTokenSuccess() {
	h := suite.newHandler()
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)

	e := handler.NewEcho()
	e.POST("/account/plaid/public_token/exchange", func(c echo.Context) error {
		cc := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)
		return h.PlaidExchangePublicToken(cc)
	})

	// Plaid's sdk will check this format; this came from an actual api call:
	publicToken := "public-sandbox-5244f171-a521-489b-8176-82ab9d2c61e4"
	account := plaid.PlaidLinkAccount{
		ID:      "testid0",
		Type:    "depository",
		Subtype: "checking",
	}
	accounts := []plaid.PlaidLinkAccount{account}
	exchangeRequest := handler.PlaidExchangePublicTokenRequest{
		PlaidLinkOnSuccess: plaid.PlaidLinkOnSuccess{
			Accounts:      accounts,
			LinkSessionID: "link-session-id",
		},
		PublicToken: publicToken,
	}
	requestBody, _ := json.Marshal(exchangeRequest)

	req := httptest.NewRequest(http.MethodPost, "/account/plaid/public_token/exchange", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Equal(http.StatusCreated, rec.Code, "Response code should be 201 for successfully exchanging a public token")

	plaidItemId := "j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB4av" // hardcoded in mockoon
	var record dao.PlaidItemDao
	err := db.DB.Where("user_id = ? AND plaid_item_id = ?", user.Id, plaidItemId).First(&record).Error
	suite.Require().NoError(err, "A PlaidItem record was not created")

	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b4" // hardcoded in mockoon
	suite.NotEqual(unencryptedAccessToken, string(record.KmsEncryptedAccessToken), "Plaid access tokens must be encrypted")
	accessToken, _ := utils.DecryptKmsBinary(record.KmsEncryptedAccessToken)
	suite.Equal(unencryptedAccessToken, accessToken, "Plaid access tokens must be properly encrypted")

	plaidAccountID := "vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK980" // hardcoded in mockoon
	var plaidAccount dao.PlaidAccountDao
	err = db.DB.Model(dao.PlaidAccountDao{}).Where("plaid_account_id=?", plaidAccountID).Find(&plaidAccount).Error
	suite.Require().NoError(err)
	suite.Equal(plaid_go.AccountSubtype("checking"), plaidAccount.Subtype)
	suite.Equal(int64(10000), *plaidAccount.AvailableBalanceCents)
	suite.Equal("9600", *plaidAccount.Mask)
}

func (suite *IntegrationTestSuite) TestPlaidExchangePublicTokenFailure() {
	h := suite.newHandler()
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)

	e := handler.NewEcho()
	e.POST("/account/plaid/public_token/exchange", func(c echo.Context) error {
		cc := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)
		return h.PlaidExchangePublicToken(cc)
	})

	account := plaid.PlaidLinkAccount{
		ID:      "testid0",
		Type:    "depository",
		Subtype: "checking",
	}
	accounts := []plaid.PlaidLinkAccount{account}
	exchangeRequest := handler.PlaidExchangePublicTokenRequest{
		PlaidLinkOnSuccess: plaid.PlaidLinkOnSuccess{
			Accounts:      accounts,
			LinkSessionID: "link-session-id",
		},
	}
	requestBody, _ := json.Marshal(exchangeRequest)
	req := httptest.NewRequest(http.MethodPost, "/account/plaid/public_token/exchange", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Equal(http.StatusBadRequest, rec.Code, "Response code should be 400 for missing PublicToken in body")

	var errorResponse response.BadRequestErrors
	err := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	suite.Require().NoError(err, "Failed to parse response body")
	suite.Equal(errorResponse.Errors, []response.BadRequestError{{FieldName: "publicToken", Error: "required"}}, "publicToken is required")
}

func (suite *IntegrationTestSuite) callPlaidPublicTokenExchangeForAutomatedMicroDeposits(userToken string) string {
	h := suite.newHandler()
	e := handler.NewEcho()
	defer func() { _ = e.Shutdown(context.Background()) }()
	h.BuildRoutes(e, "", "test")

	// All the data needed to construct the exchangeRequest comes would normally
	// come from Plaid Link sending the data to the client, which would then hit
	// the middleware's plaid/public_token/exchange endpoint.
	plaidItemID := "DWVAAPWq4RHGlEaNyGKRTAnPLaEmo8Cvq7aa0"
	publicToken := "public-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379caa0"
	linkAccount := plaid.PlaidLinkAccount{
		ID:                 plaidItemID,
		Type:               "depository",
		Subtype:            "savings",
		VerificationStatus: utils.Pointer("pending_automatic_verification"),
	}
	accounts := []plaid.PlaidLinkAccount{linkAccount}
	exchangeRequest := handler.PlaidExchangePublicTokenRequest{
		PlaidLinkOnSuccess: plaid.PlaidLinkOnSuccess{
			InstitutionID: utils.Pointer("ins_117650"),
			Accounts:      accounts,
			LinkSessionID: "link-session-id",
		},
		PublicToken: publicToken,
	}
	requestBody, err := json.Marshal(exchangeRequest)
	suite.Require().NoError(err, "Failed to marshal request body")

	req := httptest.NewRequest(http.MethodPost, "/account/plaid/public_token/exchange", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	suite.Require().Equal(http.StatusCreated, rec.Code, "Expected status code 201 Created")
	return plaidItemID
}

func (suite *IntegrationTestSuite) TestPlaidExchangePublicToken_AutomatedMicroDeposits() {
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(userRecord.Id)
	token, err := security.GenerateOnboardedJwt(userRecord.Id, userPublicKey.PublicKey, nil)
	suite.Require().NoError(err, "Failed to generate JWT token")

	plaidItemID := suite.callPlaidPublicTokenExchangeForAutomatedMicroDeposits(token)

	var item dao.PlaidItemDao
	err = suite.TestDB.Model(dao.PlaidItemDao{}).Where("plaid_item_id=?", plaidItemID).First(&item).Error
	suite.Require().NoError(err, "Failed to get an item")

	var account dao.PlaidAccountDao
	err = suite.TestDB.Model(dao.PlaidAccountDao{}).Where("plaid_item_id=?", plaidItemID).First(&account).Error
	suite.Require().NoError(err, "Failed to get an account")

	suite.Require().Equal("AUTOMATED_MICRODEPOSITS", string(*account.AuthMethod), "Account should be AUTOMATED_MICRODEPOSITS")
	suite.Require().Equal("Baku Endif", *account.PrimaryOwnerName, "Account should be have primary owner name set for automated micro-deposits flow")
	suite.Require().Nil(account.AvailableBalanceCents, "Account should not yet have available balance.")
	suite.Require().Nil(account.BalanceRefreshedAt, "Automated micro-deposit linked accounts do not have up-to-date balance")
	suite.Require().Equal("pending_automatic_verification", *account.VerificationStatus, "Account should be pending automatic verification")
}
