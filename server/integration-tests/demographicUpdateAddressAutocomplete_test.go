package test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"process-api/pkg/config"
	"process-api/pkg/handler"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestDemographicUpdateAddressAutoCompleteSuccess() {
	mockSmartyServer := SetupMockForSmarty(suite)
	defer mockSmartyServer.Close()

	config.Config.SmartyStreets.LookupUrl = mockSmartyServer.URL
	config.Config.SmartyStreets.AuthId = "test-authId"
	config.Config.SmartyStreets.AuthToken = "test-auth-token"

	testUser := suite.createTestUser(PartialMasterUserRecordDao{})

	e := handler.NewEcho()

	params := url.Values{}

	params.Add("search", "1600 Pennsylvania")

	req := httptest.NewRequest(http.MethodGet, "/account/customer/demographic-update/address-autocomplete?"+params.Encode(), nil)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateLoggedInRegisteredUserContext(testUser.Id, "examplePublicKey", c)

	err := handler.AddressAutoComplete(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)
}
