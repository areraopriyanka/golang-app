package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"process-api/pkg/clock"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/model/response"
	"time"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type SignablePayloadDao struct {
	Id         string     `json:"id" gorm:"column:id;primaryKey"`
	Payload    string     `json:"payload" gorm:"column:payload;type:text" mask:"true"`
	UserId     *string    `json:"userId" gorm:"column:user_id"`
	ConsumedAt *time.Time `json:"consumedAt" gorm:"column:consumed_at"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (SignablePayloadDao) TableName() string {
	return "signable_payloads"
}

func (SignablePayloadDao) FindById(payloadId string) (*SignablePayloadDao, error) {
	var payloadRecord SignablePayloadDao
	result := db.DB.Where("id = ?", payloadId).Take(&payloadRecord)
	// payloadRecord not found
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, errtrace.Wrap(result.Error)
	}
	return &payloadRecord, nil
}

func RequireSignablePayload(payloadId string) (*SignablePayloadDao, *response.ErrorResponse) {
	payloadRecord, err := SignablePayloadDao{}.FindById(payloadId)
	if err != nil {
		return nil, &response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("DB Error: %s", err), MaybeInnerError: errtrace.Wrap(err)}
	}
	if payloadRecord == nil {
		return nil, &response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, LogMessage: "payloadRecord not found in DB", MaybeInnerError: errtrace.New("")}
	}
	return payloadRecord, nil
}

func ConsumePayload(userId string, payloadId string) (*SignablePayloadDao, *response.ErrorResponse) {
	var payloadRecord SignablePayloadDao
	// Ensure the payload's associated with this user. If we want to later, we can return
	// a specific error if the payload record exists but the payload isn't associated with the user.
	err := db.DB.Where("user_id = ? AND id = ?", userId, payloadId).Find(&payloadRecord).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &response.ErrorResponse{ErrorCode: "PAYLOAD_NOT_FOUND", StatusCode: http.StatusNotFound, LogMessage: "could not find payload", MaybeInnerError: errtrace.Wrap(err)}
	}
	if err != nil {
		return nil, &response.ErrorResponse{ErrorCode: "PAYLOAD_INTERNAL_ERROR", StatusCode: http.StatusInternalServerError, LogMessage: "error fetching payload from the db", MaybeInnerError: errtrace.Wrap(err)}
	}
	if payloadRecord.ConsumedAt != nil {
		return nil, &response.ErrorResponse{ErrorCode: "PAYLOAD_ALREADY_CONSUMED", StatusCode: http.StatusConflict, LogMessage: "payload already consumed", MaybeInnerError: errtrace.New("")}
	}

	// TODO: for now, hardcoding this to 3 minutes:
	expirationTime := payloadRecord.CreatedAt.Add(3 * time.Minute)
	if clock.Now().After(expirationTime) {
		return nil, &response.ErrorResponse{ErrorCode: "PAYLOAD_EXPIRED", StatusCode: http.StatusGone, LogMessage: "payload expired", MaybeInnerError: errtrace.New("")}
	}

	now := clock.Now()
	payloadRecord.ConsumedAt = &now
	err = db.DB.Save(&payloadRecord).Error
	if err != nil {
		return nil, &response.ErrorResponse{ErrorCode: "PAYLOAD_INTERNAL_ERROR", StatusCode: http.StatusInternalServerError, LogMessage: "error saving payload to db", MaybeInnerError: errtrace.Wrap(err)}
	}

	return &payloadRecord, nil
}

func CreateSignablePayloadForUser(userId string, payload any) (*response.BuildPayloadResponse, *response.ErrorResponse) {
	jsonPayloadBytes, marshallErr := json.Marshal(payload)
	if marshallErr != nil {
		return nil, &response.ErrorResponse{ErrorCode: "PAYLOAD_MARSHAL_FAILED", StatusCode: http.StatusInternalServerError, LogMessage: "error marshaling payload", MaybeInnerError: errtrace.Wrap(marshallErr)}
	}

	id := uuid.New().String()
	payloadRecord := SignablePayloadDao{
		Id:      id,
		UserId:  &userId,
		Payload: string(jsonPayloadBytes),
	}

	result := db.DB.Create(&payloadRecord)
	if result.Error != nil {
		return nil, &response.ErrorResponse{ErrorCode: "PAYLOAD_CREATE_FAILED", StatusCode: http.StatusInternalServerError, LogMessage: "error creating payload", MaybeInnerError: errtrace.Wrap(result.Error)}
	}

	response := &response.BuildPayloadResponse{
		PayloadId: id,
		Payload:   string(jsonPayloadBytes),
	}
	return response, nil
}
