package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"time"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @summary AddPersonalDetails
// @description Updates a customer's first name, last name, suffix and DOB
// @tags onboarding
// @accept json
// @param AddPersonalDetailsRequest body request.AddPersonalDetailsRequest true "AddPersonalDetails payload"
// @param Authorization header string true "Bearer token for user authentication"
// @success 200 "OK"
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /onboarding/customer/details [post]
func AddPersonalDetails(c echo.Context) error {
	var requestData request.AddPersonalDetailsRequest
	cc, ok := c.(*security.OnboardingUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId
	logger := logging.GetEchoContextLogger(c)

	err := c.Bind(&requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	dob, err := time.Parse("01/02/2006", requestData.DOB)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	user, errResponse := dao.RequireUserWithState(
		userId, "AGREEMENTS_REVIEWED", "AGE_VERIFICATION_PASSED", "PHONE_VERIFICATION_OTP_SENT",
	)
	if errResponse != nil {
		return errResponse
	}

	updateData := dao.MasterUserRecordDao{
		FirstName: requestData.FirstName,
		LastName:  requestData.LastName,
		Suffix:    requestData.Suffix,
		DOB:       dob,
	}
	if user.UserStatus == constant.AGREEMENTS_REVIEWED {
		updateData.UserStatus = constant.AGE_VERIFICATION_PASSED
	}

	result := db.DB.Model(&dao.MasterUserRecordDao{}).Where("id=?", userId).Updates(updateData)
	if result.Error != nil {
		logger.Error("Error while updating user", "error", result.Error)
		return response.InternalServerError(fmt.Sprintf("Error while updating user: %s", result.Error), errtrace.Wrap(result.Error))
	}

	return c.NoContent(http.StatusOK)
}
