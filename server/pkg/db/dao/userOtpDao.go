package dao

import (
	"time"
)

// define model for user
type UserOtpDao struct {
	Id         int64     `gorm:"column:id"`
	CustomerNo string    `json:"cno" gorm:"column:customer_no" mask:"true"`
	Receiver   string    `json:"receiver" gorm:"column:receiver"`
	Otp        string    `json:"otp" gorm:"column:otp"`
	Type       string    `gorm:"column:type"`
	Status     string    `gorm:"column:status"`
	Count      int       `gorm:"column:count"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	CreatedBy  string    `gorm:"column:created_by"`
}

// TableName overrides the table name used by CustomerDao to `customers`
func (UserOtpDao) TableName() string {
	return "user_otp"
}
