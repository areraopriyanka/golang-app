package dao

import (
	"encoding/json"
	"time"
)

// define model for kycErrorResponseDao
type KycErrorResponseDao struct {
	Id          int64           `gorm:"column:id"`
	CustomerNo  string          `json:"cno" gorm:"column:customer_no"`
	KycResponse json.RawMessage `json:"kycResponse" gorm:"column:kyc_response"  mask:"true"`
	CreatedAt   time.Time       `gorm:"column:created_at"`
}

// TableName overrides the table name used by KycErrorResponseDao to `kyc_error_response`
func (KycErrorResponseDao) TableName() string {
	return "kyc_error_response"
}
