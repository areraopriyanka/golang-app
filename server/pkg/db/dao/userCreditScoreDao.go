package dao

import (
	"errors"
	"process-api/pkg/db"
	"time"

	"braces.dev/errtrace"
	"github.com/jinzhu/gorm"
)

type EncryptableCreditScoreData struct {
	Score                   int     `gorm:"column:score"`
	Increase                int     `gorm:"column:increase"`
	DebtwiseCustomerNumber  int     `gorm:"column:debtwise_customer_number"`
	PaymentHistoryAmount    float64 `gorm:"column:payment_history_amount"`
	PaymentHistoryFactor    string  `gorm:"column:payment_history_factor"`
	CreditUtilizationAmount float64 `gorm:"column:credit_utilization_amount"`
	CreditUtilizationFactor string  `gorm:"column:credit_utilization_factor"`
	DerogatoryMarksAmount   int     `gorm:"column:derogatory_marks_amount"`
	DerogatoryMarksFactor   string  `gorm:"column:derogatory_marks_factor"`
	CreditAgeAmount         float64 `gorm:"column:credit_age_amount"`
	CreditAgeFactor         string  `gorm:"column:credit_age_factor"`
	CreditMixAmount         int     `gorm:"column:credit_mix_amount"`
	CreditMixFactor         string  `gorm:"column:credit_mix_factor"`
	NewCreditAmount         int     `gorm:"column:new_credit_amount"`
	NewCreditFactor         string  `gorm:"column:new_credit_factor"`
	TotalAccountsAmount     int     `gorm:"column:total_accounts_amount"`
	TotalAccountsFactor     string  `gorm:"column:total_accounts_factor"`
}

type UserCreditScoreDao struct {
	Id                  string    `gorm:"column:id"`
	Date                time.Time `gorm:"column:date"`
	EncryptedCreditData []byte    `gorm:"column:encrypted_credit_data"`
	UserId              string    `gorm:"column:user_id;foreignKey:UserId;references:Id"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (UserCreditScoreDao) TableName() string {
	return "user_credit_score"
}

func (UserCreditScoreDao) FindLatestUserCreditScoreByUserId(userId string) (*UserCreditScoreDao, error) {
	var creditScore UserCreditScoreDao
	result := db.DB.Where("user_id = ?", userId).Order("created_at DESC").Take(&creditScore)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, errtrace.Wrap(result.Error)
	}
	return &creditScore, nil
}

func (UserCreditScoreDao) FindSufficientlyOldCreditScores(cutoffDate time.Time) ([]UserCreditScoreDao, error) {
	var oldRecords []UserCreditScoreDao
	err := db.DB.Raw(`
		SELECT DISTINCT ON (user_id) *
		FROM user_credit_score
		WHERE user_id IN (
			SELECT user_id
			FROM user_credit_score
			GROUP BY user_id
			HAVING MAX(date) < ?
		)
		ORDER BY user_id, date DESC
	`, cutoffDate).Scan(&oldRecords).Error
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	return oldRecords, nil
}

type EncryptedCreditScoreSummary struct {
	Date                string `json:"date" validate:"required" mask:"true"`
	EncryptedCreditData []byte `json:"encrypted_credit_data" validate:"required" mask:"true"`
}

type CreditScoreSummary struct {
	Date  string `json:"date" validate:"required" mask:"true"`
	Score int    `json:"score" validate:"required" mask:"true"`
}

func (UserCreditScoreDao) GetCreditScoreHistory(userId string) ([]EncryptedCreditScoreSummary, error) {
	var rawHistory []struct {
		Date                time.Time `gorm:"column:date"`
		EncryptedCreditData []byte    `gorm:"column:encrypted_credit_data"`
	}

	// NOTE: The credit score dashboard has a credit score history line graph that displays the trend in
	// credit score over the last 6 months so we are collecting that here to be returned with the score response
	result := db.DB.Model(&UserCreditScoreDao{}).Select("date, encrypted_credit_data").Where("user_id = ?", userId).Order("created_at DESC").Limit(6).Scan(&rawHistory)

	if result.Error != nil {
		return nil, errtrace.Wrap(result.Error)
	}

	var creditScoreHistory []EncryptedCreditScoreSummary
	for _, record := range rawHistory {
		creditScoreHistory = append(creditScoreHistory, EncryptedCreditScoreSummary{
			Date:                record.Date.Format("2006-01-02"),
			EncryptedCreditData: record.EncryptedCreditData,
		})
	}

	return creditScoreHistory, nil
}
