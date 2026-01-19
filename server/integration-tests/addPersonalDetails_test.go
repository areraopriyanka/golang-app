package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestAddPersonalDetailsForValidUserStatus() {
	e := handler.NewEcho()

	userStatus := constant.AGREEMENTS_REVIEWED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	requestBody := request.AddPersonalDetailsRequest{
		FirstName: "Test",
		LastName:  "User",
		Suffix:    "Jr.",
		DOB:       "02/02/2000",
	}

	body, err := json.Marshal(requestBody)
	suite.Require().NoError(err, "Failed to marshall request")

	req := httptest.NewRequest(http.MethodPost, "/customer/details", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err = handler.AddPersonalDetails(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var user dao.MasterUserRecordDao
	err = suite.TestDB.Model(&dao.MasterUserRecordDao{}).Where("id=?", testUser.Id).Find(&user).Error
	suite.Require().NoError(err, "Failed to fetch updated user data")
	suite.Equal(user.FirstName, "Test")
	suite.Equal(user.LastName, "User")
	suite.Equal(user.Suffix, "Jr.")
	suite.Equal(user.DOB.Format("01/02/2000"), "02/02/2000")
	suite.Equal(constant.AGE_VERIFICATION_PASSED, user.UserStatus)

	// Verify user_status from DB
	suite.ValidateUserStatusInDB(testUser.Id, constant.AGE_VERIFICATION_PASSED)
}

func (suite *IntegrationTestSuite) TestAddPersonalDetailsForInvalidUserStatus() {
	e := handler.NewEcho()

	userStatus := constant.PHONE_NUMBER_VERIFIED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	requestBody := request.AddPersonalDetailsRequest{
		FirstName: "Test",
		LastName:  "User",
		Suffix:    "Jr.",
		DOB:       "02/02/2000",
	}
	body, err := json.Marshal(requestBody)
	suite.Require().NoError(err, "Failed to marshall request")

	req := httptest.NewRequest(http.MethodPost, "/customer/details", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err = handler.AddPersonalDetails(customContext)

	suite.Require().NotNil(err, "Handler should return an error")
	errResp := err.(*response.ErrorResponse)
	suite.Equal(http.StatusPreconditionFailed, errResp.StatusCode, "Expected status code 412 StatusPreconditionFailed")
	suite.Equal(constant.INVALID_USER_STATE, errResp.ErrorCode, "Expected error code INVALID_USER_STATE")
}

func (suite *IntegrationTestSuite) TestAddPersonalDetailsForGreaterUserStatus() {
	e := handler.NewEcho()

	userStatus := constant.PHONE_VERIFICATION_OTP_SENT
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	requestBody := request.AddPersonalDetailsRequest{
		FirstName: "Test",
		LastName:  "User",
		Suffix:    "Jr.",
		DOB:       "02/02/2000",
	}

	body, err := json.Marshal(requestBody)
	suite.Require().NoError(err, "Failed to marshall request")

	req := httptest.NewRequest(http.MethodPost, "/customer/details", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err = handler.AddPersonalDetails(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var user dao.MasterUserRecordDao
	err = suite.TestDB.Model(&dao.MasterUserRecordDao{}).Where("id=?", testUser.Id).Find(&user).Error
	suite.Require().NoError(err, "Failed to fetch updated user data")
	suite.Equal(user.FirstName, "Test")
	suite.Equal(user.LastName, "User")
	suite.Equal(user.Suffix, "Jr.")
	suite.Equal(user.DOB.Format("01/02/2000"), "02/02/2000")
	suite.Equal(constant.PHONE_VERIFICATION_OTP_SENT, user.UserStatus)

	// Verify user_status from DB
	suite.ValidateUserStatusInDB(testUser.Id, constant.PHONE_VERIFICATION_OTP_SENT)
}
