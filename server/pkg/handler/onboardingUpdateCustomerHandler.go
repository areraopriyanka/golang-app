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

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @summary UpdateCustomer
// @description Updating customer's firstName, lastName, email and suffix
// @tags onboarding
// @accept json
// @produce json
// @Param userId path string true "User ID"
// @param updateCustomerRequest body request.CreateOrUpdateOnboardingUserRequest true "UpdateCustomerRequest payload"
// @Success 200 "OK"
// @failure 500 {object} response.ErrorResponse
// @failure 400 {object} response.BadRequestErrors
// @failure 404 "Not Found"
// @router /onboarding/customer/{userId} [post]
func UpdateCustomer(c echo.Context) error {
	request := new(request.CreateOrUpdateOnboardingUserRequest)
	userId := c.Param("userId")

	logger := logging.GetEchoContextLogger(c)

	if err := c.Bind(request); err != nil {
		return response.BadRequestInvalidBody
	}

	// validations
	if err := c.Validate(request); err != nil {
		return err
	}

	// Used map[string]interface{} instead of struct in the Updates() to update fields with zero values(suffix) as well
	result := db.DB.Model(&dao.MasterUserRecordDao{}).Where("id=?", userId).Updates(map[string]interface{}{
		"first_name": request.FirstName,
		"last_name":  request.LastName,
		"email":      request.Email,
		"suffix":     request.Suffix,
	})

	if result.Error != nil {
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, Message: constant.INTERNAL_SERVER_ERROR_MSG, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error in updating user's data: %s", result.Error.Error()), MaybeInnerError: errtrace.Wrap(result.Error)}
	} else if result.RowsAffected <= 0 {
		logger.Error("Error in updating user's data. No record found")
		return c.NoContent(http.StatusNotFound)
	}

	logger.Info("Data updated successfully for user", "userId", userId)
	return c.NoContent(http.StatusOK)
}
