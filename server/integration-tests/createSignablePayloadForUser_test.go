package test

import (
	"process-api/pkg/db/dao"
)

func (suite *IntegrationTestSuite) TestCreateSignablePayloadForUser_Success() {
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	testPayload := map[string]interface{}{"key": "value"}

	result, err := dao.CreateSignablePayloadForUser(user.Id, testPayload)

	suite.Require().Nil(err)

	var payloadRecord dao.SignablePayloadDao
	dbErr := suite.TestDB.Where("id = ?", result.PayloadId).First(&payloadRecord).Error
	suite.NoError(dbErr)
	suite.Equal(user.Id, *payloadRecord.UserId)
	suite.Equal(result.Payload, payloadRecord.Payload)
	suite.Nil(payloadRecord.ConsumedAt)
}

func (suite *IntegrationTestSuite) TestCreateSignablePayloadForUser_UnmarshalablePayload() {
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	testPayload := make(chan int)

	result, err := dao.CreateSignablePayloadForUser(user.Id, testPayload)

	suite.Nil(result)
	suite.Require().NotNil(err)
	suite.Equal("PAYLOAD_MARSHAL_FAILED", err.ErrorCode)
	suite.Equal(500, err.StatusCode)
}
