package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/request"
	"process-api/pkg/security"

	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

func (suite *IntegrationTestSuite) TestChangePassword() {
	hashedOldPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	suite.Require().NoError(err, "Failed to encrypt password")

	sessionId := uuid.New().String()
	resetToken := uuid.New().String()

	userRecord := dao.MasterUserRecordDao{
		Id:         sessionId,
		FirstName:  "Foo",
		LastName:   "Bar",
		Email:      "foo@bar.com",
		Password:   hashedOldPassword,
		ResetToken: resetToken,
		UserStatus: "ACTIVE",
	}

	err = suite.TestDB.Select("id", "first_name", "last_name", "email", "password", "reset_token", "user_status").Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	changeRequest := request.ChangePasswordRequest{
		OldPassword: "Password123!",
		NewPassword: "Password456!",
		ResetToken:  resetToken,
	}
	requestBody, err := json.Marshal(changeRequest)
	suite.Require().NoError(err, "Failed to marshall request")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/change-password")

	customContext := security.GenerateLoggedInRegisteredUserContext(sessionId, "examplePublicKey", c)

	err = handler.ChangePassword(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var user dao.MasterUserRecordDao

	err = suite.TestDB.Model(&dao.MasterUserRecordDao{}).Where("id=?", sessionId).Find(&user).Error
	suite.Require().NoError(err, "Failed to fetch user data")

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("Password456!"))
	suite.Require().NoError(err, "User record password does not match new password from request")
	suite.Require().Empty(user.ResetToken, "User record's reset token is cleared after use")
}
