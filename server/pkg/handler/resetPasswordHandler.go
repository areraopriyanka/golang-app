package handler

import (
	"errors"
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"

	"braces.dev/errtrace"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// @summary ResetPassword
// @description Reset password
// @tags resetPassword
// @accept json
// @produce json
// @param resetPasswordRequest body request.ResetPasswordRequest true "payload"
// @Success 200 "OK"
// @Failure 400 {object} response.BadRequestErrors
// @failure 404 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /reset-password [post]
func ResetPassword(c echo.Context) error {
	logger := logging.GetEchoContextLogger(c)

	requestData := new(request.ResetPasswordRequest)

	err := c.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	// validations
	if err := c.Validate(requestData); err != nil {
		return err
	}

	var user dao.MasterUserRecordDao
	result := db.DB.Where("reset_token= ?", requestData.ResetToken).Take(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, StatusCode: http.StatusNotFound, LogMessage: "Record not found in DB", MaybeInnerError: errtrace.Wrap(result.Error)}
		}
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("DB Error: %s", result.Error.Error()), MaybeInnerError: errtrace.New("")}
	}

	// Reset password only for ACTIVE user

	if user.UserStatus != constant.ACTIVE {
		response := response.ErrorResponse{
			ErrorCode:       "USER_NOT_ACTIVE",
			Message:         "User must be active to reset their password",
			StatusCode:      http.StatusPreconditionFailed,
			LogMessage:      "User must be active to reset their password",
			MaybeInnerError: errtrace.New(""),
		}

		return response
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestData.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error while generating hash of the password: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	updatedData := map[string]interface{}{
		"reset_token": "",
		"password":    hashedPassword,
	}
	// clear resetToken and update password
	result = db.DB.Model(dao.MasterUserRecordDao{}).Where("reset_token=?", requestData.ResetToken).Updates(updatedData)
	// Handle other DB error's
	if result.Error != nil {
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("DB Error: %s", result.Error.Error()), MaybeInnerError: errtrace.Wrap(result.Error)}
	}
	// If record not present
	if result.RowsAffected == 0 {
		logger.Error("Record not found. Invalid Reset Token")
		return c.NoContent(http.StatusNotFound)
	}
	return c.NoContent(http.StatusOK)
}
