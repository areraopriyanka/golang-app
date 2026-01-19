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

// @Summary ReviewCardAgreements
// @Description
// @Accept json
// @Produce json
// @Param reviewCardAgreementsRequest body ReviewCardAgreementsRequest true "ReviewCardAgreementsRequest"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 "OK"
// @Header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /onboarding/customer/agreements/card-agreements [post]
func ReviewCardAgreements(c echo.Context) error {
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

	var requestData ReviewCardAgreementsRequest

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
		userId, constant.MEMBERSHIP_ACCEPTED,
	)
	if errResponse != nil {
		return errResponse
	}

	errors := []response.BadRequestError{}

	if requestData.CardAndDepositHash != agreements.Agreements.CardAndDeposit.Hash {
		errors = append(errors, response.BadRequestError{Error: "invalid", FieldName: "cardAndDeposit"})
	}

	if requestData.DreamfiAchAuthorizationHash != agreements.Agreements.DreamfiAchAuthorization.Hash {
		errors = append(errors, response.BadRequestError{Error: "invalid", FieldName: "dreamfiAchAuthorization"})
	}

	if len(errors) != 0 {
		return response.BadRequestErrors{
			Errors: errors,
		}
	}

	now := clock.Now()

	updateData := dao.MasterUserRecordDao{
		AgreementCardAndDepositHash:              &requestData.CardAndDepositHash,
		AgreementDreamfiAchAuthorizationHash:     &requestData.DreamfiAchAuthorizationHash,
		AgreementCardAndDepositSignedAt:          &now,
		AgreementDreamfiAchAuthorizationSignedAt: &now,
		UserStatus:                               constant.CARD_AGREEMENTS_REVIEWED,
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

type ReviewCardAgreementsRequest struct {
	CardAndDepositHash          string `json:"cardAndDepositHash" validate:"required"`
	DreamfiAchAuthorizationHash string `json:"dreamfiAchAuthorizationHash" validate:"required"`
}
