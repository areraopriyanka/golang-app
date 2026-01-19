package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/request"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestUpdateCustomer() {
	suite.TestDB.AutoMigrate(&dao.MasterUserRecordDao{})

	sessionId := uuid.New().String()
	userRecord := dao.MasterUserRecordDao{
		Id:         sessionId,
		FirstName:  "Sample",
		LastName:   "User",
		Suffix:     "Jr",
		Email:      "example@email.com",
		UserStatus: constant.USER_CREATED,
	}

	err := suite.TestDB.Select("id", "first_name", "last_name", "email", "suffix", "user_status").Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	updateRequest := request.UpdateCustomerRequest{
		FirstName: "UpdatedName",
		LastName:  "UpdatedLastName",
		Email:     "updatedemail@example.com",
	}
	requestBody, _ := json.Marshal(updateRequest)

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/onboarding/customer/%s", sessionId), bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/onboarding/customer/:userId")
	c.SetParamNames("userId")
	c.SetParamValues(sessionId)

	err = handler.UpdateCustomer(c)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var user dao.MasterUserRecordDao
	err = suite.TestDB.Model(&dao.MasterUserRecordDao{}).Where("id=?", sessionId).Find(&user).Error
	suite.Require().NoError(err, "Failed to fetch updated user data")

	suite.Equal("UpdatedName", user.FirstName)
	suite.Equal("UpdatedLastName", user.LastName)
	suite.Equal("updatedemail@example.com", user.Email)

	// Suffix not provided in the update, so it is updated to empty string
	suite.Equal("", user.Suffix)
}
