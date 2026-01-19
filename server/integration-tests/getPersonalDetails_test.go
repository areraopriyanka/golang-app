package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/security"
	"time"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestGetPersonalDetails() {
	// Create a test user with all personal details
	user := suite.createUserWithPersonalDetails()
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)

	// Create a new Echo instance
	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodGet, "/account/personal-details", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/personal-details")

	// Create a custom context with the user ID
	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)

	// Call the handler
	err := handler.GetPersonalDetails(customContext)

	// Assert no error occurred
	suite.Require().NoError(err, "Handler should not return an error")

	// Assert the status code is 200 OK
	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	// Parse the response
	var responseBody handler.PersonalDetailsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	// Assert the response contains the expected personal details
	suite.Require().Equal(user.FirstName, responseBody.FirstName, "First name should match")
	suite.Require().Equal(user.LastName, responseBody.LastName, "Last name should match")
	suite.Require().Equal(user.Email, responseBody.Email, "Email should match")
	suite.Require().Equal(user.MobileNo, responseBody.MobileNumber, "Mobile number should match")

	// Assert address details
	suite.Require().Equal(user.StreetAddress, responseBody.Address.AddressLine1, "Address line 1 should match")
	suite.Require().Equal(user.ApartmentNo, responseBody.Address.AddressLine2, "Address line 2 should match")
	suite.Require().Equal(user.City, responseBody.Address.City, "City should match")
	suite.Require().Equal(user.State, responseBody.Address.State, "State should match")
	suite.Require().Equal(user.ZipCode, responseBody.Address.PostalCode, "Postal code should match")
}

// createUserWithPersonalDetails creates a test user with all personal details
func (suite *IntegrationTestSuite) createUserWithPersonalDetails() dao.MasterUserRecordDao {
	userId := uuid.New().String()
	dob, _ := time.Parse("2006-01-02", "1990-01-01")

	userRecord := dao.MasterUserRecordDao{
		Id:                   userId,
		FirstName:            "John",
		LastName:             "Doe",
		Email:                "john.doe@example.com",
		MobileNo:             "1234567890",
		DOB:                  dob,
		UserStatus:           constant.ACTIVE,
		StreetAddress:        "123 Main St",
		ApartmentNo:          "Apt 4B",
		City:                 "New York",
		State:                "NY",
		ZipCode:              "10001",
		LedgerCustomerNumber: "100000000006001",
	}

	err := suite.TestDB.Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user with personal details")
	return userRecord
}
