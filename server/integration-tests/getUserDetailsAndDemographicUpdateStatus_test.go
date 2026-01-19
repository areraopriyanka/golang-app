package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestGetUserDetailsAndDemographicUpdateStatus() {
	user := suite.createTestUser(PartialMasterUserRecordDao{})

	e := handler.NewEcho()

	req := httptest.NewRequest(http.MethodGet, "/account/customer/demographic-update", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/customer/demographic-update")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, "publicKey", c)

	// Create full_name demographic update record
	fullNameRequest := dao.UpdateFullNameRequest{
		FirstName: "Test",
		LastName:  "User",
		Suffix:    "Jr",
	}

	full_name, err := json.Marshal(fullNameRequest)
	suite.Require().NoError(err, "Failed to marshall full_name")
	suite.createDemographicUpdateRecord(json.RawMessage(full_name), "full_name", user.Id)

	// Create address demographic update record
	addressRequest := dao.UpdateCustomerAddressRequest{
		StreetAddress: "1600 Pennsylvania Ave NW",
		ApartmentNo:   "#1",
		ZipCode:       "20500",
		City:          "Washington",
		State:         "DC",
	}
	address, err := json.Marshal(addressRequest)
	suite.Require().NoError(err, "Failed to marshall address")
	suite.createDemographicUpdateRecord(json.RawMessage(address), "address", user.Id)

	err = handler.GetUserDetailsAndDemographicUpdateStatus(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.GetUserDetailsAndUpdateStatus
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to parse response body")

	suite.Require().Equal("Test", responseBody.FirstName, "Invalid firstName received in response")
	suite.Require().Equal("Bar", responseBody.LastName, "Invalid lastName received in response")
	suite.Require().Equal("testuser@gmail.com", responseBody.Email, "Invalid email received in response")
	suite.Require().Equal("+14159871234", responseBody.MobileNumber, "Invalid mobileNumber received in response")
	suite.Require().Equal("pending", responseBody.FullNameStatus)
	suite.Equal("pending", responseBody.AddressStatus)
}

func (suite *IntegrationTestSuite) createDemographicUpdateRecord(value json.RawMessage, demographicUpdateType, userId string) {
	id := uuid.New().String()
	demographicUpdate := dao.DemographicUpdatesDao{
		Id:           id,
		Type:         demographicUpdateType,
		Status:       "pending",
		UpdatedValue: value,
		UserId:       userId,
	}

	err := db.DB.Select("id", "type", "status", "updated_value", "user_id").Create(&demographicUpdate).Error
	suite.Require().NoError(err, "Failed to create test demographic update record")
}
