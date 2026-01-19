package plaid

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/utils"
	"time"

	"braces.dev/errtrace"
	"github.com/jinzhu/gorm"
	"github.com/plaid/plaid-go/v34/plaid"
)

func NewPlaid(config *config.Configs) *plaid.APIClient {
	cfg := plaid.NewConfiguration()
	cfg.AddDefaultHeader("PLAID-CLIENT-ID", config.Plaid.ClientId)
	cfg.AddDefaultHeader("PLAID-SECRET", config.Plaid.Secret)
	cfg.UseEnvironment(plaid.Environment(config.Plaid.Environment))
	return plaid.NewAPIClient(cfg)
}

type PlaidService struct {
	Plaid      *plaid.APIClient
	Logger     *slog.Logger
	DB         *gorm.DB
	WebhookURL string
}

type ItemPublicTokenExchangeResponse struct {
	PlaidItemId string
	AccessToken string
}

func (ps *PlaidService) LinkTokenCreateRequest(userId, platform, redirectURI, env string, accessToken *string) (string, error) {
	plaidUser := plaid.LinkTokenCreateRequestUser{
		ClientUserId: userId,
	}
	request := plaid.NewLinkTokenCreateRequest(
		"DreamFi",
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_US},
		plaidUser,
	)

	// Update mode if accessToken is provided
	if accessToken != nil {
		request.SetAccessToken(*accessToken)
		request.SetUpdate(plaid.LinkTokenCreateRequestUpdate{})
		// If, in the future, we want to allow update mode to add newly available accounts,
		// we'll need to set `request.Update.AccountSelectionEnabled` to `true`, and will
		// need to set the same acount filters `request.SetAccountFilters` we set in the
		// non-update branch (see below).
		// This feature was not enabled for the original Plaid Link update work since it
		// also allows the user to disconnect their previously connected account; that
		// will need some deliberation and planning. Noted in DT-1255
	} else {
		auth := plaid.NewLinkTokenCreateRequestAuth()
		auth.SetAutomatedMicrodepositsEnabled(true)
		request.SetAuth(*auth)
		request.SetProducts([]plaid.Products{plaid.PRODUCTS_AUTH, plaid.PRODUCTS_IDENTITY})
		if ps.WebhookURL != "" {
			request.SetWebhook(ps.WebhookURL)
		}
		request.SetAccountFilters(plaid.LinkTokenAccountFilters{
			Depository: &plaid.DepositoryFilter{
				AccountSubtypes: []plaid.DepositoryAccountSubtype{plaid.DEPOSITORYACCOUNTSUBTYPE_CHECKING, plaid.DEPOSITORYACCOUNTSUBTYPE_SAVINGS},
			},
		})
	}

	request.SetLinkCustomizationName("default")
	// From Plaid docs:
	// > The name of your app's Android package. Required if using the `link_token` to initialize Link on Android.
	// > Any package name specified here must also be added to the Allowed Android package names setting on the [developer dashboard](https://dashboard.plaid.com/team/api).
	// > When creating a `link_token` for initializing Link on other platforms, `android_package_name` must be left blank and `redirect_uri` should be used instead.
	if platform == "android" {
		if env == constant.PROD {
			request.SetAndroidPackageName("com.dreamfi.mobileapp")
		} else {
			request.SetAndroidPackageName("com.dreamfi.mobileapp.dev")
		}
	} else {
		request.SetRedirectUri(redirectURI)
	}
	ctx := context.Background()
	resp, _, err := ps.Plaid.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		plaidErr, plaidErrErr := plaid.ToPlaidError(err)
		if plaidErrErr != nil {
			ps.Logger.Error("PlaidApi.LinkTokenCreate failed", "error", err.Error(), "call to plaid.ToPlaidError failed", plaidErrErr.Error())
		} else {
			ps.Logger.Error("PlaidApi.LinkTokenCreate failed", "errorMessage", plaidErr.ErrorMessage, "ErrorCode", plaidErr.ErrorCode, "ErrorCodeReason", plaidErr.ErrorCodeReason, "Causes", plaidErr.Causes)
		}
		return "", errtrace.Wrap(err)
	}
	return resp.GetLinkToken(), nil
}

