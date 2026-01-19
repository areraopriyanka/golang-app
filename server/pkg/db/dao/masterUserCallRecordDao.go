package dao

import "time"

type MasterUserCallRecordDao struct {
	CallSid    string    `gorm:"column:call_sid;primaryKey"`
	CallStatus string    `gorm:"column:call_status"`
	To         string    `gorm:"column:to"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (MasterUserCallRecordDao) TableName() string {
	return "master_user_call_record"
}
