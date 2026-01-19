package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) RunUpdateAddress(apartmentNumber string) {
	userStatus := constant.PHONE_NUMBER_VERIFIED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	updateRequest := dao.UpdateCustomerAddressRequest{
		StreetAddress: "1600 Pennsylvania Ave NW",
		ApartmentNo:   apartmentNumber,
		ZipCode:       "20500",
		City:          "Washington",
		State:         "DC",
	}

	requestBody, _ := json.Marshal(updateRequest)

	e := handler.NewEcho()

	req := httptest.NewRequest(http.MethodPost, "/onboarding/customer/address", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err := handler.UpdateCustomerAddress(customContext)
	suite.Require().NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	// Verify if the database record was successfully updated
	var user dao.MasterUserRecordDao
	err = suite.TestDB.Model(&dao.MasterUserRecordDao{}).Where("id=?", testUser.Id).Find(&user).Error
	suite.Require().NoError(err, "Failed to fetch updated user data")

	suite.Equal("1600 Pennsylvania Ave NW", user.StreetAddress)
	suite.Equal(apartmentNumber, user.ApartmentNo)
	suite.Equal("20500", user.ZipCode)
	suite.Equal("Washington", user.City)
	suite.Equal("DC", user.State)

	suite.Equal(constant.ADDRESS_CONFIRMED, user.UserStatus, "status should be ADDRESS_CONFIRMED if successful")
}

func (suite *IntegrationTestSuite) TestUpdateAddressWithApartmentNumber() {
	suite.RunUpdateAddress("102")
}

func (suite *IntegrationTestSuite) TestUpdateAddressWithoutApartmentNumber() {
	suite.RunUpdateAddress("")
}

func (suite *IntegrationTestSuite) TestUpdateAddressForInvalidUserState() {
	userStatus := constant.KYC_PASS
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	updateRequest := dao.UpdateCustomerAddressRequest{
		StreetAddress: "1600 Pennsylvania Ave NW",
		ApartmentNo:   "Apt 1",
		ZipCode:       "20500",
		City:          "Washington",
		State:         "DC",
	}

	requestBody, _ := json.Marshal(updateRequest)

	e := handler.NewEcho()

	req := httptest.NewRequest(http.MethodPost, "/onboarding/customer/address", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err := handler.UpdateCustomerAddress(customContext)

	suite.Require().NotNil(err, "Handler should return an error for invalid state")
	errResp := err.(*response.ErrorResponse)
	suite.Equal(http.StatusPreconditionFailed, errResp.StatusCode, "Expected status code 412 StatusPreconditionFailed")
	suite.Equal(constant.INVALID_USER_STATE, errResp.ErrorCode, "Expected error code INVALID_USER_STATE")
}
