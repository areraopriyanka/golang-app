package dao

import (
	"time"
)

type SardineDataDao struct {
	Id                 int       `gorm:"column:id;primaryKey;autoIncrement"`
	SessionKey         string    `gorm:"column:session_key" mask:"true"`
	SessionKeyExpiryDT time.Time `gorm:"column:session_key_expiry_dt"`
	// Explicitly specifying the type as TEXT. Otherwise, db.AutoMigrate(&dao.SardineDataDao{}) will default to using VARCHAR(255).
	SardineResponse string    `gorm:"column:sardine_response;type:text" mask:"true"`
	UserId          string    `gorm:"column:user_id;foreignKey:UserId;references:Id"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (SardineDataDao) TableName() string {
	return "sardine_kyc_data"
}
