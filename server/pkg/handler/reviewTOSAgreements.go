package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/clock"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/resource/agreements"
	"process-api/pkg/security"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @Summary ReviewTOSAgreements
// @Description
// @Accept json
// @Produce json
// @Param reviewTOSAgreementsRequest body ReviewTOSAgreementsRequest true "ReviewTOSAgreementsRequest"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 "OK"
// @Header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /onboarding/customer/agreements/terms-of-service [post]
func ReviewTOSAgreements(c echo.Context) error {
	cc, ok := c.(*security.OnboardingUserContext)
	if !ok {
		return response.ErrorResponse{
			ErrorCode:       constant.UNAUTHORIZED_ACCESS_ERROR,
			Message:         constant.UNAUTHORIZED_ACCESS_ERROR_MSG,
			StatusCode:      http.StatusUnauthorized,
			LogMessage:      "Invalid type of custom context",
			MaybeInnerError: errtrace.New(""),
		}
	}

	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	var requestData ReviewTOSAgreementsRequest

	if err := c.Bind(&requestData); err != nil {
		logger.Error("Invalid request", "error", err.Error())
		return response.BadRequestErrors{
			Errors: []response.BadRequestError{
				{Error: err.Error()},
			},
		}
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	user, errResponse := dao.RequireUserWithState(
		userId, "USER_CREATED", "AGREEMENTS_REVIEWED", "AGE_VERIFICATION_PASSED", "PHONE_VERIFICATION_OTP_SENT",
	)
	if errResponse != nil {
		return errResponse
	}

	errors := []response.BadRequestError{}

	if requestData.ESignHash != agreements.Agreements.ESign.Hash {
		errors = append(errors, response.BadRequestError{Error: "invalid", FieldName: "eSign"})
	}

	if requestData.PrivacyNoticeHash != agreements.Agreements.PrivacyNotice.Hash {
		errors = append(errors, response.BadRequestError{Error: "invalid", FieldName: "privacyNotice"})
	}

	if requestData.TermsOfServiceHash != agreements.Agreements.TermsOfService.Hash {
		errors = append(errors, response.BadRequestError{Error: "invalid", FieldName: "termsOfService"})
	}

	if len(errors) != 0 {
		return response.BadRequestErrors{
			Errors: errors,
		}
	}

	now := clock.Now()

	updateData := dao.MasterUserRecordDao{
		AgreementESignHash:              &requestData.ESignHash,
		AgreementPrivacyNoticeHash:      &requestData.PrivacyNoticeHash,
		AgreementTermsOfServiceHash:     &requestData.TermsOfServiceHash,
		AgreementESignSignedAt:          &now,
		AgreementPrivacyNoticeSignedAt:  &now,
		AgreementTermsOfServiceSignedAt: &now,
	}
	if user.UserStatus == constant.USER_CREATED {
		updateData.UserStatus = constant.AGREEMENTS_REVIEWED
	}

	result := db.DB.Model(user).Updates(updateData)
	if result.Error != nil {
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Error saving user's agreement status: %s", result.Error.Error()),
			MaybeInnerError: errtrace.Wrap(result.Error),
		}
	}
	return c.NoContent(http.StatusOK)
}

type ReviewTOSAgreementsRequest struct {
	ESignHash          string `json:"eSignHash" validate:"required"`
	PrivacyNoticeHash  string `json:"privacyNoticeHash" validate:"required"`
	TermsOfServiceHash string `json:"termsOfServiceHash" validate:"required"`
}
