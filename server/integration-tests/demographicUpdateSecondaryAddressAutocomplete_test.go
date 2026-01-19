package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestDemographicUpdateSecondaryAddressAutoCompleteSuccess() {
	mockSmartyServer := SetupMockForSmarty(suite)
	defer mockSmartyServer.Close()

	config.Config.SmartyStreets.LookupUrl = mockSmartyServer.URL
	config.Config.SmartyStreets.AuthId = "test-authId"
	config.Config.SmartyStreets.AuthToken = "test-auth-token"

	testUser := suite.createTestUser(PartialMasterUserRecordDao{})

	e := handler.NewEcho()

	request := response.Suggestion{
		Street:    "1042 E Center St",
		Secondary: "#",
		City:      "Wallingford",
		State:     "CT",
		ZipCode:   "06492",
		Entries:   2,
	}

	requestBody, err := json.Marshal(request)
	suite.Require().NoError(err, "Failed to marshall request")

	req := httptest.NewRequest(http.MethodPost, "/account/customer/secondary-address-autocomplete", bytes.NewReader(requestBody))

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateLoggedInRegisteredUserContext(testUser.Id, "examplePublicKey", c)

	err = handler.DemographicUpdateSecondaryAddressAutoComplete(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)
}
