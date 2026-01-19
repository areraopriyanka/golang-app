package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/constant"
	"process-api/pkg/handler"
	"process-api/pkg/model/request"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestCreateUserAccount() {
	defer SetupMockForLedger(suite).Close()
	e := handler.NewEcho()

	request := request.CreateUserAccountRequest{
		Email:    "testuser123@gmail.com",
		Password: "Test@123",
	}

	requestBody, err := json.Marshal(request)
	suite.Require().NoError(err, "Failed to marshall request")

	req := httptest.NewRequest(http.MethodPost, "/onboarding/customer/account", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err = handler.CreateUserAccount(c)

	suite.Require().NoError(err, "Handler should not return an error")

	// Verify the onboarding JWT received in the response header
	onboardingToken := rec.Header().Get("Authorization")
	suite.Require().NotEmpty(onboardingToken, "Onboarding JWT is missing in the response header")
	suite.ValidateOnboardingTokenClaims(onboardingToken)
}

func (suite *IntegrationTestSuite) ValidateOnboardingTokenClaims(token string) {
	claims := security.GetClaimsFromToken(token)
	suite.Require().NotNil(claims, "Failed to retrieve jwt claims")
	suite.Require().NotEmpty(claims.Subject, "User ID is missing in JWT")
	suite.Require().NotEmpty(claims.UserState, "User state is missing in JWT")
	suite.Require().Equal(constant.ONBOARDING, claims.UserState, "Invalid user state")
}
