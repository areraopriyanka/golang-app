package dao

import (
	"time"
)

type MasterUserLoginDao struct {
	Id        string    `json:"id"`
	UserId    string    `gorm:"column:user_id"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (MasterUserLoginDao) TableName() string {
	return "master_user_logins"
}