func (ps *PlaidService) ItemPublicTokenExchangeRequest(publicToken string) (*ItemPublicTokenExchangeResponse, error) {
	logger := ps.Logger.WithGroup("PlaidExchangePublicToken")
	ctx := context.Background()
	exchangePublicTokenReq := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	exchangePublicTokenResp, _, err := ps.Plaid.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(
		*exchangePublicTokenReq,
	).Execute()
	if err != nil {
		plaidErr, plaidErrErr := plaid.ToPlaidError(err)
		if plaidErrErr != nil {
			logger.Error("PlaidApi.ItemPublicTokenExchange failed", "error", err.Error(), "call to plaid.ToPlaidError failed", plaidErrErr.Error())
		} else {
			logger.Error("PlaidApi.ItemPublicTokenExchange failed", "errorMessage", plaidErr.ErrorMessage, "ErrorCode", plaidErr.ErrorCode, "ErrorCodeReason", plaidErr.ErrorCodeReason, "Causes", plaidErr.Causes)
		}
		return nil, errtrace.Wrap(err)
	}
	plaidItemId := exchangePublicTokenResp.GetItemId()
	accessToken := exchangePublicTokenResp.GetAccessToken()
	logger.Info("successfully exchanged public token for access token", "plaidItemId", plaidItemId)
	return &ItemPublicTokenExchangeResponse{PlaidItemId: plaidItemId, AccessToken: accessToken}, nil
}

func (ps *PlaidService) InsertItem(userId, plaidItemId, accessToken string) error {
	logger := ps.Logger.WithGroup("InsertItem").With("userId", userId, "plaidItemId", plaidItemId)
	encryptedAccessToken, err := utils.EncryptKmsBinary(accessToken)
	if err != nil {
		logger.Error("error encrypting accessToken", "error", err.Error())
		return errtrace.Wrap(err)
	}
	item := dao.PlaidItemDao{
		UserId:                  userId,
		PlaidItemID:             plaidItemId,
		KmsEncryptedAccessToken: []byte(encryptedAccessToken),
	}

	if err := ps.DB.
		Select("user_id", "plaid_item_id", "kms_encrypted_access_token").
		Create(&item).Error; err != nil {
		logger.Error("error inserting Plaid item record", "error", err.Error())
		return errtrace.Wrap(err)
	}
	return nil
}

type PlaidBalanceUpdateStatus int

const (
	PlaidBalanceOmitted PlaidBalanceUpdateStatus = iota
	PlaidBalanceUpdated
)

var PlaidBalanceInitialized = PlaidBalanceUpdated

// Plaid's `accounts/balance/get` can be slow (usually < 10s; sometimes > 30 s) and costly;
// right after a user links an account, `/accounts/get` has fresh balance info. Use this first.
// [source 1](https://support.plaid.com/hc/en-us/articles/16179450406295-How-can-I-obtain-account-balances-using-Plaid-s-API)
// [source 2](https://support.plaid.com/hc/en-us/articles/14977441562647-Why-is-an-Item-returning-an-ITEM-LOGIN-REQUIRED-error-soon-after-creation)
func (rec *PlaidService) InitialAccountsGetRequest(userId, plaidItemId, accessToken string) error {
	logger := rec.Logger.WithGroup("InitialAccountsGetRequest").With("userId", userId, "plaidItemId", plaidItemId)
	ctx := context.Background()
	resp, _, err := rec.Plaid.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
		*plaid.NewAccountsGetRequest(accessToken),
	).Execute()
	if err != nil {
		logger.Error("plaid /accounts/get call failed", "error", err.Error())
		rec.handlePlaidErrorSideEffects(err, plaidItemId)
		return errtrace.Wrap(err)
	}
	logger.Debug("success")
	return rec.insertAccounts(userId, plaidItemId, resp)
}

func (rec *PlaidService) AccountsBalanceGetRequest(userId, plaidItemId, accessToken string) error {
	logger := rec.Logger.WithGroup("AccountsBalanceGet").With("userId", userId, "plaidItemId", plaidItemId)
	ctx := context.Background()
	resp, _, err := rec.Plaid.PlaidApi.AccountsBalanceGet(ctx).AccountsBalanceGetRequest(
		*plaid.NewAccountsBalanceGetRequest(accessToken),
	).Execute()
	if err != nil {
		logger.Error("plaid /accounts/balance/get call failed", "error", err.Error())
		rec.handlePlaidErrorSideEffects(err, plaidItemId)
		return errtrace.Wrap(err)
	}
	logger.Debug("success")
	return rec.updateAccounts(userId, plaidItemId, resp, PlaidBalanceUpdated)
}

