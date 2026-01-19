package dao

import (
	"time"
)

// define model for user
type UserDao struct {
	CustomerNo     string `json:"cno" gorm:"column:customer_no" mask:"true"`
	Email          string `json:"email" mask:"true"`
	Mobile         string `json:"mobile" mask:"true"`
	Password       []byte `mask:"true"`
	Status         string
	IsLoggedIn     bool       `gorm:"column:is_logged_in"`
	CurrentLogInDT time.Time  `gorm:"column:current_log_in_dt;default:''"`
	LastLogInDT    time.Time  `gorm:"column:last_log_in_dt;default:''"`
	TokenExpiryDT  *time.Time `gorm:"column:token_expiry_dt"`
	LastLogOutDT   time.Time  `gorm:"column:last_log_out_dt;default:''"`
	UserType       string     `gorm:"column:user_type"`
	ApiName        string     `gorm:"column:api_name"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	CreatedBy      string     `gorm:"column:created_by"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
	UpdatedBy      string     `gorm:"column:updated_by"`
}

// TableName overrides the table name used by CustomerDao to `customers`
func (UserDao) TableName() string {
	return "users"
}
