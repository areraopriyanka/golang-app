package dao

import (
	"errors"
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/model/response"
	"time"

	"braces.dev/errtrace"
	"github.com/jinzhu/gorm"
)

type UserPublicKey struct {
	ID                 uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId             string    `json:"userId" gorm:"column:user_id;size:36;not null;index"`
	KeyId              string    `json:"keyId" gorm:"size:50;not null;column:key_id" mask:"true"`
	ApiKey             []byte    `json:"apiKey" gorm:"column:encrypted_api_key" mask:"true"`
	KmsEncryptedApiKey []byte    `json:"kmsEncryptedApiKey" gorm:"column:kms_encrypted_api_key" mask:"true"`
	PublicKey          string    `json:"publicKey" gorm:"column:public_key;size:255;not null;uniqueIndex" mask:"true"`
	CreatedAt          time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (UserPublicKey) TableName() string {
	return "user_public_keys"
}

func (UserPublicKey) FindUserPublicRecord(userId, publicKey string) (*UserPublicKey, error) {
	var userPublicKey UserPublicKey
	result := db.DB.Model(UserPublicKey{}).Where("user_id = ? AND public_key = ?", userId, publicKey).Take(&userPublicKey)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, errtrace.Wrap(result.Error)
	}
	return &userPublicKey, nil
}

func RequireUserPublicKey(userId, publicKey string) (*UserPublicKey, *response.ErrorResponse) {
	userPublicKey, err := UserPublicKey{}.FindUserPublicRecord(userId, publicKey)
	if err != nil {
		return nil, &response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("DB Error: %s", err), MaybeInnerError: errtrace.Wrap(err)}
	}
	if userPublicKey == nil {
		return nil, &response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, LogMessage: "userPublicKey record not found in DB", MaybeInnerError: errtrace.New("")}
	}
	return userPublicKey, nil
}

func (UserPublicKey) FindAll() ([]UserPublicKey, error) {
	var records []UserPublicKey
	err := db.DB.Find(&records).Error
	return records, errtrace.Wrap(err)
}
