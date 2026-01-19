package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"strings"
)

func SetupMockForSmarty(suite *IntegrationTestSuite) *httptest.Server {
	// Sets up mock smarty to specify response to sardine API
	mockLedgerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")

		var response map[string]interface{}

		switch {
		case strings.HasPrefix(search, "1600 Pennsylv"):
			response = map[string]interface{}{
				"suggestions": []map[string]interface{}{
					{"street_line": "1600 Pennsylvania Ave", "secondary": "# 1", "city": "Tyrone", "state": "PA", "zipcode": "16686", "entries": 1},
					{"street_line": "1600 Pennsylvania Ave", "secondary": "# 292", "city": "Hundred", "state": "WV", "zipcode": "26575", "entries": 1},
					{"street_line": "1600 Pennsylvania Ave", "secondary": "Apt", "city": "Miami Beach", "state": "FL", "zipcode": "33139", "entries": 20},
					{"street_line": "1600 Pennsylvania Ave", "secondary": "Apt", "city": "Stoughton", "state": "MA", "zipcode": "02072", "entries": 10},
					{"street_line": "1600 Pennsylvania Ave", "secondary": "", "city": "Austin", "state": "TX", "zipcode": "78702", "entries": 0},
					{"street_line": "1600 Pennsylvania Ave", "secondary": "", "city": "Charleston", "state": "WV", "zipcode": "25302", "entries": 0},
					{"street_line": "1600 Pennsylvania Ave", "secondary": "", "city": "Clearmont", "state": "WY", "zipcode": "82835", "entries": 0},
					{"street_line": "1600 N Pennsylvania Ave", "secondary": "", "city": "Mangum", "state": "OK", "zipcode": "73554", "entries": 0},
					{"street_line": "1600 N Pennsylvania Ave", "secondary": "", "city": "Oklahoma City", "state": "OK", "zipcode": "73107", "entries": 0},
					{"street_line": "1600 N Pennsylvania St", "secondary": "", "city": "Denver", "state": "CO", "zipcode": "80203", "entries": 0},
				},
			}

		case strings.HasPrefix(search, "PO Box"):
			response = map[string]interface{}{
				"suggestions": []map[string]interface{}{
					{"street_line": "PO Box 0", "secondary": "", "city": "Burnt Ranch", "state": "CA", "zipcode": "95527", "entries": 0},
					{"street_line": "PO Box 0", "secondary": "", "city": "Champlain", "state": "NY", "zipcode": "12919", "entries": 0},
					{"street_line": "PO Box 0", "secondary": "", "city": "Dannemora", "state": "NY", "zipcode": "12929", "entries": 0},
					{"street_line": "PO Box 0", "secondary": "", "city": "Friendswood", "state": "TX", "zipcode": "77549", "entries": 0},
					{"street_line": "PO Box 0", "secondary": "", "city": "Groton", "state": "CT", "zipcode": "06349", "entries": 0},
					{"street_line": "PO Box 0", "secondary": "", "city": "La Canada Flintridge", "state": "CA", "zipcode": "91012", "entries": 0},
					{"street_line": "PO Box 0", "secondary": "", "city": "La Marque", "state": "TX", "zipcode": "77568", "entries": 0},
					{"street_line": "PO Box 0", "secondary": "", "city": "Montgomery", "state": "AL", "zipcode": "36132", "entries": 0},
					{"street_line": "PO Box 0", "secondary": "", "city": "Pelion", "state": "SC", "zipcode": "29123", "entries": 0},
					{"street_line": "PO Box 0", "secondary": "", "city": "Plain Dealing", "state": "LA", "zipcode": "71064", "entries": 0},
				},
			}

		case strings.HasPrefix(search, "1042 E Center St"):
			response = map[string]interface{}{
				"suggestions": []map[string]interface{}{
					{"street_line": "1042 E Center St", "secondary": "# 1", "city": "Wallingford", "state": "CT", "zipcode": "06492", "entries": 1},
					{"street_line": "1042 E Center St", "secondary": "# 2", "city": "Wallingford", "state": "CT", "zipcode": "06492", "entries": 1},
				},
			}

		default:
			response = map[string]interface{}{
				"suggestions": []map[string]interface{}{},
			}

		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(response)
		suite.Require().NoError(err, "Mock Smarty response encoding failed")
	}))

	return mockLedgerServer
}

func (suite *IntegrationTestSuite) TestAddressAutoCompleteSuccess() {
	mockSmartyServer := SetupMockForSmarty(suite)
	defer mockSmartyServer.Close()

	config.Config.SmartyStreets.LookupUrl = mockSmartyServer.URL
	config.Config.SmartyStreets.AuthId = "test-authId"
	config.Config.SmartyStreets.AuthToken = "test-auth-token"

	userStatus := constant.PHONE_NUMBER_VERIFIED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	e := handler.NewEcho()

	params := url.Values{}

	params.Add("search", "1600 Pennsylvania")

	req := httptest.NewRequest(http.MethodPost, "/onboarding/customer/address?"+params.Encode(), nil)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err := handler.AddressAutoComplete(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *IntegrationTestSuite) RunAddressAutoCompleteForEmptyResponse(search string) {
	mockSmartyServer := SetupMockForSmarty(suite)
	defer mockSmartyServer.Close()

	config.Config.SmartyStreets.LookupUrl = mockSmartyServer.URL
	config.Config.SmartyStreets.AuthId = "test-authId"
	config.Config.SmartyStreets.AuthToken = "test-auth-token"

	userStatus := constant.PHONE_NUMBER_VERIFIED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	e := handler.NewEcho()

	params := url.Values{}
	params.Add("search", search)

	req := httptest.NewRequest(http.MethodPost, "/onboarding/customer/address?"+params.Encode(), nil)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err := handler.AddressAutoComplete(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var suggestions response.Address
	err = json.Unmarshal(rec.Body.Bytes(), &suggestions)
	suite.Require().NoError(err)
	suite.Len(suggestions.Suggestions, 0, "Expected empty suggestions array")
}

func (suite *IntegrationTestSuite) TestAddressAutoCompleteForPostOfficeAddress() {
	// Skip PO BOX addresses
	suite.RunAddressAutoCompleteForEmptyResponse("PO Box")
}

func (suite *IntegrationTestSuite) TestAddressAutoCompleteForInvalidAddress() {
	suite.RunAddressAutoCompleteForEmptyResponse("abc")
}

func (suite *IntegrationTestSuite) TestAddressAutoCompleteForInvalidSearchTextLength() {
	mockSmartyServer := SetupMockForSmarty(suite)
	defer mockSmartyServer.Close()

	config.Config.SmartyStreets.LookupUrl = mockSmartyServer.URL
	config.Config.SmartyStreets.AuthId = "test-authId"
	config.Config.SmartyStreets.AuthToken = "test-auth-token"

	userStatus := constant.PHONE_NUMBER_VERIFIED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	e := handler.NewEcho()

	params := url.Values{}

	// Dummy 130 byte text
	params.Add("search", "1234 Elm Street, Apartment 56B, Springfield, IL 62704. Please leave package at the back door before 5:00 PM. Thank you for your service")

	req := httptest.NewRequest(http.MethodPost, "/onboarding/customer/address?"+params.Encode(), nil)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err := handler.AddressAutoComplete(customContext)
	suite.Require().NotNil(err, "Handler should return an error if the search text exceeds the maximum allowed length")
	errResp, ok := err.(response.BadRequestErrors)
	suite.Require().True(ok, "Expected error of type response.BadRequestErrors")
	suite.Require().Equal("search", errResp.Errors[0].FieldName, "Expected error for field 'search'")
}
