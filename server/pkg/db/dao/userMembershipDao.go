package dao

import (
	"errors"
	"process-api/pkg/db"
	"time"

	"braces.dev/errtrace"
	"github.com/jinzhu/gorm"
)

type UserMembershipDao struct {
	Id               string    `gorm:"column:id;primaryKey"`
	UserID           string    `gorm:"column:user_id;foreignKey:UserId;references:Id"`
	MembershipStatus string    `gorm:"column:membership_status"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (UserMembershipDao) FindOneByUserId(userId string) (*UserMembershipDao, error) {
	var userMembership UserMembershipDao
	result := db.DB.First(&userMembership, "user_id= ?", userId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, errtrace.Wrap(result.Error)
		}
	}
	return &userMembership, nil
}

func (UserMembershipDao) TableName() string {
	return "user_membership"
}
