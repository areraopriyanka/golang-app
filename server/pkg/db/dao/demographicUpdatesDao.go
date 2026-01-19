package dao

import (
	"encoding/json"
	"errors"
	"process-api/pkg/db"
	"strings"
	"time"

	"braces.dev/errtrace"
	"github.com/jinzhu/gorm"
)

type DemographicUpdatesDao struct {
	Id           string          `json:"id" gorm:"column:id;primaryKey"`
	Type         string          `json:"type"`
	Status       string          `json:"status"`
	UpdatedValue json.RawMessage `json:"value" gorm:"column:updated_value;type:jsonb"`
	ExtraInfo    string          `json:"extraInfo" gorm:"column:extra_info"`
	UserId       string          `gorm:"column:user_id;foreignKey:UserId;references:Id"`
	CreatedAt    time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time       `gorm:"column:updated_at;autoUpdateTime"`
}

func (DemographicUpdatesDao) TableName() string {
	return "demographic_updates"
}

type UpdateFullNameRequest struct {
	FirstName string `json:"firstName" validate:"required,validateName"`
	LastName  string `json:"lastName" validate:"required,validateName"`
	Suffix    string `json:"suffix"`
}
type UpdateCustomerAddressRequest struct {
	StreetAddress string `json:"streetAddress" validate:"required,validateAddress"`
	ApartmentNo   string `json:"apartmentNo" validate:"validateAddress"`
	ZipCode       string `json:"zipCode" validate:"required,validateZipCode"`
	City          string `json:"city" validate:"required,validateCity"`
	State         string `json:"state" validate:"required,validateState"`
}

func (DemographicUpdatesDao) FindByTypeAndUserId(userId, updateType string) (*DemographicUpdatesDao, error) {
	var demographicUpdateRecord DemographicUpdatesDao
	result := db.DB.Where("user_id = ? AND type = ? AND status = ?", userId, updateType, "pending").First(&demographicUpdateRecord)
	// record not found
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, errtrace.Wrap(result.Error)
	}
	return &demographicUpdateRecord, nil
}

func (DemographicUpdatesDao) FindDemographicUpdatesByUserId(userId string) ([]DemographicUpdatesDao, error) {
	var demographicUpdates []DemographicUpdatesDao

	result := db.DB.
		Select("DISTINCT ON (type) status, type, updated_value").
		Where("user_id = ?", userId).
		Order("type, created_at DESC").
		Find(&demographicUpdates)

	if result.Error != nil {
		return nil, errtrace.Wrap(result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return demographicUpdates, nil
}

func (DemographicUpdatesDao) FindById(demographicId string) (*DemographicUpdatesDao, error) {
	var demographicUpdateRecord DemographicUpdatesDao
	result := db.DB.Where("id = ?", demographicId).Take(&demographicUpdateRecord)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, errtrace.Wrap(result.Error)
	}
	return &demographicUpdateRecord, nil
}

type DemographicUpdateWithUser struct {
	DemographicUpdate DemographicUpdatesDao `gorm:"embedded"`
	User              MasterUserRecordDao   `gorm:"embedded"`
}

func (DemographicUpdatesDao) SearchDemographicUpdates(searchTerm string, demographicStatus string, demographicType string, limit int, offset int) ([]DemographicUpdateWithUser, int64, error) {
	var demographicUsers []DemographicUpdateWithUser
	var totalCount int64

	query := db.DB.Table("demographic_updates AS update_record").
		Select("update_record.*, user_record.*").
		Joins("INNER JOIN master_user_records AS user_record ON update_record.user_id = user_record.id")

	if demographicStatus != "" {
		query = query.Where("update_record.status = ?", demographicStatus)
	}

	if demographicType != "" {
		query = query.Where("update_record.type = ?", demographicType)
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
						"update_record.extra_info",
						"update_record.updated_value::text",
					)
				case FieldHintNumber:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"user_record.ledger_customer_number",
						"user_record.id::text",
						"user_record.mobile_no",
						"update_record.id::text",
						"update_record.updated_value::text",
					)
				case FieldHintGUID:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"user_record.id::text",
						"update_record.id::text",
					)
				case FieldHintUnknown:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"user_record.email",
						"user_record.first_name",
						"user_record.id::text",
						"user_record.last_name",
						"user_record.ledger_customer_number",
						"user_record.mobile_no",
						"update_record.extra_info",
						"update_record.id::text",
						"update_record.updated_value::text",
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

	if err := query.Order("update_record.created_at DESC").Limit(limit).Offset(offset).Scan(&demographicUsers).Error; err != nil {
		return nil, 0, errtrace.Wrap(err)
	}

	return demographicUsers, totalCount, nil
}
