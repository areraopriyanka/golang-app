package test

import (
	"net/http"
	"net/http/httptest"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestAcceptMembership() {
	e := handler.NewEcho()

	userStatus := constant.KYC_PASS
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})
	req := httptest.NewRequest(http.MethodPost, "/onboarding/customer/membership", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/onboarding/customer/membership")

	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err := handler.AcceptMembership(customContext)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var user dao.MasterUserRecordDao
	result := suite.TestDB.Where("id=?", testUser.Id).Find(&user)
	suite.Require().NoError(result.Error, "failed to fetch user record")
	suite.Require().Equal(constant.MEMBERSHIP_ACCEPTED, user.UserStatus, "user's status should be updated to MEMBERSHIP_ACCEPTED")
}
