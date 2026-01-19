package test

import (
	"process-api/pkg/clock"
	"process-api/pkg/db/dao"
	"time"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestConsumePayload_Success() {
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	payloadId := uuid.New().String()
	payload := dao.SignablePayloadDao{
		Id:      payloadId,
		Payload: `{"test": "data"}`,
		UserId:  &user.Id,
	}
	err := suite.TestDB.Create(&payload).Error
	suite.Require().NoError(err)

	result, err := dao.ConsumePayload(user.Id, payloadId)

	suite.Require().Nil(err)
	suite.Equal(payloadId, result.Id)
	suite.Equal(`{"test": "data"}`, result.Payload)
	suite.NotNil(result.ConsumedAt)
}

func (suite *IntegrationTestSuite) TestConsumePayload_NotFound() {
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	nonExistentPayloadId := uuid.New().String()

	result, err := dao.ConsumePayload(user.Id, nonExistentPayloadId)

	suite.Nil(result)
	suite.Equal("PAYLOAD_NOT_FOUND", err.ErrorCode)
}

func (suite *IntegrationTestSuite) TestConsumePayload_WrongUser() {
	user1 := suite.createTestUser(PartialMasterUserRecordDao{})
	user2Email := "anotheruser+123@example.com"
	user2MobileNo := "+12125551234"
	user2 := suite.createTestUser(PartialMasterUserRecordDao{Email: &user2Email, MobileNo: &user2MobileNo})
	payloadId := uuid.New().String()
	payload := dao.SignablePayloadDao{
		Id:      payloadId,
		Payload: `{"test": "data"}`,
		UserId:  &user1.Id,
	}
	err := suite.TestDB.Create(&payload).Error
	suite.Require().NoError(err)

	result, errResponse := dao.ConsumePayload(user2.Id, payloadId)

	suite.Nil(result)
	suite.Equal("PAYLOAD_NOT_FOUND", errResponse.ErrorCode)
}

func (suite *IntegrationTestSuite) TestConsumePayload_AlreadyConsumed() {
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	payloadId := uuid.New().String()
	consumedTime := clock.Now().Add(-1 * time.Minute)
	payload := dao.SignablePayloadDao{
		Id:         payloadId,
		Payload:    `{"test": "data"}`,
		UserId:     &user.Id,
		ConsumedAt: &consumedTime,
	}
	err := suite.TestDB.Create(&payload).Error
	suite.Require().NoError(err)

	result, errResponse := dao.ConsumePayload(user.Id, payloadId)

	suite.Nil(result)
	suite.Equal("PAYLOAD_ALREADY_CONSUMED", errResponse.ErrorCode)
}

func (suite *IntegrationTestSuite) TestConsumePayload_Expired() {
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	payloadId := uuid.New().String()
	payload := dao.SignablePayloadDao{
		Id:        payloadId,
		Payload:   `{"test": "data"}`,
		UserId:    &user.Id,
		CreatedAt: clock.Now().Add(-5 * time.Minute),
	}
	err := suite.TestDB.Create(&payload).Error
	suite.Require().NoError(err)

	result, errResponse := dao.ConsumePayload(user.Id, payloadId)

	suite.Nil(result)
	suite.Equal("PAYLOAD_EXPIRED", errResponse.ErrorCode)
}

func (suite *IntegrationTestSuite) TestConsumePayload_ReplayAttack() {
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	payloadId := uuid.New().String()
	payload := dao.SignablePayloadDao{
		Id:      payloadId,
		Payload: `{"test": "data"}`,
		UserId:  &user.Id,
	}
	err := suite.TestDB.Create(&payload).Error
	suite.Require().NoError(err)

	result1, err1 := dao.ConsumePayload(user.Id, payloadId)
	suite.Require().Nil(err1)
	suite.NotNil(result1)
	suite.NotNil(result1.ConsumedAt)

	result2, errRepsonse2 := dao.ConsumePayload(user.Id, payloadId)

	suite.Nil(result2)
	suite.Equal("PAYLOAD_ALREADY_CONSUMED", errRepsonse2.ErrorCode)
}

func (suite *IntegrationTestSuite) TestConsumePayload_BoundaryExpiration() {
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	payloadId := uuid.New().String()
	payload := dao.SignablePayloadDao{
		Id:        payloadId,
		Payload:   `{"test": "data"}`,
		UserId:    &user.Id,
		CreatedAt: clock.Now().Add(-3*time.Minute - time.Second),
	}
	err := suite.TestDB.Create(&payload).Error
	suite.Require().NoError(err)

	result, errResponse := dao.ConsumePayload(user.Id, payloadId)

	suite.Nil(result)
	suite.Equal("PAYLOAD_EXPIRED", errResponse.ErrorCode)
}

func (suite *IntegrationTestSuite) TestConsumePayload_NullUserId() {
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	payloadId := uuid.New().String()
	payload := dao.SignablePayloadDao{
		Id:      payloadId,
		Payload: `{"test": "data"}`,
		UserId:  nil,
	}
	err := suite.TestDB.Create(&payload).Error
	suite.Require().NoError(err)

	result, errResponse := dao.ConsumePayload(user.Id, payloadId)

	suite.Nil(result)
	suite.Equal("PAYLOAD_NOT_FOUND", errResponse.ErrorCode)
}
