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

func (suite *IntegrationTestSuite) TestReviewCardAgreements() {
	unfreeze := clock.FreezeNow()
	defer unfreeze()

	e := handler.NewEcho()

	userStatus := constant.MEMBERSHIP_ACCEPTED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	requestBody := handler.ReviewCardAgreementsRequest{
		CardAndDepositHash:          "card-and-deposit-hash",
		DreamfiAchAuthorizationHash: "dreamfi-ach-authorization-hash",
	}

	body, err := json.Marshal(requestBody)
	suite.Require().NoError(err, "Failed to marshall request")

	req := httptest.NewRequest(http.MethodPost, "/customer/agreements/card-agreements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err = handler.ReviewCardAgreements(customContext)
	suite.NoError(err, "Handler should not return an error")

	var user dao.MasterUserRecordDao
	err = suite.TestDB.Model(&dao.MasterUserRecordDao{}).Where("id=?", testUser.Id).Find(&user).Error
	suite.Require().NoError(err, "Failed to fetch updated user data")
	suite.Equal("card-and-deposit-hash", *user.AgreementCardAndDepositHash)
	suite.Equal("dreamfi-ach-authorization-hash", *user.AgreementDreamfiAchAuthorizationHash)
	suite.Nil(user.AgreementESignHash)
	suite.Nil(user.AgreementPrivacyNoticeHash)
	suite.Nil(user.AgreementTermsOfServiceHash)
	suite.Equal(constant.CARD_AGREEMENTS_REVIEWED, user.UserStatus)

	suite.WithinDuration(clock.Now(), *user.AgreementCardAndDepositSignedAt, 0)
	suite.WithinDuration(clock.Now(), *user.AgreementDreamfiAchAuthorizationSignedAt, 0)
	suite.Nil(user.AgreementESignSignedAt)
	suite.Nil(user.AgreementPrivacyNoticeSignedAt)
	suite.Nil(user.AgreementTermsOfServiceSignedAt)
}

func (suite *IntegrationTestSuite) TestReviewCardAgreementsForInvalidRequest() {
	e := handler.NewEcho()

	userStatus := constant.MEMBERSHIP_ACCEPTED
	testUser := suite.createTestUser(PartialMasterUserRecordDao{UserStatus: &userStatus})

	requestBody := handler.ReviewCardAgreementsRequest{
		CardAndDepositHash:          "invalid-card-and-deposit-hash",
		DreamfiAchAuthorizationHash: "invalid-dreamfi-ach-authorization-hash",
	}

	body, err := json.Marshal(requestBody)
	suite.Require().NoError(err, "Failed to marshall request")

	req := httptest.NewRequest(http.MethodPost, "/onboarding/customer/agreements/card-agreements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	customContext := security.GenerateOnboardingUserContext(testUser.Id, c)

	err = handler.ReviewCardAgreements(customContext)
	suite.Require().NotNil(err, "Handler should return an error")

	errorResponse := err.(response.BadRequestErrors)
	suite.Equal(errorResponse.Errors, []response.BadRequestError{
		{FieldName: "cardAndDeposit", Error: "invalid"},
		{FieldName: "dreamfiAchAuthorization", Error: "invalid"},
	})

	var user dao.MasterUserRecordDao
	err = suite.TestDB.Model(&dao.MasterUserRecordDao{}).Where("id=?", testUser.Id).Find(&user).Error
	suite.Require().NoError(err, "Failed to fetch updated user data")
	suite.Nil(user.AgreementCardAndDepositHash)
	suite.Nil(user.AgreementDreamfiAchAuthorizationHash)
	suite.Nil(user.AgreementESignHash)
	suite.Nil(user.AgreementPrivacyNoticeHash)
	suite.Nil(user.AgreementTermsOfServiceHash)
	suite.Equal(constant.MEMBERSHIP_ACCEPTED, user.UserStatus)

	suite.Nil(user.AgreementCardAndDepositSignedAt)
	suite.Nil(user.AgreementDreamfiAchAuthorizationSignedAt)
	suite.Nil(user.AgreementESignSignedAt)
	suite.Nil(user.AgreementPrivacyNoticeSignedAt)
	suite.Nil(user.AgreementTermsOfServiceSignedAt)
}
