package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/clock"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
)

func (suite *IntegrationTestSuite) TestReviewTOSAgreements() {
	unfreeze := clock.FreezeNow()
	defer unfreeze()

	e := handler.NewEcho()

	userStatus := constant.USER_CREATED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	requestBody := handler.ReviewTOSAgreementsRequest{
		ESignHash:          "e-sign-hash",
		PrivacyNoticeHash:  "privacy-notice-hash",
		TermsOfServiceHash: "terms-of-service-hash",
	}

	body, err := json.Marshal(requestBody)
	suite.Require().NoError(err, "Failed to marshall request")

	req := httptest.NewRequest(http.MethodPost, "/customer/agreements/terms-of-service", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err = handler.ReviewTOSAgreements(customContext)
	suite.NoError(err, "Handler should not return an error")

	var user dao.MasterUserRecordDao
	err = suite.TestDB.Model(&dao.MasterUserRecordDao{}).Where("id=?", testUser.Id).Find(&user).Error
	suite.Require().NoError(err, "Failed to fetch updated user data")
	suite.Nil(user.AgreementCardAndDepositHash)
	suite.Nil(user.AgreementDreamfiAchAuthorizationHash)
	suite.Equal("e-sign-hash", *user.AgreementESignHash)
	suite.Equal("privacy-notice-hash", *user.AgreementPrivacyNoticeHash)
	suite.Equal("terms-of-service-hash", *user.AgreementTermsOfServiceHash)
	suite.Equal(constant.AGREEMENTS_REVIEWED, user.UserStatus)

	suite.Nil(user.AgreementCardAndDepositSignedAt)
	suite.Nil(user.AgreementDreamfiAchAuthorizationSignedAt)
	suite.WithinDuration(clock.Now(), *user.AgreementESignSignedAt, 0)
	suite.WithinDuration(clock.Now(), *user.AgreementPrivacyNoticeSignedAt, 0)
	suite.WithinDuration(clock.Now(), *user.AgreementTermsOfServiceSignedAt, 0)
}

func (suite *IntegrationTestSuite) TestReviewTOSAgreementsForInvalidRequest() {
	e := handler.NewEcho()

	userStatus := constant.USER_CREATED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	requestBody := handler.ReviewTOSAgreementsRequest{
		ESignHash:          "invalid-e-sign-hash",
		PrivacyNoticeHash:  "invalid-privacy-notice-hash",
		TermsOfServiceHash: "invalid-terms-of-service-hash",
	}

	body, err := json.Marshal(requestBody)
	suite.Require().NoError(err, "Failed to marshall request")

	req := httptest.NewRequest(http.MethodPost, "/onboarding/customer/agreements/terms-of-service", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err = handler.ReviewTOSAgreements(customContext)
	suite.Require().NotNil(err, "Handler should return an error")

	errorResponse := err.(response.BadRequestErrors)
	suite.Equal(errorResponse.Errors, []response.BadRequestError{
		{FieldName: "eSign", Error: "invalid"},
		{FieldName: "privacyNotice", Error: "invalid"},
		{FieldName: "termsOfService", Error: "invalid"},
	})

	var user dao.MasterUserRecordDao
	err = suite.TestDB.Model(&dao.MasterUserRecordDao{}).Where("id=?", testUser.Id).Find(&user).Error
	suite.Require().NoError(err, "Failed to fetch updated user data")
	suite.Nil(user.AgreementCardAndDepositHash)
	suite.Nil(user.AgreementDreamfiAchAuthorizationHash)
	suite.Nil(user.AgreementESignHash)
	suite.Nil(user.AgreementPrivacyNoticeHash)
	suite.Nil(user.AgreementTermsOfServiceHash)
	suite.Equal(constant.USER_CREATED, user.UserStatus)

	suite.Nil(user.AgreementCardAndDepositSignedAt)
	suite.Nil(user.AgreementDreamfiAchAuthorizationSignedAt)
	suite.Nil(user.AgreementESignSignedAt)
	suite.Nil(user.AgreementPrivacyNoticeSignedAt)
	suite.Nil(user.AgreementTermsOfServiceSignedAt)
}
