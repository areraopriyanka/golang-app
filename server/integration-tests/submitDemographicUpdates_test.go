package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestSubmitFullNameDemographicUpdate() {
	user := suite.createTestUser(PartialMasterUserRecordDao{})

	fullNameRequest := dao.UpdateFullNameRequest{
		FirstName: "Test",
		LastName:  "User",
	}
	requestBody, err := json.Marshal(fullNameRequest)
	suite.Require().NoError(err, "Failed to marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/account/customer/demographic-update/full-name", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/customer/demographic-update/full-name")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, "publicKey", c)

	err = handler.SubmitFullNameDemographicUpdates(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusCreated, rec.Code, "Expected status code 201 Created")

	var demographicUpdateRecord dao.DemographicUpdatesDao
	err = suite.TestDB.Model(&dao.DemographicUpdatesDao{}).Where("user_id = ?", user.Id).First(&demographicUpdateRecord).Error
	suite.Require().NoError(err, "Failed to fetch demographic update record from database")
	suite.Require().Equal("pending", demographicUpdateRecord.Status, "Invalid demographic update status")
	suite.Require().NotEmpty(demographicUpdateRecord.UpdatedValue, "Updated value should not be empty")

	var expectedRequestBody dao.UpdateFullNameRequest
	err = json.Unmarshal(requestBody, &expectedRequestBody)
	suite.Require().NoError(err)

	var dbValue dao.UpdateFullNameRequest
	err = json.Unmarshal(demographicUpdateRecord.UpdatedValue, &dbValue)
	suite.Require().NoError(err)

	suite.Equal(expectedRequestBody, dbValue)
}

func (suite *IntegrationTestSuite) TestSubmitAddressDemographicUpdate() {
	user := suite.createTestUser(PartialMasterUserRecordDao{})

	addressRequest := dao.UpdateCustomerAddressRequest{
		StreetAddress: "1600 Pennsylvania Ave NW",
		ApartmentNo:   "102",
		ZipCode:       "20500",
		City:          "Washington",
		State:         "DC",
	}
	requestBody, err := json.Marshal(addressRequest)
	suite.Require().NoError(err, "Failed to marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/account/customer/demographic-update/address", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/account/customer/demographic-update/address")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, "publicKey", c)

	err = handler.SubmitAddressDemographicUpdates(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusCreated, rec.Code, "Expected status code 201 Created")

	var demographicUpdateRecord dao.DemographicUpdatesDao
	err = suite.TestDB.Model(&dao.DemographicUpdatesDao{}).Where("user_id = ?", user.Id).First(&demographicUpdateRecord).Error
	suite.Require().NoError(err, "Failed to fetch demographic update record from database")
	suite.Require().Equal("pending", demographicUpdateRecord.Status, "Invalid demographic update status")

	suite.Require().NotEmpty(demographicUpdateRecord.UpdatedValue, "Updated value should not be empty")

	var expectedRequestBody dao.UpdateCustomerAddressRequest
	err = json.Unmarshal(requestBody, &expectedRequestBody)
	suite.Require().NoError(err)

	var dbValue dao.UpdateCustomerAddressRequest
	err = json.Unmarshal(demographicUpdateRecord.UpdatedValue, &dbValue)
	suite.Require().NoError(err)

	suite.Equal(expectedRequestBody, dbValue)
}

func (suite *IntegrationTestSuite) TestSubmitDemographicUpdateForValidationError() {
	user := suite.createTestUser(PartialMasterUserRecordDao{})

	updateRequest := dao.UpdateCustomerAddressRequest{
		StreetAddress: "",
		ApartmentNo:   "102",
		ZipCode:       "20500",
		City:          "Washington",
		State:         "DC",
	}

	requestBody, err := json.Marshal(updateRequest)
	suite.Require().NoError(err, "Failed to marshall request body")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/account/customer/demographic-update/address", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, "publicKey", c)

	err = handler.SubmitAddressDemographicUpdates(customContext)
	suite.Require().NotNil(err, "Handler should return an error for empty street address")

	e.HTTPErrorHandler(err, c)

	var responseBody response.BadRequestErrors
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(http.StatusBadRequest, rec.Code, "Expected status code 400 badRequest")
	suite.Equal("streetAddress", responseBody.Errors[0].FieldName, "Invalid field name")
	suite.Equal("required", responseBody.Errors[0].Error, "Invalid error message")
}
