package dao

import "time"

type ConfigsDao struct {
	Id         int       `gorm:"column:id"`
	ConfigName string    `gorm:"column:config_name"`
	Type       string    `gorm:"column:type"`
	Value      string    `gorm:"column:value"`
	UserType   string    `gorm:"column:user_type"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	CreatedBy  string    `gorm:"column:created_by"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
	UpdatedBy  string    `gorm:"column:updated_by"`
}

func (ConfigsDao) TableName() string {
	return "configs"
}
