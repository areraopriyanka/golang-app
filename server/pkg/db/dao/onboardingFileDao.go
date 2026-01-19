package dao

import (
	"time"
)

type OnboardingFileDao struct {
	Uuid           string    `json:"uuid" gorm:"column:uuid"`
	TempCustomerNo string    `json:"tempCustomerNo" gorm:"column:temp_customer_no" mask:"true"`
	Name           string    `json:"name" gorm:"column:name"`
	Type           string    `json:"type" gorm:"column:type"`
	Url            string    `json:"url" gorm:"column:url"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at"`
	CreatedBy      string    `json:"createdBy" gorm:"column:created_by"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"column:updated_at"`
	UpdatedBy      string    `json:"updatedBy" gorm:"column:updated_by"`
}

// TableName overrides the table name used by OnboardingFileDao to `onboarding_files`
func (OnboardingFileDao) TableName() string {
	return "onboarding_files"
}
