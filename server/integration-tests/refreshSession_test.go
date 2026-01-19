package test

import (
	"net/http"
	"net/http/httptest"
	"process-api/pkg/constant"
	"process-api/pkg/handler"
	"process-api/pkg/security"

	"github.com/labstack/echo/v4"
)

func (suite *IntegrationTestSuite) TestRefreshSessionOnboarding() {
	e := echo.New()
	userStatus := constant.USER_CREATED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})
	req := httptest.NewRequest(http.MethodGet, "/refresh-session", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)
	err := handler.RefreshSession(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *IntegrationTestSuite) TestRefreshSessionOnboarded() {
	userId := suite.SetupTestData()
	e := handler.NewEcho()
	userPublicKey := suite.createUserPublicKeyRecord(userId)
	req := httptest.NewRequest(http.MethodGet, "/refresh-session", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	customContext := security.GenerateLoggedInRegisteredUserContext(userId, userPublicKey.PublicKey, c)
	err := handler.RefreshSession(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *IntegrationTestSuite) TestRefreshSessionUnregistered() {
	e := handler.NewEcho()
	testUser := suite.createTestUser(PartialMasterUserRecordDao{})
	req := httptest.NewRequest(http.MethodGet, "/refresh-session", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	customContext := security.GenerateLoggedInUnregisteredUserContext(testUser.Id, c)
	err := handler.RefreshSession(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *IntegrationTestSuite) TestRefreshSessionRecoverOnboarding() {
	e := handler.NewEcho()
	userStatus := constant.USER_CREATED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})
	req := httptest.NewRequest(http.MethodGet, "/refresh-session", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	customContext := security.GenerateRecoverOnboardingUserContext(testUser.Id, c)
	err := handler.RefreshSession(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *IntegrationTestSuite) TestRefreshSessionUnauthorized() {
	e := handler.NewEcho()
	e.GET("/refresh-session", func(c echo.Context) error {
		// No authentication context provided
		return handler.RefreshSession(c)
	})
	req := httptest.NewRequest(http.MethodGet, "/refresh-session", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	suite.Require().Equal(http.StatusUnauthorized, rec.Code, "Expected 401 Unauthorized")
}
