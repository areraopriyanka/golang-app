package dao

import (
	"errors"
	"process-api/pkg/db"
	"strings"
	"time"

	"braces.dev/errtrace"
	"github.com/jinzhu/gorm"
)

type TransactionDisputeDao struct {
	Id                           string `gorm:"column:id;primaryKey"`
	Status                       string
	TransactionIdentifier        string `gorm:"column:transaction_identifier"`
	Reason                       string
	Details                      string
	ExtraInfo                    string     `gorm:"column:extra_info"`
	UserId                       string     `gorm:"column:user_id;foreignKey:UserId;references:Id"`
	CreatedAt                    time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt                    time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	ProvisionalCreditTransaction *string    `gorm:"column:provisional_credit_transaction"`
	VoidCreditTransaction        *string    `gorm:"column:void_credit_transaction"`
	CreditedAt                   *time.Time `gorm:"column:credited_at"`
	VoidedAt                     *time.Time `gorm:"column:voided_at"`
}

func (TransactionDisputeDao) TableName() string {
	return "transaction_disputes"
}

type SubmitTransactionDisputeRequest struct {
	Reason  string `json:"reason" validate:"required,oneof='Incorrect amount charged' 'Duplicate transaction' 'Goods not received' 'Goods not as described' 'Billing error' 'Unauthorized transaction' 'Identity theft' 'Card skimming' 'Online fraud'"`
	Details string `json:"details"`
}

func (TransactionDisputeDao) FindByUserIdAndReferenceId(userId string, referenceId string) (*TransactionDisputeDao, error) {
	var transactionDisputeRecord TransactionDisputeDao
	result := db.DB.Where("user_id =? AND transaction_identifier = ?", userId, referenceId).First(&transactionDisputeRecord)
	// record not found
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, errtrace.Wrap(result.Error)
	}
	return &transactionDisputeRecord, nil
}

func (TransactionDisputeDao) FindByUserIdAndReferenceIds(userId string, referenceIds []string) (*[]TransactionDisputeDao, error) {
	var records []TransactionDisputeDao
	result := db.DB.Where("user_id = ? AND transaction_identifier IN (?)", userId, referenceIds).Find(&records)
	if result.Error != nil {
		return nil, errtrace.Wrap(result.Error)
	}
	return &records, nil
}

func (TransactionDisputeDao) FindById(disputeId string) (*TransactionDisputeDao, error) {
	var transactionDisputeRecord TransactionDisputeDao
	result := db.DB.Where("id = ?", disputeId).Take(&transactionDisputeRecord)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, errtrace.Wrap(result.Error)
	}
	return &transactionDisputeRecord, nil
}

type DisputeWithUser struct {
	Dispute TransactionDisputeDao `gorm:"embedded"`
	User    MasterUserRecordDao   `gorm:"embedded"`
}

func (TransactionDisputeDao) SearchDisputes(searchTerm string, disputeStatus string, limit int, offset int) ([]DisputeWithUser, int64, error) {
	var disputeUsers []DisputeWithUser
	var totalCount int64

	query := db.DB.Table("transaction_disputes AS dispute_record").
		Select("dispute_record.*, user_record.*").
		Joins("INNER JOIN master_user_records AS user_record ON dispute_record.user_id = user_record.id")

	if disputeStatus != "" {
		query = query.Where("dispute_record.status = ?", disputeStatus)
	}

	if searchTerm != "" {
		searchTerms := parseSearchTerms(searchTerm)

		if len(searchTerms) > 0 {
			var conditions []string
			var args []any

			for _, term := range searchTerms {
				searchPattern := "%" + term.Value + "%"

				switch term.FieldHint {
				case FieldHintEmail:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"user_record.email",
					)
				case FieldHintPhone:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"user_record.mobile_no",
					)
				case FieldHintAlpha:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"user_record.email",
						"user_record.first_name",
						"user_record.last_name",
						"dispute_record.details",
						"dispute_record.extra_info",
						"dispute_record.reason",
					)
				case FieldHintNumber:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"user_record.ledger_customer_number",
						"user_record.id::text",
						"user_record.mobile_no",
						"dispute_record.id::text",
						"dispute_record.transaction_identifier",
					)
				case FieldHintGUID:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"user_record.id::text",
						"dispute_record.id::text",
						"dispute_record.transaction_identifier",
					)
				case FieldHintReferenceId:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"dispute_record.transaction_identifier",
					)
				case FieldHintUnknown:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"user_record.email",
						"user_record.first_name",
						"user_record.id::text",
						"user_record.last_name",
						"user_record.ledger_customer_number",
						"user_record.mobile_no",
						"dispute_record.details",
						"dispute_record.extra_info",
						"dispute_record.id::text",
						"dispute_record.reason",
						"dispute_record.transaction_identifier",
					)
				}
			}

			// Join all conditions with AND
			whereClause := strings.Join(conditions, " AND ")
			query = query.Where(whereClause, args...)
		}
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, errtrace.Wrap(err)
	}

	if err := query.Order("dispute_record.created_at DESC").Limit(limit).Offset(offset).Scan(&disputeUsers).Error; err != nil {
		return nil, 0, errtrace.Wrap(err)
	}

	return disputeUsers, totalCount, nil
}
