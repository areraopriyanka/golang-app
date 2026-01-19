package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"github.com/labstack/echo/v4"
)

func (suite *IntegrationTestSuite) TestPlaidCreateLinkTokeniOSSuccess() {
	h := suite.newHandler()
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)

	e := handler.NewEcho()
	tokenRequest := handler.PlaidLinkTokenRequest{Platform: "ios"}
	requestBody, _ := json.Marshal(tokenRequest)
	e.POST("/account/plaid/link/token", func(c echo.Context) error {
		cc := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)
		return h.PlaidCreateLinkToken(cc)
	})

	req := httptest.NewRequest(http.MethodPost, "/account/plaid/link/token", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	var body handler.PlaidLinkTokenResponse
	suite.NoError(json.NewDecoder(bytes.NewReader(rec.Body.Bytes())).Decode(&body))
	suite.Equal("link-sandbox-5c2ad447-5e16-4905-90f1-317cfc0a7067", body.LinkToken, "linkToken must not be empty, and must match mocked value")
}

func (suite *IntegrationTestSuite) TestPlaidCreateLinkTokenAndroidSuccess() {
	h := suite.newHandler()
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)

	e := handler.NewEcho()
	tokenRequest := handler.PlaidLinkTokenRequest{Platform: "android"}
	requestBody, _ := json.Marshal(tokenRequest)
	e.POST("/account/plaid/link/token", func(c echo.Context) error {
		cc := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)
		return h.PlaidCreateLinkToken(cc)
	})

	req := httptest.NewRequest(http.MethodPost, "/account/plaid/link/token", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	var body handler.PlaidLinkTokenResponse
	suite.NoError(json.NewDecoder(bytes.NewReader(rec.Body.Bytes())).Decode(&body))
	suite.Equal("link-sandbox-5c2ad447-5e16-4905-90f1-317cfc0a7067", body.LinkToken, "linkToken must not be empty, and must match mocked value")
}

func (suite *IntegrationTestSuite) TestPlaidCreateLinkTokenFailure() {
	h := suite.newHandler()
	user := suite.createTestUser(PartialMasterUserRecordDao{})
	userPublicKey := suite.createUserPublicKeyRecord(user.Id)

	e := handler.NewEcho()
	e.POST("/account/plaid/link/token", func(c echo.Context) error {
		cc := security.GenerateLoggedInRegisteredUserContext(user.Id, userPublicKey.PublicKey, c)
		return h.PlaidCreateLinkToken(cc)
	})

	req := httptest.NewRequest(http.MethodPost, "/account/plaid/link/token", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	suite.Equal(http.StatusBadRequest, rec.Code, "Response code should be 400 for missing Platform in body")

	var errorResponse response.BadRequestErrors
	err := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	suite.Require().NoError(err, "Failed to parse response body")
	suite.Equal(errorResponse.Errors, []response.BadRequestError{{FieldName: "platform", Error: "required"}}, "platform is required")
}
