package dao

import (
	"errors"
	"fmt"
	"process-api/pkg/clock"
	"process-api/pkg/db"
	"slices"
	"time"

	"braces.dev/errtrace"
	"github.com/jinzhu/gorm"
	"github.com/plaid/plaid-go/v34/plaid"
)

const (
	CheckingSubtype plaid.AccountSubtype = plaid.ACCOUNTSUBTYPE_CHECKING
	SavingsSubtype  plaid.AccountSubtype = plaid.ACCOUNTSUBTYPE_SAVINGS
)

// A balance is considered stale if more than this amount of time has transpired
// since we've last queried Plaid
const staleBalanceInterval = -24 * time.Hour

type PlaidAccountDao struct {
	ID                    string               `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PlaidAccountID        string               `gorm:"not null;uniqueIndex:plaid_accounts_item_account_unique"`
	UserID                string               `gorm:"foreignKey:UserID;references:Id"`
	PlaidItemID           string               `gorm:"foreignKey:PlaidItemID;references:Id;size:255;not null;uniqueIndex:plaid_accounts_item_account_unique"`
	Name                  string               `gorm:"not null"`
	Subtype               plaid.AccountSubtype `gorm:"not null"`
	BalanceRefreshedAt    *time.Time
	Mask                  *string
	InstitutionName       *string
	InstitutionID         *string
	AvailableBalanceCents *int64
	PrimaryOwnerName      *string
	//  The way that a user links an external account (the auth method) changes  what kind of information we can expect from Plaid about the account,
	// e.g., for the micro-deposits auth flow, Identity (name) and Balance information _can't_ be retrieved:
	// https://support.plaid.com/hc/en-us/articles/14977532310167-Can-I-obtain-Balance-and-Identity-data-for-Items-created-using-Same-Day-Micro-deposits
	// > Since Plaid does not establish a connection to the bank,
	// > the information Plaid is able to provide for Same Day Micro-deposit
	// > accounts is limited.
	// > We are able to provide verified account and routing numbers as well
	// > as the account type (checking or savings).
	// > Plaid is not able to provide Identity or Balance details for the account.
	AuthMethod         *plaid.ItemAuthMethod `gorm:"type:plaid_auth_method"`
	VerificationStatus *string               `gorm:"type:plaid_verification_status"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (PlaidAccountDao) TableName() string { return "plaid_accounts" }

func (PlaidAccountDao) GetAccountForUser(userId, plaidItemId, plaidAccountId string) (*PlaidAccountDao, error) {
	var record PlaidAccountDao
	// The query needs plaid_item_id because Plaid doesn't guarantee plaid_account_id won't collide with other plaid items' account ids
	err := db.DB.Model(PlaidAccountDao{}).Where("user_id = ? AND plaid_item_id = ? AND plaid_account_id = ?", userId, plaidItemId, plaidAccountId).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	return &record, nil
}

func (PlaidAccountDao) GetAccountForUserByID(userId, id string) (*PlaidAccountDao, error) {
	var record PlaidAccountDao
	err := db.DB.Model(PlaidAccountDao{}).Where("id=? AND user_id=?", id, userId).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, errtrace.Wrap(err)
		}
	}
	return &record, nil
}

func (PlaidAccountDao) FindAccountsForUser(userId string) ([]PlaidAccountDao, error) {
	var accounts []PlaidAccountDao

	err := db.DB.Model(&PlaidAccountDao{}).Where("user_id = ?", userId).Order("created_at DESC").Find(&accounts).Error
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("could not find accounts for user %s: %w", userId, err))
	}
	return accounts, nil
}

func (PlaidAccountDao) FindNonErrorAccountsWithStaleBalanceForUser(userId string, now *time.Time) ([]PlaidAccountDao, error) {
	if now == nil {
		tmp := clock.Now()
		now = &tmp
	}
	balanceRefreshedAt := now.Add(staleBalanceInterval)
	var accounts []PlaidAccountDao
	err := db.DB.
		Joins("JOIN plaid_items ON plaid_items.plaid_item_id = plaid_accounts.plaid_item_id").
		Where("plaid_accounts.user_id = ? AND plaid_accounts.balance_refreshed_at < ? AND plaid_items.item_error IS NULL", userId, balanceRefreshedAt).
		Find(&accounts).Error
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("could not find stale accounts for user %s: %w", userId, err))
	}
	return accounts, nil
}

// These come from plaid.LinkDeliveryVerificationStatus
var verifiedStatuses = []string{
	"automatically_verified",
	"manually_verified",
	"database_matched",
}

func (p *PlaidAccountDao) IsVerified() bool {
	// Don't blame me, blame Plaid. `nil` is used to indicate an
	// account has been verified through either Instant Auth or Instant Match
	if p.VerificationStatus == nil {
		return true
	}
	return slices.Contains(verifiedStatuses, *p.VerificationStatus)
}
