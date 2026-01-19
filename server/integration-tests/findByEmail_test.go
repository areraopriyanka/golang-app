package test

import (
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestFindUserByEmail() {
	sessionId := uuid.New().String()

	userRecord := dao.MasterUserRecordDao{
		Id:         sessionId,
		Email:      "testuser@gmail.com",
		FirstName:  "Test",
		LastName:   "User",
		UserStatus: constant.USER_CREATED,
	}

	err := suite.TestDB.Select("id", "email", "first_name", "last_name", "user_status").Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	// exact string match
	user, err := dao.MasterUserRecordDao{}.FindUserByEmail("testuser@gmail.com")
	suite.Require().NoError(err, "DB error occurred")
	suite.Require().NotNil(user, "Expected user to be found")

	// case insensitive string match
	user, err = dao.MasterUserRecordDao{}.FindUserByEmail("TesTUsEr@gmail.com")
	suite.Require().NoError(err, "DB error occurred")
	suite.Require().NotNil(user, "Expected user to be found")

	// no record found
	user, err = dao.MasterUserRecordDao{}.FindUserByEmail("test@gmail.com")
	suite.Require().NoError(err, "DB error occurred")
	suite.Require().Nil(user, "Expected no user to be found")
}
