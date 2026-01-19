package dao

import (
	"errors"
	"fmt"
	"net/http"
	"process-api/pkg/clock"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/model/response"
	"time"

	"braces.dev/errtrace"
	"github.com/jinzhu/gorm"
)

type UserAccountCardDao struct {
	Id                   int        `gorm:"column:id;primaryKey;autoIncrement"`
	CardHolderId         string     `gorm:"column:card_holder_id" mask:"true"`
	CardId               string     `gorm:"column:card_id" mask:"true"`
	CardMaskNumber       string     `gorm:"column:card_mask_number"`
	AccountId            string     `gorm:"column:account_id" mask:"true"`
	AccountNumber        string     `gorm:"column:account_number" mask:"true"`
	AccountStatus        string     `gorm:"column:account_status"`
	SuspendedAt          *time.Time `gorm:"column:suspended_at"`
	ClosedAt             *time.Time `gorm:"column:closed_at"`
	UserId               string     `gorm:"column:user_id;foreignKey:UserId;references:Id"`
	IsReissue            bool       `gorm:"column:is_reissue" mask:"true"`
	IsReplace            bool       `gorm:"column:is_replace" mask:"true"`
	AccountClosureReason string     `gorm:"column:account_closure_reason" mask:"true"`
	// TODO: Remove this once the Ledger's GetCardDetails API issue is fixed.
	CardExpirationDate     string     `gorm:"column:card_expiration_date" mask:"true"`
	CreatedAt              time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt              time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	ReplaceLockExpiresAt   *time.Time `gorm:"column:replace_lock_expires_at"`
	PreviousCardId         *string    `gorm:"column:previous_card_id" mask:"true"`
	PreviousCardMaskNumber *string    `gorm:"column:previous_card_mask_number"`
}

func (UserAccountCardDao) TableName() string {
	return "user_account_card"
}

func (UserAccountCardDao) FindOneActiveByUserId(db *gorm.DB, userId string) (*UserAccountCardDao, error) {
	var cardHolderRecords UserAccountCardDao
	// Now the user can have multiple accounts, so there may be multiple records with the same userId.
	// Therefore, we search specifically for the ACTIVE account.
	result := db.Where("user_id=? AND account_status=?", userId, "ACTIVE").Find(&cardHolderRecords)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errtrace.Wrap(result.Error)
	}

	return &cardHolderRecords, nil
}

func (UserAccountCardDao) GetAllUserWithActiveStatus(db *gorm.DB) ([]UserAccountCardDao, error) {
	var cardHolderRecords []UserAccountCardDao
	if err := db.Where("account_status = ?", "ACTIVE").Find(&cardHolderRecords).Error; err != nil {
		return nil, errtrace.Wrap(err)
	}
	return cardHolderRecords, nil
}

func (UserAccountCardDao) FindOneByUserId(db *gorm.DB, userId string) (*UserAccountCardDao, error) {
	var cardHolderRecords UserAccountCardDao
	result := db.Where("user_id=?", userId).Find(&cardHolderRecords)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errtrace.Wrap(result.Error)
	}

	return &cardHolderRecords, nil
}

func (UserAccountCardDao) FindOneByAccountNumber(db *gorm.DB, accountNumber string) (*UserAccountCardDao, error) {
	var cardHolderRecords UserAccountCardDao
	result := db.Where("account_number=?", accountNumber).Find(&cardHolderRecords)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errtrace.Wrap(result.Error)
	}

	return &cardHolderRecords, nil
}

func (UserAccountCardDao) FindOneByAccountID(db *gorm.DB, accountID string) (*UserAccountCardDao, error) {
	var record UserAccountCardDao
	result := db.Where("account_id=?", accountID).Find(&record)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errtrace.Wrap(result.Error)
	}

	return &record, nil
}

func (UserAccountCardDao) GetSuspendedAccounts60DaysOld(db *gorm.DB) ([]UserAccountCardDao, error) {
	var userAccountCards []UserAccountCardDao
	if err := db.Where("account_status = ? AND suspended_at::date <= CURRENT_DATE - INTERVAL '60 days'", "SUSPENDED").Find(&userAccountCards).Error; err != nil {
		return nil, errtrace.Wrap(err)
	}
	return userAccountCards, nil
}

func RequireActiveCardHolderForUser(userId string) (*UserAccountCardDao, *response.ErrorResponse) {
	cardHolder, err := UserAccountCardDao{}.FindOneActiveByUserId(db.DB, userId)
	if err != nil {
		return nil, &response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("DB Error: %s", err), MaybeInnerError: errtrace.Wrap(err)}
	}
	if cardHolder == nil {
		var cardHolderRecordsCount int
		result := db.DB.Model(&UserAccountCardDao{}).Where("user_id=?", userId).Count(&cardHolderRecordsCount)
		if result.Error != nil {
			return nil, &response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("DB Error: %s", err), MaybeInnerError: errtrace.Wrap(result.Error)}
		}

		if cardHolderRecordsCount == 0 {
			return nil, &response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, LogMessage: "cardHolder record not found in DB", MaybeInnerError: errtrace.New("")}
		}
		return nil, &response.ErrorResponse{ErrorCode: "ACCOUNT_NOT_ACTIVE", StatusCode: http.StatusPreconditionFailed, LogMessage: fmt.Sprintf("no active account for userID %s", userId), MaybeInnerError: errtrace.New("")}

	}
	if cardHolder.CardId == "" {
		return nil, &response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, LogMessage: "cardId not present", MaybeInnerError: errtrace.New("")}
	}
	return cardHolder, nil
}

func (m *UserAccountCardDao) IsReplaceLocked() bool {
	return m.ReplaceLockExpiresAt != nil && clock.Now().Before(*m.ReplaceLockExpiresAt)
}

// DORMANT is currently treated like SUSPENDED or CLOSED until we get more information
func (uac *UserAccountCardDao) IsActive() bool {
	return uac.AccountStatus == "ACTIVE"
}

func RequireCardHolderForUser(userId string) (*UserAccountCardDao, *response.ErrorResponse) {
	cardHolder, err := UserAccountCardDao{}.FindOneByUserId(db.DB, userId)
	if err != nil {
		return nil, &response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("DB Error: %s", err), MaybeInnerError: errtrace.Wrap(err)}
	}
	if cardHolder == nil {
		return nil, &response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, LogMessage: "cardHolder record not found in DB", MaybeInnerError: errtrace.New("")}
	}
	if cardHolder.CardId == "" {
		return nil, &response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, LogMessage: "cardId not present", MaybeInnerError: errtrace.New("")}
	}
	return cardHolder, nil
}

type UserAccountCardWithUser struct {
	UserId    string `gorm:"column:user_id"`
	AccountId string `gorm:"column:account_id"`
	FirstName string `gorm:"column:first_name"`
	Email     string `gorm:"column:email"`
}

func (UserAccountCardDao) FindUserAccountCardWithUserByAccountIds(accountIds []string) (*[]UserAccountCardWithUser, error) {
	var result []UserAccountCardWithUser
	query := db.DB.Table("user_account_card AS account_record").
		Select("account_record.user_id, account_record.account_id, user_record.first_name, user_record.email").
		Joins("INNER JOIN master_user_records AS user_record ON account_record.user_id = user_record.id").
		Where("account_record.account_id IN (?)", accountIds)
	err := query.Scan(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	return &result, nil
}
