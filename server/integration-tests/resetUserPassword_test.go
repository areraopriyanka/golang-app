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

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestResetPassword() {
	resetToken := uuid.New().String()
	userRecord := suite.createTestUser(PartialMasterUserRecordDao{ResetToken: &resetToken})

	e := handler.NewEcho()

	request := request.ResetPasswordRequest{
		ResetToken: resetToken,
		Password:   "Test@2000",
	}

	requestBody, marshalErr := json.Marshal(request)
	suite.Require().NoError(marshalErr, "Error while marshaling")
	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler.ResetPassword(c)
	suite.NoError(err, "Handler should not return an error")
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var userData dao.MasterUserRecordDao
	err = suite.TestDB.Where("id=?", userRecord.Id).Find(&userData).Error
	suite.Require().NoError(err, "Failed to find test user")

	// check if user's password is set
	suite.Require().NotEmpty(userData.Password, "User password should not be nil or empty")
}

func (suite *IntegrationTestSuite) TestRestPasswordForInActiveUser() {
	resetToken := uuid.New().String()
	userStatus := constant.USER_CREATED
	_ = suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus, ResetToken: &resetToken})

	e := handler.NewEcho()

	request := request.ResetPasswordRequest{
		ResetToken: resetToken,
		Password:   "Test@2000",
	}

	requestBody, marshalErr := json.Marshal(request)
	suite.Require().NoError(marshalErr, "Error while marshaling")
	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler.ResetPassword(c)
	suite.Require().NotNil(err, "Handler should return an error for INACTIVE user")
	errResp := err.(response.ErrorResponse)
	suite.Require().Equal(http.StatusPreconditionFailed, errResp.StatusCode, "Expected status code 412 PreconditionFailed")
	suite.Equal(constant.USER_NOT_ACTIVE, errResp.ErrorCode, "Expected error code USER_NOT_ACTIVE")
}
