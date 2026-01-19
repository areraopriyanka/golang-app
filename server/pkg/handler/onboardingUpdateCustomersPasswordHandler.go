package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/validators"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// @summary UpdateCustomerPassword
// @description Updating customer's password and setting customer's status
// @tags onboarding
// @accept json
// @produce json
// @Param userId path string true "User ID"
// @param updateCustomerPasswordRequest body request.UpdateCustomerPasswordRequest true "UpdateCustomerPasswordRequest payload"
// @Success 200 "OK"
// @failure 500 {object} response.ErrorResponse
// @failure 400 {object} response.ErrorResponse
// @failure 404 "Not found"
// @router /onboarding/customer/{userId}/password [post]
func UpdateCustomerPassword(c echo.Context) error {
	var request request.UpdateCustomerPasswordRequest
	userId := c.Param("userId")

	logger := logging.GetEchoContextLogger(c)

	if err := json.NewDecoder(c.Request().Body).Decode(&request); err != nil {
		logger.Error("Failed to decode request body", "error", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	if err := validators.ValidateStruct(request); err != nil {
		// Validation errors returned by validator are in the order of the fields as they are defined in the struct.
		validationErrors := err.(validator.ValidationErrors)
		// Converting validation errors to string to send them in response
		errors := validators.ConvertValidationErrorsToString(validationErrors)
		return response.ErrorResponse{ErrorCode: constant.INVALID_INPUT, Message: errors, StatusCode: http.StatusBadRequest, LogMessage: errors, MaybeInnerError: errtrace.Wrap(err)}
	}

	// Generates Hash of the user's password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, Message: constant.INTERNAL_SERVER_ERROR_MSG, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error while generating hash of the password: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	updateData := dao.MasterUserRecordDao{
		Password:   hashedPassword,
		UserStatus: constant.PASSWORD_SET,
	}

	result := db.DB.Model(&dao.MasterUserRecordDao{}).Where("id=?", userId).Updates(updateData)
	// Handle DB errors's
	if result.Error != nil {
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, Message: constant.INTERNAL_SERVER_ERROR_MSG, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error in updating user's state and password: %s", result.Error.Error()), MaybeInnerError: errtrace.Wrap(result.Error)}
	} else if result.RowsAffected <= 0 {
		logger.Error("Error in updating user's state and password. No record found")
		return c.NoContent(http.StatusNotFound)
	}

	logger.Info("Password updated successfully for user", "userId", userId)
	return c.NoContent(http.StatusOK)
}