func extractItemDetails(item plaid.Item) (*string, *string, *plaid.ItemAuthMethod) {
	// Plaid returns only a single institution regardless of the number of accounts returned
	// From: https://plaid.com/docs/api/institutions/#institutions-get_by_id-response-institution-institution-id
	// > Note that the same institution may have multiple records, each with different institution IDs; for example,
	// > if the institution has migrated to OAuth, there may be separate institution_ids for the OAuth and non-OAuth versions of the institution.
	// > Institutions that operate in different countries or with multiple login portals may also have separate institution_ids for each country
	// > or portal.
	// From: https://plaid.com/docs/api/products/balance/#accounts-balance-get-response-item-institution-id
	// > Field is null for Items created without an institution connection, such as Items created via Same Day Micro-deposits.
	institutionID := item.InstitutionId.Get()
	institutionName := item.InstitutionName.Get()
	authMethod := item.AuthMethod.Get()
	return institutionID, institutionName, authMethod
}

func (rec *PlaidService) insertAccounts(userId string, plaidItemId string, resp plaid.AccountsGetResponse) error {
	institutionID, institutionName, authMethod := extractItemDetails(resp.Item)
	var errs []error
	for _, account := range resp.Accounts {
		err := rec.insertAccount(account, userId, plaidItemId, institutionID, institutionName, authMethod)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (rec *PlaidService) updateAccounts(userId string, plaidItemId string, resp plaid.AccountsGetResponse, balanceStatus PlaidBalanceUpdateStatus) error {
	institutionID, institutionName, authMethod := extractItemDetails(resp.Item)
	var errs []error
	for _, account := range resp.Accounts {
		err := rec.updateAccount(account, userId, plaidItemId, institutionID, institutionName, authMethod, balanceStatus)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (rec *PlaidService) insertAccount(account plaid.AccountBase, userId, plaidItemId string, institutionID, institutionName *string, authMethod *plaid.ItemAuthMethod) error {
	logger := rec.Logger.WithGroup("insertAccount").With("userId", userId, "plaidItemId", plaidItemId)
	subtype, err := validateSubtype(account.Subtype.Get())
	if err != nil {
		logger.Error("invalid subtype for account; must be checking or savings", "error", err.Error())
		return errtrace.Wrap(fmt.Errorf("invalid subtype for account; must be checking or savings %w", err))
	}

	// BalanceRefreshedAt should only be set for instant auth methods
	var balanceRefreshedAt *time.Time
	if authMethod != nil && (*authMethod == plaid.ITEMAUTHMETHOD_INSTANT_AUTH || *authMethod == plaid.ITEMAUTHMETHOD_INSTANT_MATCH) {
		now := clock.Now()
		balanceRefreshedAt = &now
	}

	record := dao.PlaidAccountDao{
		PlaidAccountID:        account.AccountId,
		UserID:                userId,
		PlaidItemID:           plaidItemId,
		Name:                  account.Name,
		Subtype:               subtype,
		Mask:                  account.Mask.Get(),
		InstitutionID:         institutionID,
		InstitutionName:       institutionName,
		AuthMethod:            authMethod,
		AvailableBalanceCents: utils.NilableUSDtoCents(account.Balances.Available.Get()),
		BalanceRefreshedAt:    balanceRefreshedAt,
		// From Plaid docs:
		// The account holder name that was used for micro-deposit and/or database verification.
		// Only returned for Auth Items created via micro-deposit or database verification.
		// This name was manually-entered by the user during Link, unless it was otherwise provided via the `user.legal_name` request field in `/link/token/create` for the Link session that created the Item.
		PrimaryOwnerName:   account.VerificationName,
		VerificationStatus: account.VerificationStatus,
	}
	err = rec.DB.Create(&record).Error
	if err != nil {
		logger.Error("couldn't create plaid account record", "error", err.Error())
		return errtrace.Wrap(err)
	}
	logger.Debug("success")
	return nil
}

func (rec *PlaidService) updateAccount(account plaid.AccountBase, userId, plaidItemId string, institutionID, institutionName *string, authMethod *plaid.ItemAuthMethod, balanceStatus PlaidBalanceUpdateStatus) error {
	logger := rec.Logger.WithGroup("updateAccount").With("userId", userId, "plaidItemId", plaidItemId)
	record, err := dao.PlaidAccountDao{}.GetAccountForUser(userId, plaidItemId, account.AccountId)
	if err != nil {
		logger.Error("failed to update account", "plaidAccountId", account.AccountId, "error", err.Error())
		return errtrace.Wrap(fmt.Errorf("failed to update account; %w", err))
	} else if record == nil {
		logger.Error("failed to update account", "plaidAccountId", account.AccountId, "error", "account record not found")
		return errors.New("failed to update account; account record not found")
	}
	// The balance should only be considered up-to-date if this was called as the result
	// of a accounts/balance/get request, _or_ if it's the first accounts/get request made
	// after the account was connected via Plaid Link
	balanceRefreshedAt := record.BalanceRefreshedAt
	if balanceStatus == PlaidBalanceUpdated {
		now := clock.Now()
		balanceRefreshedAt = &now
	} else if balanceStatus != PlaidBalanceOmitted { // exhausting :)
		logger.Error("balance status is not valid", "balanceStatus", balanceStatus)
		return errors.New("balance status is not valid")
	}
	err = rec.DB.
		Model(&record).Updates(map[string]any{
		"name":                    account.Name,
		"mask":                    account.Mask.Get(),
		"institution_id":          institutionID,
		"institution_name":        institutionName,
		"auth_method":             authMethod,
		"available_balance_cents": utils.NilableUSDtoCents(account.Balances.Available.Get()),
		"balance_refreshed_at":    balanceRefreshedAt,
		"verification_status":     account.VerificationStatus,
	}).Error
	if err != nil {
		logger.Error("couldn't update plaid account record", "error", err.Error())
		return errtrace.Wrap(err)
	}
	return nil
}

func validateSubtype(subtype *plaid.AccountSubtype) (plaid.AccountSubtype, error) {
	if subtype == nil {
		return "", errtrace.Wrap(errors.New("subtype is nil"))
	}
	if plaid.AccountSubtype(*subtype) == dao.CheckingSubtype {
		return dao.CheckingSubtype, nil
	}
	if plaid.AccountSubtype(*subtype) == dao.SavingsSubtype {
		return dao.SavingsSubtype, nil
	}
	return "", errtrace.Wrap(errors.New("invalid subtype"))
}

func PlaidSubtypeToLedgerIdentificationType2(subtype plaid.AccountSubtype) (ledger.IdentificationType2, error) {
	switch subtype {
	case plaid.ACCOUNTSUBTYPE_SAVINGS:
		return ledger.IdentificationType2_SAVINGS, nil
	case plaid.ACCOUNTSUBTYPE_CHECKING:
		return ledger.IdentificationType2_CHECKING, nil
	default:
		return "", errtrace.Wrap(fmt.Errorf("unsupported plaid subtype for ledger identificationType2: %s", subtype))
	}
}

type PlaidACHDetails struct {
	Subtype plaid.AccountSubtype
	Account string
	Routing string
}

var ErrProductNotReady = errors.New("PRODUCT_NOT_READY")

func (ps *PlaidService) GetACHDetails(accessToken, plaidAccountID string) (*PlaidACHDetails, error) {
	logger := ps.Logger.WithGroup("GetACHDetails").With("plaidAccountID", plaidAccountID)
	ctx := context.Background()
	resp, _, err := ps.Plaid.PlaidApi.AuthGet(ctx).AuthGetRequest(*plaid.NewAuthGetRequest(accessToken)).Execute()
	if err != nil {
		plaidErr, plaidErrErr := plaid.ToPlaidError(errtrace.Wrap(err))
		if plaidErrErr != nil {
			logger.Error("PlaidApi.AuthGet failed", "error", err.Error(), "call to plaid.ToPlaidError failed", plaidErrErr.Error())
		} else {
			logger.Error("PlaidApi.AuthGet failed", "errorMessage", plaidErr.ErrorMessage, "ErrorCode", plaidErr.ErrorCode, "ErrorCodeReason", plaidErr.ErrorCodeReason, "Causes", plaidErr.Causes)
			if plaidErr.ErrorCode == "PRODUCT_NOT_READY" {
				return nil, errtrace.Wrap(ErrProductNotReady)
			}
		}
		return nil, errtrace.Wrap(err)
	}
	logger.Debug("PlaidApi.AuthGet success")
	account := getAccountFromSlice(resp.Accounts, plaidAccountID)
	if account == nil {
		return nil, errtrace.Wrap(fmt.Errorf("GetACHDetails failed; could not find account for plaidAccountID: %s", plaidAccountID))
	}
	subtype, err := validateSubtype((*account).Subtype.Get())
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("GetACHDetails failed; %w", err))
	}
	numbersACH := getNumbersACHFromSlice(resp.Numbers.Ach, plaidAccountID)
	if numbersACH == nil {
		return nil, errtrace.Wrap(fmt.Errorf("GetACHDetails failed; could not find ach numbers for plaidAccountID: %s", plaidAccountID))
	}
	details := PlaidACHDetails{
		Subtype: subtype,
		Account: numbersACH.Account,
		Routing: numbersACH.Routing,
	}
	logger.Debug("successfully retrieved ACH details")
	return &details, nil
}

func getAccountFromSlice(accounts []plaid.AccountBase, plaidAccountID string) *plaid.AccountBase {
	for _, account := range accounts {
		if account.AccountId == plaidAccountID {
			return &account
		}
	}
	return nil
}

func getNumbersACHFromSlice(achs []plaid.NumbersACH, plaidAccountID string) *plaid.NumbersACH {
	for _, numbersACH := range achs {
		if numbersACH.AccountId == plaidAccountID {
			return &numbersACH
		}
	}
	return nil
}

// This function is currently purposefully limited in scope to accomodate
// DT-1144 - the ledger... _might_ need at least the external account's first(?)
// name for ACH transfers. Plaid's identity endpoint provides much more information,
// but for the time being, we're just grabbing the intial name entry in the returned
// accounts[x].owners[x] list.
func (ps *PlaidService) GetIdentity(userId, plaidItemId, accessToken string) error {
	logger := ps.Logger.WithGroup("GetIdentity").With("userId", userId, "plaidItemId", plaidItemId)
	ctx := context.Background()
	resp, _, err := ps.Plaid.PlaidApi.IdentityGet(ctx).IdentityGetRequest(*plaid.NewIdentityGetRequest(accessToken)).Execute()
	if err != nil {
		plaidErr, plaidErrErr := plaid.ToPlaidError(err)
		if plaidErrErr != nil {
			logger.Error("PlaidApi.IdentityGet failed", "error", err.Error(), "call to plaid.ToPlaidError failed", plaidErrErr.Error())
		} else {
			logger.Error("PlaidApi.IdentityGet failed", "errorMessage", plaidErr.ErrorMessage, "ErrorCode", plaidErr.ErrorCode, "ErrorCodeReason", plaidErr.ErrorCodeReason, "Causes", plaidErr.Causes)
		}
		return errtrace.Wrap(err)
	}
	logger.Debug("PlaidApi.IdentityGet success")

	var errs []error
	for _, account := range resp.Accounts {
		var primaryOwnerName *string
		if len(account.Owners) > 0 && len(account.Owners[0].Names) > 0 {
			primaryOwnerName = &account.Owners[0].Names[0]
		}

		if primaryOwnerName != nil {
			record, err := dao.PlaidAccountDao{}.GetAccountForUser(userId, plaidItemId, account.AccountId)
			if err != nil {
				logger.Error("failed to get account for user", "plaidAccountId", account.AccountId, "error", err.Error())
				errs = append(errs, errtrace.Wrap(err))
			} else if record == nil {
				logger.Error("failed to get account for user", "plaidAccountId", account.AccountId, "error", "record not found")
				errs = append(errs, errtrace.Wrap(fmt.Errorf("failed to update primary owner name; could not find account record")))
			} else {
				err = ps.DB.Model(&record).Update("primary_owner_name", primaryOwnerName).Error
				if err != nil {
					logger.Error("failed to update primary owner name", "plaidAccountId", account.AccountId, "error", err.Error())
					errs = append(errs, errtrace.Wrap(err))
				} else {
					logger.Debug("updated primary owner name", "plaidAccountId", account.AccountId, "primaryOwnerName", *primaryOwnerName)
				}
			}
		}
	}

	return errors.Join(errs...)
}

func (ps *PlaidService) CheckForDuplicateAccounts(userID string, institutionID *string, accounts *[]PlaidLinkAccount) (bool, error) {
	var errs []error
	hasDuplicates := false
	for _, account := range *accounts {
		isDuplicate, err := ps.IsDuplicateAccount(userID, institutionID, &account)
		if isDuplicate {
			ps.Logger.Info("duplicate account found", "userID", userID, "institutionID", institutionID, "account", account)
			hasDuplicates = true
		}
		if err != nil {
			errs = append(errs, err)
		}
	}
	return hasDuplicates, errors.Join(errs...)
}

func (ps *PlaidService) IsDuplicateAccount(userID string, institutionID *string, account *PlaidLinkAccount) (bool, error) {
	var record dao.PlaidAccountDao
	var err error
	// https://plaid.com/docs/link/duplicate-items/#identifying-existing-duplicate-items
	// Occasionally, the mask or name fields may be null, in which case you can compare institution_id and client_user_id as a fallback.
	// If a user has linked to that institution, assume a duplicate, though this isn't always the case.
	// In the future, the user will be able to add new accounts even after they're linked to an institution (see: DT-1031)
	if account.Name == nil || account.Mask == nil {
		err = ps.DB.Model(dao.PlaidAccountDao{}).Where("user_id=? AND institution_id=?", userID, institutionID).Take(&record).Error
	} else {
		err = ps.DB.Model(dao.PlaidAccountDao{}).Where("user_id=? AND name=? AND mask=? AND institution_id=?", userID, account.Name, account.Mask, institutionID).Take(&record).Error
	}
	if err == nil { // No error means that an account was found, which means it's a duplicate
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	ps.Logger.Error("error querying for IsDuplicateAccount", "error", err.Error())
	return false, errtrace.Wrap(err)
}

type PlaidLinkAccount struct {
	ID                 string  `json:"id" validate:"required"`
	Name               *string `json:"name,omitempty"`
	Mask               *string `json:"mask,omitempty"`
	Type               string  `json:"type" validate:"required,eq=depository"`
	Subtype            string  `json:"subtype" validate:"required,oneof=checking savings"`
	VerificationStatus *string `json:"verificationStatus,omitempty"`
}

type PlaidLinkOnSuccess struct {
	// From: https://plaid.com/docs/link/web/#link-web-onsuccess-metadata-institution
	// > If the Item was created via Same-Day micro-deposit verification, [... InstitutionID ...] will be null.
	InstitutionID *string            `json:"institutionId,omitempty"`
	Accounts      []PlaidLinkAccount `json:"accounts" validate:"required"`
	LinkSessionID string             `json:"linkSessionId" validate:"required"`
}

func (ps *PlaidService) UnlinkItem(userId, plaidItemId, accessToken string) error {
	logger := ps.Logger.WithGroup("UnlinkItem").With("userId", userId, "plaidItemId", plaidItemId)
	ctx := context.Background()
	removeRequest := plaid.NewItemRemoveRequest(accessToken)
	if removeRequest == nil {
		logger.Error("plaid.NewItemRemoveRequest returned a nil pointer")
		return errtrace.Wrap(errors.New("plaid.NewItemRemoveRequest returned a nil pointer"))
	}
	_, _, err := ps.Plaid.PlaidApi.ItemRemove(ctx).ItemRemoveRequest(*removeRequest).Execute()
	if err != nil {
		plaidErr, plaidErrErr := plaid.ToPlaidError(err)
		if plaidErrErr != nil {
			logger.Error("PlaidApi.ItemRemove failed", "error", err.Error(), "call to plaid.ToPlaidError failed", plaidErrErr.Error())
		} else {
			logger.Error("PlaidApi.ItemRemove failed", "errorMessage", plaidErr.ErrorMessage, "ErrorCode", plaidErr.ErrorCode, "ErrorCodeReason", plaidErr.ErrorCodeReason, "Causes", plaidErr.Causes)
		}
		return errtrace.Wrap(err)
	}

	err = ps.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("plaid_item_id=?", plaidItemId).Delete(&dao.PlaidAccountDao{}).Error
		if err != nil {
			logger.Error("error deleting plaid accounts", "error", err.Error())
			return errtrace.Wrap(err)
		}
		err = tx.Where("plaid_item_id=?", plaidItemId).Delete(&dao.PlaidItemDao{}).Error
		if err != nil {
			logger.Error("error deleting plaid item", "error", err.Error())
			return errtrace.Wrap(err)
		}
		return nil
	})
	if err != nil {
		logger.Error("transaction failed", "error", err.Error())
		return errtrace.Wrap(err)
	}

	logger.Debug("successfully unlinked plaid item")
	return nil
}

func (ps *PlaidService) ItemWebhookUpdateRequest(accessToken, webhookURL string) error {
	logger := ps.Logger.WithGroup("ItemWebhookUpdateRequest").With("webhookURL", webhookURL)
	ctx := context.Background()

	request := plaid.NewItemWebhookUpdateRequest(accessToken)
	if request == nil {
		logger.Error("NewItemWebhookUpdateRequest returned nil")
		return errors.New("NewItemWebhookUpdateRequest returned nil")
	}
	request.SetWebhook(webhookURL)
	_, _, err := ps.Plaid.PlaidApi.ItemWebhookUpdate(ctx).ItemWebhookUpdateRequest(*request).Execute()
	if err != nil {
		plaidErr, plaidErrErr := plaid.ToPlaidError(err)
		if plaidErrErr != nil {
			logger.Error("PlaidApi.ItemWebhookUpdate failed", "error", err.Error(), "call to plaid.ToPlaidError failed", plaidErrErr.Error())
		} else {
			logger.Error("PlaidApi.ItemWebhookUpdate failed", "errorMessage", plaidErr.ErrorMessage, "ErrorCode", plaidErr.ErrorCode, "ErrorCodeReason", plaidErr.ErrorCodeReason, "Causes", plaidErr.Causes)
		}
		return errtrace.Wrap(err)
	}

	logger.Debug("Successfully updated webhook for item")
	return nil
}

func (ps *PlaidService) handlePlaidErrorSideEffects(err error, plaidItemId string) {
	logger := ps.Logger.WithGroup("handlePlaidErrorSideEffects").With("plaidItemId", plaidItemId)

	plaidErr, plaidErrErr := plaid.ToPlaidError(err)
	if plaidErrErr != nil {
		logger.Error("error converting err to PlaidError", "original error", err.Error(), "call to plaid.ToPlaidError failed", plaidErrErr.Error())
		return
	}

	logger.Error("plaid API call failed", "errorMessage", plaidErr.ErrorMessage, "ErrorCode", plaidErr.ErrorCode, "ErrorCodeReason", plaidErr.ErrorCodeReason, "Causes", plaidErr.Causes)

	if plaidErr.ErrorCode != "ITEM_LOGIN_REQUIRED" {
		return
	}

	logger.Debug("ITEM_LOGIN_REQUIRED error")
	err = dao.PlaidItemDao{}.SetItemError(plaidItemId, plaidErr.ErrorCode)
	if err != nil {
		logger.Error("Failed to mark Plaid item as ITEM_LOGIN_REQUIRED", "error", err.Error())
	} else {
		logger.Debug("Marked item for update mode")
	}
}

type WebhookPayloadError struct {
	ErrorType    string `json:"error_type"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}
type WebhookPayload struct {
	WebhookType string               `json:"webhook_type"`
	WebhookCode string               `json:"webhook_code"`
	ItemID      string               `json:"item_id"`
	AccountID   *string              `json:"account_id,omitempty"`
	Error       *WebhookPayloadError `json:"error,omitempty"`
}

// TODO: [Use river](https://dreamfi.atlassian.net/browse/DT-1365) - plaid docs:
// It's best to keep your receiver as simple as possible, such as a receiver whose
// only job is to write the webhook into a queue or reliable storage. This is
// important for two reasons. First, if the receiver does not respond within 10
// seconds, the delivery is considered failed. Second, because webhooks can arrive
// at unpredictable rates. Therefore if you do a lot of work in your receiver -
// e.g. generating and sending an email - spikes are likely to overwhelm your
// downstream services, or cause you to be rate-limited if the downstream is a
// third-party.
func (ps *PlaidService) HandleItemWebhook(webhookCode, itemID string, webhookError *WebhookPayloadError) error {
	switch webhookCode {
	case "ERROR":
		itemError := webhookCode
		if webhookError != nil {
			itemError = webhookError.ErrorCode
		}
		return dao.PlaidItemDao{}.SetItemError(itemID, itemError)
	case "LOGIN_REPAIRED":
		var errs []error
		err := dao.PlaidItemDao{}.ClearItemError(itemID)
		if err != nil {
			errs = append(errs, err)
		}
		err = dao.PlaidItemDao{}.SetIsPendingDisconnect(itemID, false)
		if err != nil {
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	case "USER_ACCOUNT_REVOKED", "USER_PERMISSION_REVOKED":
		return ps.revokeItem(itemID)
	case "PENDING_EXPIRATION", "PENDING_DISCONNECT":
		return dao.PlaidItemDao{}.SetIsPendingDisconnect(itemID, true)
	}
	return nil
}

func (ps *PlaidService) revokeItem(itemID string) error {
	item, err := dao.PlaidItemDao{}.GetItemByPlaidItemID(itemID)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("revokeItem: error retrieving item %s: %w", itemID, err))
	}
	if item == nil {
		return errtrace.Wrap(fmt.Errorf("revokeItem: item not found %s", itemID))
	}

	accessToken, err := utils.DecryptPlaidAccessToken(item.EncryptedAccessToken, item.KmsEncryptedAccessToken)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("revokeItem: could not decrypt access token for plaid_item_id %s: %w", itemID, err))
	}

	return ps.UnlinkItem(item.UserId, itemID, accessToken)
}

func (ps *PlaidService) HandleAuthWebhook(webhookCode, plaidAccountID string) error {
	switch webhookCode {
	case "AUTOMATICALLY_VERIFIED", "VERIFICATION_EXPIRED":
		balanceStatus := PlaidBalanceOmitted
		// The first call after an automatically_verified verification status update means
		// that micro-deposits were verified by Plaid, and the balance from an accounts/get
		// call will be up-to-date
		if webhookCode == "AUTOMATICALLY_VERIFIED" {
			balanceStatus = PlaidBalanceUpdated
		}
		return ps.updateAccountsAssociatedWithPlaidAccountID(plaidAccountID, balanceStatus)
	}
	return nil
}

// Multiple accounts (e.g., checking + savings) can be connected via a single Plaid item (the result of the Plaid Link process).
// If we know that the status of one of the accounts changed, it's likely that the status of any other
// accounts associated through the item have changed. This fn allows us to update everything at once.
func (rec *PlaidService) updateAccountsAssociatedWithPlaidAccountID(plaidAccountID string, balanceStatus PlaidBalanceUpdateStatus) error {
	logger := rec.Logger.WithGroup("updateAccountsAssociatedWithPlaidAccountID").With("plaidAccountID", plaidAccountID)
	item, err := dao.PlaidItemDao{}.FindFirstItemByPlaidAccountID(plaidAccountID)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("error querying for associated plaid_item given plaid_account_id %s: %w", plaidAccountID, err))
	}
	if item == nil {
		return errtrace.Wrap(fmt.Errorf("could not find for associated plaid_item given plaid_account_id %s", plaidAccountID))
	}
	plaidItemID := item.PlaidItemID
	accessToken, err := utils.DecryptPlaidAccessToken(item.EncryptedAccessToken, item.KmsEncryptedAccessToken)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("could not decrypt access token for plaid_item_id %s: %w", plaidItemID, err))
	}

	ctx := context.Background()

	resp, _, err := rec.Plaid.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
		*plaid.NewAccountsGetRequest(accessToken),
	).Execute()
	if err != nil {
		logger.Error("plaid /accounts/get call failed", "error", err.Error())
		rec.handlePlaidErrorSideEffects(err, plaidItemID)
		return errtrace.Wrap(err)
	}
	logger.Debug("success")
	return rec.updateAccounts(item.UserId, plaidItemID, resp, balanceStatus)
}
