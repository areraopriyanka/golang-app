package dao

import "time"

type MasterUserOtpDao struct {
	OtpId     string `gorm:"column:otp_id;primaryKey"`
	Otp       string
	OtpStatus string `gorm:"column:otp_status"`
	OtpType   string `gorm:"column:otp_type"`
	Count     int
	IP        string
	MobileNo  string `gorm:"column:mobile_no" mask:"true"`
	Email     string `gorm:"column:email" mask:"true"`
	UserId    string `gorm:"column:user_id;foreignKey:UserId;references:Id"`
	ApiPath   string `gorm:"column:api_path"`
	// Declaring it as a pointer allows us to assign a nil value. Otherwise,
	// GORM will attempt to insert the default value '0001-01-01 00:00:00 UTC',
	// which may lead to errors.
	UsedAt             *time.Time `gorm:"column:used_at"`
	CreatedAt          time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	ChallengeExpiredAt *time.Time `gorm:"column:challenge_expired_at"`
}

func (MasterUserOtpDao) TableName() string {
	return "master_user_otp"
}
