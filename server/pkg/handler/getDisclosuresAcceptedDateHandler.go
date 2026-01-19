package handler

import (
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"time"

	"github.com/labstack/echo/v4"
)

// @Summary GetDisclosuresAcceptedDate
// @Description Returns Disclosures details.
// @Tags Disclosures
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} GetDisclosuresDataResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @router /account/customer/disclosure-accepted-date [get]
func GetDisclosuresAcceptedDate(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.ErrorResponse{ErrorCode: constant.UNAUTHORIZED_ACCESS_ERROR, Message: constant.UNAUTHORIZED_ACCESS_ERROR_MSG, StatusCode: http.StatusUnauthorized, LogMessage: "Failed to get user Id from custom context"}
	}

	userId := cc.UserId

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	if user.AgreementESignSignedAt == nil || user.AgreementPrivacyNoticeSignedAt == nil || user.AgreementTermsOfServiceSignedAt == nil || user.AgreementCardAndDepositSignedAt == nil || user.AgreementDreamfiAchAuthorizationSignedAt == nil {
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: "Failed to show"}
	}

	response := GetDisclosuresDataResponse{
		AgreementESignSignedAt:                   *user.AgreementESignSignedAt,
		AgreementPrivacyNoticeSignedAt:           *user.AgreementPrivacyNoticeSignedAt,
		AgreementTermsOfServiceSignedAt:          *user.AgreementTermsOfServiceSignedAt,
		AgreementCardAndDepositSignedAt:          *user.AgreementCardAndDepositSignedAt,
		AgreementDreamfiAchAuthorizationSignedAt: *user.AgreementDreamfiAchAuthorizationSignedAt,
	}

	return c.JSON(http.StatusOK, response)
}

type GetDisclosuresDataResponse struct {
	AgreementESignSignedAt                   time.Time `json:"agreementESignSignedSignedAt" validate:"required"`
	AgreementPrivacyNoticeSignedAt           time.Time `json:"agreementPrivacyNoticeSignedAt" validate:"required"`
	AgreementTermsOfServiceSignedAt          time.Time `json:"agreementTermsOfServiceSignedAt" validate:"required"`
	AgreementCardAndDepositSignedAt          time.Time `json:"agreementCardAndDepositSignedAt" validate:"required"`
	AgreementDreamfiAchAuthorizationSignedAt time.Time `json:"agreementDreamfiAchAuthorizationSignedAt" validate:"required"`
}
