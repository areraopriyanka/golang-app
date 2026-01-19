package test

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"process-api/pkg/clock"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/logging"
	"process-api/pkg/plaid"
	"process-api/pkg/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	plaid_go "github.com/plaid/plaid-go/v34/plaid"
)

func (suite *IntegrationTestSuite) beforePlaid() (*handler.Handler, string) {
	h := suite.newHandler()
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	return h, user.Id
}

func (suite *IntegrationTestSuite) createPlaidItemWithCheckingAndSavingsAccounts(ps plaid.PlaidService, user dao.MasterUserRecordDao) dao.PlaidItemDao {
	// All hardcoded values are set from mockoon
	plaidItemID := "DWVAAPWq4RHGlEaNyGKRTAnPLaEmo8Cvq7nc0"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0c0"
	err := ps.InsertItem(user.Id, plaidItemID, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(user.Id, plaidItemID, unencryptedAccessToken)
	suite.Require().NoError(err)
	var item dao.PlaidItemDao
	err = ps.DB.Where("plaid_item_id=?", plaidItemID).Find(&item).Error
	suite.Require().NoError(err)
	return item
}

func (suite *IntegrationTestSuite) TestLinkTokenCreateRequest_ios() {
	h, userId := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	linkToken, err := ps.LinkTokenCreateRequest(userId, "ios", h.Config.Plaid.LinkRedirectURI, "test", nil)
	suite.Require().NoError(err)
	// Values are hardcoded in mockoon
	suite.Equal("link-sandbox-5c2ad447-5e16-4905-90f1-317cfc0a7067", linkToken)
}

func (suite *IntegrationTestSuite) TestLinkTokenCreateRequest_android() {
	h, userId := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	linkToken, err := ps.LinkTokenCreateRequest(userId, "android", h.Config.Plaid.LinkRedirectURI, "test", nil)
	suite.Require().NoError(err)
	suite.Equal("link-sandbox-5c2ad447-5e16-4905-90f1-317cfc0a7067", linkToken)
}

func (suite *IntegrationTestSuite) TestItemPublicTokenExchangeRequest() {
	h, _ := suite.beforePlaid()
	// Plaid's sdk will check this format; this came from an actual api call:
	publicToken := "public-sandbox-5244f171-a521-489b-8176-82ab9d2c61e0"
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	resp, err := ps.ItemPublicTokenExchangeRequest(publicToken)
	suite.Require().NoError(err)

	plaidItemId := resp.PlaidItemId
	accessToken := resp.AccessToken
	// Values are hardcoded in mockoon
	suite.Equal("j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB4av", plaidItemId)
	suite.Equal("access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b4", accessToken)
}

// Plaid items represent a link to one or more checking/savings depository accounts.
func (suite *IntegrationTestSuite) TestPlaidInsertItem() {
	h, userId := suite.beforePlaid()
	// Values are hardcoded in mockoon
	plaidItemId := "j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB4a0"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b0"
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	err := ps.InsertItem(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)

	var record dao.PlaidItemDao
	err = ps.DB.Where("plaid_item_id = ?", plaidItemId).First(&record).Error
	suite.Require().NoError(err)
	suite.Equal("j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB4a0", record.PlaidItemID)
	accessToken, err := utils.DecryptKmsBinary(record.KmsEncryptedAccessToken)
	suite.Require().NoError(err)
	suite.Equal("access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b0", accessToken)

	// ensure unique constraint holds
	err = ps.InsertItem(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().Contains(err.Error(), "duplicate key value violates unique constraint")
}

func (suite *IntegrationTestSuite) TestInitialAccountsGetRequest_Checking_OK() {
	h, userId := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	// All hardcoded values are set from mockoon
	plaidItemId := "j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB4a0"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b0"
	err := ps.InsertItem(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)

	plaidAccountID := "vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK980"
	var record dao.PlaidAccountDao
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_account_id=?", plaidAccountID).Find(&record).Error
	suite.Require().NoError(err)
	suite.Equal(plaid_go.AccountSubtype("checking"), record.Subtype)
	suite.Equal(int64(10000), *record.AvailableBalanceCents)
	suite.Equal("9600", *record.Mask)
}

func (suite *IntegrationTestSuite) TestAccountsBalanceGetRequest_Checking_OK() {
	h, userId := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	// All hardcoded values are set from mockoon
	plaidItemId := "j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB4a0"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b0"
	err := ps.InsertItem(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.AccountsBalanceGetRequest(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)

	plaidAccountID := "vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK980"
	var record dao.PlaidAccountDao
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_account_id=?", plaidAccountID).Find(&record).Error
	suite.Require().NoError(err)
	suite.Equal(plaid_go.AccountSubtype("checking"), record.Subtype)
	suite.Equal(int64(10000), *record.AvailableBalanceCents)
	suite.Equal("9600", *record.Mask)
}

func (suite *IntegrationTestSuite) TestAccountsBalanceGetRequest_Savings_OK() {
	h, userId := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	plaidItemId := "j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB4a1"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b1"
	err := ps.InsertItem(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.AccountsBalanceGetRequest(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)

	plaidAccountID := "vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK981"
	var record dao.PlaidAccountDao
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_account_id=?", plaidAccountID).Find(&record).Error
	suite.Require().NoError(err)
	suite.Equal(plaid_go.AccountSubtype("savings"), record.Subtype)
	suite.Equal(int64(10000), *record.AvailableBalanceCents)
	suite.Equal("9601", *record.Mask)
}

// We currently only support checking and savings depository accounts
func (suite *IntegrationTestSuite) TestInitialAccountsGetRequest_Invalid401k_Error() {
	h, userId := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	plaidItemId := "j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB400"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c000"
	err := ps.InsertItem(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().EqualError(err, "invalid subtype for account; must be checking or savings invalid subtype")
}

// Plaid items can represent multiple accounts; this covers getting the balance for an
// item associated with checking and savings accounts.
func (suite *IntegrationTestSuite) TestAccountsBalanceGetRequest_CheckingAndSavings_OK() {
	h, userId := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	plaidItemId := "j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB4a2"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b2"
	err := ps.InsertItem(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.AccountsBalanceGetRequest(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)

	var records []dao.PlaidAccountDao
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_item_id=?", plaidItemId).Order("plaid_account_id asc").Find(&records).Error
	suite.Require().NoError(err)
	suite.Require().Len(records, 2, "must return 2 accounts")
	record0 := records[0]
	suite.Equal("vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK982", record0.PlaidAccountID)
	suite.Equal(plaid_go.AccountSubtype("checking"), record0.Subtype)
	suite.Equal(int64(5000), *record0.AvailableBalanceCents)
	suite.Equal("9602", *record0.Mask)
	record1 := records[1]
	suite.Equal("vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK983", record1.PlaidAccountID)
	suite.Equal(plaid_go.AccountSubtype("savings"), record1.Subtype)
	suite.Equal(int64(6000), *record1.AvailableBalanceCents)
	suite.Equal("9603", *record1.Mask)
}

func (suite *IntegrationTestSuite) TestInitialAccountsGetRequest_CheckingAndInvalid401_OK() {
	h, userId := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	plaidItemId := "j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB401"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c001"
	err := ps.InsertItem(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userId, plaidItemId, unencryptedAccessToken)
	// The 401k that was returned should result in an error:
	suite.Require().EqualError(err, "invalid subtype for account; must be checking or savings invalid subtype")

	// Even though one of the accounts had an error, we should still create a record for the non-error account:
	var records []dao.PlaidAccountDao
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_item_id=?", plaidItemId).Order("plaid_account_id asc").Find(&records).Error
	suite.Require().NoError(err)
	suite.Require().Len(records, 1, "must return 1 valid account")
	record := records[0]
	suite.Equal(plaid_go.AccountSubtype("checking"), record.Subtype)
}

func (suite *IntegrationTestSuite) TestAccountsBalanceGetRequest_CheckingIncreasedBalance_OK() {
	h, userId := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	// All hardcoded values are set from mockoon
	plaidItemId := "j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB4a0"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b0"
	err := ps.InsertItem(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.AccountsBalanceGetRequest(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)

	plaidAccountID := "vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK980"
	var record dao.PlaidAccountDao
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_account_id=?", plaidAccountID).Find(&record).Error
	suite.Require().NoError(err)
	suite.Equal(plaid_go.AccountSubtype("checking"), record.Subtype)
	suite.Equal(int64(10000), *record.AvailableBalanceCents)
	suite.Equal("9600", *record.Mask)

	// This access token tells mockoon to serve the same account details, but with an increased balance
	unencryptedAccessToken = "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b6"
	err = ps.AccountsBalanceGetRequest(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)

	plaidAccountID = "vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK980"
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_account_id=?", plaidAccountID).Find(&record).Error
	suite.Require().NoError(err)
	suite.Equal(plaid_go.AccountSubtype("checking"), record.Subtype)
	suite.Equal(int64(11000), *record.AvailableBalanceCents)
	suite.Equal("9600", *record.Mask)
}

func (suite *IntegrationTestSuite) TestAccountsGetIdentityRequest_CheckingAndSavings_OK() {
	h, userId := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	plaidItemId := "j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB4a2"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b2"
	err := ps.InsertItem(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.GetIdentity(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)

	var records []dao.PlaidAccountDao
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_item_id=?", plaidItemId).Order("plaid_account_id asc").Find(&records).Error
	suite.Require().NoError(err)
	suite.Require().Len(records, 2, "must return 2 accounts")
	record0 := records[0]
	suite.Equal("vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK982", record0.PlaidAccountID)
	suite.Equal(plaid_go.AccountSubtype("checking"), record0.Subtype)
	suite.Equal(int64(5000), *record0.AvailableBalanceCents)
	suite.Equal("9602", *record0.Mask)
	suite.Equal("Alberta Bobbeth Charleson", *record0.PrimaryOwnerName)
	record1 := records[1]
	suite.Equal("vzeNDwK7KQIm4yEog683uElbp9GRLEFXGK983", record1.PlaidAccountID)
	suite.Equal(plaid_go.AccountSubtype("savings"), record1.Subtype)
	suite.Equal(int64(6000), *record1.AvailableBalanceCents)
	suite.Equal("9603", *record1.Mask)
	suite.Equal("Alberta Bobbeth Charleson", *record0.PrimaryOwnerName)
}

func (suite *IntegrationTestSuite) TestCheckForDuplicateAccounts() {
	h, userId := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}
	account0 := plaid.PlaidLinkAccount{
		ID:      "testid0",
		Type:    "depository",
		Subtype: "checking",
	}
	accounts := []plaid.PlaidLinkAccount{account0}
	// No existing accounts; no duplicates
	hasDuplicates, err := ps.CheckForDuplicateAccounts(userId, nil, &accounts)
	suite.Require().NoError(err)
	suite.Require().Equal(false, hasDuplicates)

	// values come from mockoon; this access token triggers mockoon to return a checking and savings account
	plaidItemId := "j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB4a0"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b2"
	err = ps.InsertItem(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)

	institutionID := "ins_117654"
	var records []dao.PlaidAccountDao
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_item_id=?", plaidItemId).Order("plaid_account_id asc").Find(&records).Error
	suite.Require().NoError(err)
	suite.Require().Len(records, 2, "must return 2 accounts")

	account1 := plaid.PlaidLinkAccount{
		ID:   "testid1",
		Mask: records[0].Mask,
		Name: &records[0].Name,
	}
	account2 := plaid.PlaidLinkAccount{
		ID:   "testid2",
		Mask: records[1].Mask,
		Name: &records[1].Name,
	}
	// both accounts are duplicates
	accounts = []plaid.PlaidLinkAccount{account1, account2}
	hasDuplicates, err = ps.CheckForDuplicateAccounts(userId, &institutionID, &accounts)
	suite.Require().NoError(err)
	suite.Require().Equal(true, hasDuplicates)

	// one acount is a duplicates
	accounts = []plaid.PlaidLinkAccount{account1, account0}
	hasDuplicates, err = ps.CheckForDuplicateAccounts(userId, &institutionID, &accounts)
	suite.Require().NoError(err)
	suite.Require().Equal(true, hasDuplicates)
}

func (suite *IntegrationTestSuite) TestIsDuplicateAccount() {
	h, userId := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}

	account := plaid.PlaidLinkAccount{
		ID:      "testid0",
		Type:    "depository",
		Subtype: "checking",
	}
	// No existing accounts; no duplicates
	isDuplicate, err := ps.IsDuplicateAccount(userId, nil, &account)
	suite.Require().NoError(err)
	suite.Require().Equal(false, isDuplicate)

	// values come from mockoon; this access token triggers mockoon to return a checking and savings account
	plaidItemId := "j91ByvRRqwuGBygwnB8Au8j6ZvmjKAt1wB4a0"
	unencryptedAccessToken := "access-sandbox-1b7e6039-337b-34d7-a3cd-7e13e379c0b2"
	err = ps.InsertItem(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)
	err = ps.InitialAccountsGetRequest(userId, plaidItemId, unencryptedAccessToken)
	suite.Require().NoError(err)

	institutionID := "ins_117654"
	var records []dao.PlaidAccountDao
	err = ps.DB.Model(dao.PlaidAccountDao{}).Where("plaid_item_id=?", plaidItemId).Order("plaid_account_id asc").Find(&records).Error
	suite.Require().NoError(err)
	suite.Require().Len(records, 2, "must return 2 accounts")

	// if mask and/or name are missing, assume duplicate if user has linked to the institutionID before
	account.Mask = nil
	account.Name = nil
	isDuplicate, err = ps.IsDuplicateAccount(userId, &institutionID, &account)
	suite.Require().NoError(err)
	suite.Require().Equal(true, isDuplicate)

	record := records[0]
	account.Mask = record.Mask
	account.Name = &record.Name
	isDuplicate, err = ps.IsDuplicateAccount(userId, &institutionID, &account)
	suite.Require().NoError(err)
	suite.Require().Equal(true, isDuplicate)

	// mask and name, even if present, can overlap with other accounts at different institutions
	differentInstitutionID := "a_different_institution"
	isDuplicate, err = ps.IsDuplicateAccount(userId, &differentInstitutionID, &account)
	suite.Require().NoError(err)
	suite.Require().Equal(false, isDuplicate)

	// apparently, Link's `onSuccess` doesn't always return an institutionID
	isDuplicate, err = ps.IsDuplicateAccount(userId, nil, &account)
	suite.Require().NoError(err)
	suite.Require().Equal(false, isDuplicate)

	// ensure the other account linked will register for duplicates, too
	record = records[1]
	account.Mask = record.Mask
	account.Name = &record.Name
	isDuplicate, err = ps.IsDuplicateAccount(userId, &institutionID, &account)
	suite.Require().NoError(err)
	suite.Require().Equal(true, isDuplicate)
}

func (suite *IntegrationTestSuite) getPlaidWebhookPrivateKey(kid string) *ecdsa.PrivateKey {
	var privateKeyPEM string

	switch kid {
	case "test-key-id-123":
		privateKeyPEM = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIItMYFoLtbq5sI+4frLdc29L0Tf06HAanyrYKfdQmjmToAoGCCqGSM49
AwEHoUQDQgAEPAnDztuSc+oWn/Ib0VCzHx2NhCBhsnW69w0Y+Bn9trOEbY1U3kUT
mi8zSDWGMB+17elEIpPXOpUv0dvIm7v6Jw==
-----END EC PRIVATE KEY-----`
	case "test-key-id-456":
		privateKeyPEM = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIzrzAQRDCXvWWS7YvMb91mHBcPmYy17ZMNwhMexCvUxoAoGCCqGSM49
AwEHoUQDQgAEydX1GupZoBxGTDUzFYtZbCtVwUf57QXIl4dXJ2Kt3aI9ESgiGi6G
pWKALVLfKJXbDxoaoyLSzIm73KdQuum+jA==
-----END EC PRIVATE KEY-----`
	case "test-key-id-expired":
		privateKeyPEM = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIG1tZcPUxT+ySRgg1OrFsB1osfXmcpWOa9j6y3fhW1RioAoGCCqGSM49
AwEHoUQDQgAEuh6E5PXx2ZI1wEytm7YPnF01/Z5UGQJFFjm/3I1XOwz8yCbh40IR
+Zf0cgAGJVyxnPgexCCFH8QSmeMekFZ+BQ==
-----END EC PRIVATE KEY-----`
	default:
		suite.Require().Fail("Unknown test key ID: %s", kid)
		return nil
	}

	block, _ := pem.Decode([]byte(privateKeyPEM))
	suite.Require().NotNil(block, "Failed to parse PEM block")

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	suite.Require().NoError(err, "Failed to parse EC private key")

	return privateKey
}

func (suite *IntegrationTestSuite) createPlaidWebhookJWT(privateKey *ecdsa.PrivateKey, webhookBody string, kid string, issuedAt time.Time) string {
	hasher := sha256.New()
	hasher.Write([]byte(webhookBody))
	bodyHash := hex.EncodeToString(hasher.Sum(nil))

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iat":                 issuedAt.Unix(),
		"request_body_sha256": bodyHash,
	})
	token.Header["kid"] = kid

	tokenString, err := token.SignedString(privateKey)
	suite.Require().NoError(err)
	return tokenString
}

func (suite *IntegrationTestSuite) TestVerifyWebhook_ValidWebhook_Success() {
	h, _ := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}

	webhookBody := `{"webhook_type":"ITEM","webhook_code":"NEW_ITEM","item_id":"test_item"}`

	kid := "test-key-id-123"
	privateKey := suite.getPlaidWebhookPrivateKey(kid)

	issuedAt := clock.Now().Add(-1 * time.Minute)
	tokenString := suite.createPlaidWebhookJWT(privateKey, webhookBody, kid, issuedAt)

	result := ps.VerifyWebhook(webhookBody, tokenString)
	suite.True(result, "Valid webhook should pass verification")
}

func (suite *IntegrationTestSuite) TestVerifyWebhook_MissingHeader_Failure() {
	h, _ := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}

	webhookBody := `{"webhook_type":"ITEM","webhook_code":"NEW_ITEM","item_id":"test_item"}`

	result := ps.VerifyWebhook(webhookBody, "")
	suite.False(result, "Webhook with missing header should fail verification")
}

func (suite *IntegrationTestSuite) TestVerifyWebhook_InvalidJWT_Failure() {
	h, _ := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}

	webhookBody := `{"webhook_type":"ITEM","webhook_code":"NEW_ITEM","item_id":"test_item"}`

	result := ps.VerifyWebhook(webhookBody, "invalid.jwt.token")
	suite.False(result, "Webhook with invalid JWT should fail verification")
}

func (suite *IntegrationTestSuite) TestVerifyWebhook_WrongAlgorithm_Failure() {
	h, _ := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}

	webhookBody := `{"webhook_type":"ITEM","webhook_code":"NEW_ITEM","item_id":"test_item"}`

	// Token with wrong algorithm (HS256 instead of ES256); guard against algorithm confusion attacks
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":                 clock.Now().Unix(),
		"request_body_sha256": "dummy_hash",
	})
	token.Header["kid"] = "test-key-id"

	tokenString, err := token.SignedString([]byte("secret"))
	suite.Require().NoError(err)

	result := ps.VerifyWebhook(webhookBody, tokenString)
	suite.False(result, "Webhook with wrong algorithm should fail verification")
}

func (suite *IntegrationTestSuite) TestVerifyWebhook_ExpiredWebhook_Failure() {
	h, _ := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}

	webhookBody := `{"webhook_type":"ITEM","webhook_code":"NEW_ITEM","item_id":"test_item"}`

	kid := "test-key-id-expired"
	privateKey := suite.getPlaidWebhookPrivateKey(kid)

	issuedAt := clock.Now().Add(-1 * time.Minute)
	tokenString := suite.createPlaidWebhookJWT(privateKey, webhookBody, kid, issuedAt)

	result := ps.VerifyWebhook(webhookBody, tokenString)
	suite.False(result, "Expired webhook should fail verification")
}

func (suite *IntegrationTestSuite) TestVerifyWebhook_ValidJWT_Failure() {
	h, _ := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}

	webhookBody := `{"webhook_type":"ITEM","webhook_code":"NEW_ITEM","item_id":"test_item"}`

	kid := "test-key-id-123"
	privateKey := suite.getPlaidWebhookPrivateKey(kid)

	issuedAt := clock.Now().Add(-6 * time.Minute) // too old
	tokenString := suite.createPlaidWebhookJWT(privateKey, webhookBody, kid, issuedAt)

	result := ps.VerifyWebhook(webhookBody, tokenString)
	suite.False(result, "Expired jwt should fail verification")
}

func (suite *IntegrationTestSuite) TestVerifyWebhook_InvalidHash_Failure() {
	h, _ := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}

	webhookBody := `{"webhook_type":"ITEM","webhook_code":"NEW_ITEM","item_id":"test_item"}`
	differentBody := `{"webhook_type":"ITEM","webhook_code":"DIFFERENT","item_id":"test_item"}`

	kid := "test-key-id-123"
	privateKey := suite.getPlaidWebhookPrivateKey(kid)

	issuedAt := clock.Now().Add(-1 * time.Minute)
	tokenString := suite.createPlaidWebhookJWT(privateKey, differentBody, kid, issuedAt)

	result := ps.VerifyWebhook(webhookBody, tokenString)
	suite.False(result, "Webhook with mismatched body hash should fail verification")
}

func (suite *IntegrationTestSuite) TestVerifyWebhook_MissingKid_Failure() {
	h, _ := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}

	webhookBody := `{"webhook_type":"ITEM","webhook_code":"NEW_ITEM","item_id":"test_item"}`

	privateKey := suite.getPlaidWebhookPrivateKey("test-key-id-123")

	hasher := sha256.New()
	hasher.Write([]byte(webhookBody))
	bodyHash := hex.EncodeToString(hasher.Sum(nil))

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iat":                 clock.Now().Unix(),
		"request_body_sha256": bodyHash,
	})

	tokenString, err := token.SignedString(privateKey)
	suite.Require().NoError(err)

	result := ps.VerifyWebhook(webhookBody, tokenString)
	suite.False(result, "Webhook without kid header should fail verification")
}

func (suite *IntegrationTestSuite) TestVerifyWebhook_MultipleKeys_Success() {
	h, _ := suite.beforePlaid()
	ps := plaid.PlaidService{Logger: logging.Logger, Plaid: h.Plaid, DB: suite.TestDB}

	webhookBody := `{"webhook_type":"ITEM","webhook_code":"NEW_ITEM","item_id":"test_item"}`
	issuedAt := clock.Now().Add(-1 * time.Minute)

	kid1 := "test-key-id-123"
	privateKey1 := suite.getPlaidWebhookPrivateKey(kid1)
	tokenString1 := suite.createPlaidWebhookJWT(privateKey1, webhookBody, kid1, issuedAt)

	result1 := ps.VerifyWebhook(webhookBody, tokenString1)
	suite.True(result1, "First webhook should pass verification")

	kid2 := "test-key-id-456"
	privateKey2 := suite.getPlaidWebhookPrivateKey(kid2)
	tokenString2 := suite.createPlaidWebhookJWT(privateKey2, webhookBody, kid2, issuedAt)

	result2 := ps.VerifyWebhook(webhookBody, tokenString2)
	suite.True(result2, "Second webhook with different key should pass verification")

	tokenString3 := suite.createPlaidWebhookJWT(privateKey1, webhookBody, kid1, issuedAt)
	result3 := ps.VerifyWebhook(webhookBody, tokenString3)
	suite.True(result3, "First webhook should still pass verification (from cache)")
}
