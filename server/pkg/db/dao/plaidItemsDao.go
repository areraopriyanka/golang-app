package dao

import (
	"errors"
	"process-api/pkg/db"
	"time"

	"braces.dev/errtrace"
	"github.com/jinzhu/gorm"
)

type PlaidItemDao struct {
	Id                      uint64    `gorm:"primaryKey"`
	UserId                  string    `gorm:"column:user_id;foreignKey:UserId;references:Id"`
	PlaidItemID             string    `gorm:"column:plaid_item_id;unique"`
	EncryptedAccessToken    []byte    `gorm:"column:encrypted_access_token" mask:"true"`
	KmsEncryptedAccessToken []byte    `gorm:"column:kms_encrypted_access_token" mask:"true"`
	ItemError               *string   `gorm:"column:item_error"`
	IsPendingDisconnect     bool      `gorm:"column:is_pending_disconnect;default:false"`
	CreatedAt               time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt               time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (PlaidItemDao) TableName() string { return "plaid_items" }

func (PlaidItemDao) GetItemForUserByItemID(userId, plaidItemID string) (*PlaidItemDao, error) {
	var record PlaidItemDao
	err := db.DB.Model(PlaidItemDao{}).Where("user_id=? AND plaid_item_id=?", userId, plaidItemID).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, errtrace.Wrap(err)
		}
	}
	return &record, nil
}

func (PlaidItemDao) GetItemForUserIDByPlaidItemID(userID, plaidItemID string) (*PlaidItemDao, error) {
	var record PlaidItemDao
	err := db.DB.Model(PlaidItemDao{}).Where("plaid_item_id=? AND user_id=?", plaidItemID, userID).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errtrace.Wrap(err)
	}
	return &record, nil
}

func (PlaidItemDao) GetItemByPlaidItemID(plaidItemID string) (*PlaidItemDao, error) {
	var record PlaidItemDao
	err := db.DB.Model(PlaidItemDao{}).Where("plaid_item_id=?", plaidItemID).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, errtrace.Wrap(err)
		}
	}
	return &record, nil
}

func (PlaidItemDao) FindFirstItemByPlaidAccountID(plaidAccountID string) (*PlaidItemDao, error) {
	var item PlaidItemDao
	err := db.DB.Model(&PlaidItemDao{}).
		Joins("JOIN plaid_accounts ON plaid_accounts.plaid_item_id = plaid_items.plaid_item_id").
		Where("plaid_accounts.plaid_account_id = ?", plaidAccountID).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, errtrace.Wrap(err)
		}
	}
	return &item, nil
}

func (PlaidItemDao) GetItemsByUserId(userId string) ([]PlaidItemDao, error) {
	var records []PlaidItemDao
	err := db.DB.Model(PlaidItemDao{}).Where("user_id=?", userId).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (PlaidItemDao) FindFirstItemForUserIDByPlaidAccountID(userID, plaidAccountID string) (*PlaidItemDao, error) {
	var item PlaidItemDao
	err := db.DB.Model(&PlaidItemDao{}).
		Joins("JOIN plaid_accounts ON plaid_accounts.plaid_item_id = plaid_items.plaid_item_id").
		Where("plaid_items.user_id = ? AND plaid_accounts.plaid_account_id = ?", userID, plaidAccountID).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, errtrace.Wrap(err)
		}
	}
	return &item, nil
}

func (PlaidItemDao) SetItemError(itemID, itemError string) error {
	var item PlaidItemDao
	err := db.DB.Where("plaid_item_id = ?", itemID).First(&item).Error
	if err != nil {
		return errtrace.Wrap(err)
	}
	return errtrace.Wrap(db.DB.Model(&item).Update("item_error", itemError).Error)
}

func (PlaidItemDao) ClearItemError(itemID string) error {
	var item PlaidItemDao
	err := db.DB.Where("plaid_item_id = ?", itemID).First(&item).Error
	if err != nil {
		return errtrace.Wrap(err)
	}
	return errtrace.Wrap(db.DB.Model(&item).Update("item_error", nil).Error)
}

func (PlaidItemDao) SetIsPendingDisconnect(itemID string, isPending bool) error {
	var item PlaidItemDao
	err := db.DB.Where("plaid_item_id = ?", itemID).First(&item).Error
	if err != nil {
		return errtrace.Wrap(err)
	}
	return errtrace.Wrap(db.DB.Model(&item).Update("is_pending_disconnect", isPending).Error)
}

func (PlaidItemDao) FindAll() ([]PlaidItemDao, error) {
	var records []PlaidItemDao
	err := db.DB.Find(&records).Error
	return records, errtrace.Wrap(err)
}
