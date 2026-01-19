package dao

import (
	"time"
)

// define model for user
type LedgerUserDao struct {
	UserEmail string    `json:"userEmail" gorm:"column:user_email" mask:"true"`
	JwtToken  string    `json:"jwtToken" gorm:"column:jwt_token" mask:"true"`
	CreatedAt time.Time `gorm:"column:created_at"`
	CreatedBy string    `gorm:"column:created_by"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	UpdatedBy string    `gorm:"column:updated_by"`
}

// TableName overrides the table name
func (LedgerUserDao) TableName() string {
	return "ledger_users"
}
