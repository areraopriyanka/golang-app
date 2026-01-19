package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"process-api/pkg/clock"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/utils"
	"process-api/pkg/validators"
	"time"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// @summary UpdateCustomerDOB
// @description Updating customer's date of birth and setting customer's age verification status
// @tags onboarding
// @accept json
// @produce json
// @Param userId path string true "User ID"
// @param updateCustomerDOBRequest body request.UpdateCustomerDOBRequest true "UpdateCustomerDOB payload"
// @Success 200 {object} response.UpdateDOBResponse
// @failure 500 {object} response.ErrorResponse
// @failure 400 {object} response.ErrorResponse
// @failure 404 "Not found"
// @router /onboarding/customer/{userId}/dob [post]
func UpdateCustomerDOB(c echo.Context) error {
	var request request.UpdateCustomerDOBRequest
	userId := c.Param("userId")

	logger := logging.GetEchoContextLogger(c)

	err := json.NewDecoder(c.Request().Body).Decode(&request)
	if err != nil {
		logger.Error("Failed to decode request body", "error", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	if err := validators.ValidateStruct(request); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			switch e.Tag() {
			case "required":
				return response.ErrorResponse{ErrorCode: constant.DATE_OF_BIRTH_REQUIRED, Message: constant.DATE_OF_BIRTH_REQUIRED_MSG, StatusCode: http.StatusBadRequest, LogMessage: constant.DATE_OF_BIRTH_REQUIRED_MSG, MaybeInnerError: errtrace.Wrap(err)}

			case "validateDate":
				return response.ErrorResponse{ErrorCode: constant.INVALID_DATE_OF_BIRTH, Message: constant.INVALID_DATE_OF_BIRTH_MSG, StatusCode: http.StatusBadRequest, LogMessage: fmt.Sprintf("Invalid DOB. Error: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
			}
		}
	}
	dob, _ := time.Parse("01/02/2006", request.DOB)
	// Check user's age is valid or not
	isValidAge := utils.IsValidAge(dob, func() time.Time {
		return clock.Now().UTC()
	})
	var ageStatus string
	if isValidAge {
		ageStatus = constant.AGE_VERIFICATION_PASSED
	} else {
		ageStatus = constant.AGE_VERIFICATION_FAILED
	}

	// Update user's state
	updateData := dao.MasterUserRecordDao{
		UserStatus: ageStatus,
		DOB:        dob,
	}
	result := db.DB.Model(&dao.MasterUserRecordDao{}).Where("id=?", userId).Updates(updateData)
	// Handle DB errors
	if result.Error != nil {
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, Message: constant.INTERNAL_SERVER_ERROR_MSG, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error in updating user's state and DOB: %s", result.Error.Error()), MaybeInnerError: errtrace.Wrap(result.Error)}
	} else if result.RowsAffected <= 0 {
		logger.Error("Error in updating user's state and DOB. No record found")
		return c.NoContent(http.StatusNotFound)
	}

	response := response.UpdateDOBResponse{
		UserStatus: ageStatus,
	}

	logger.Info("DOB updated successfully for user", "userId", userId)
	return c.JSON(http.StatusOK, response)
}
