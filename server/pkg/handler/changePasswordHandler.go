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

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// @summary ChangePassword
// @description Change password
// @tags changePassword
// @accept json
// @produce json
// @param changePasswordRequest body request.ChangePasswordRequest true "payload"
// @Success 200 "OK"
// @failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/change-password [post]
func ChangePassword(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	var requestData request.ChangePasswordRequest

	if err := c.Bind(&requestData); err != nil {
		logger.Error("Invalid request", "error", err.Error())
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(requestData.OldPassword)); err != nil {
		logger.Warn("Incorrect current password entered for user", "userId", userId)
		return response.GenerateErrResponse(constant.INVALID_CREDENTIALS_ERROR, constant.INCORRECT_PASSWORD_ENTERED_MSG, fmt.Sprintf("Incorrect current password entered for user: %s", err.Error()), http.StatusBadRequest, errtrace.Wrap(err))
	}

	if user.ResetToken != requestData.ResetToken {
		logger.Warn("Incorrect ResetToken provided for user", "userId", userId)
		return response.GenerateErrResponse(constant.INVALID_CREDENTIALS_ERROR, constant.INCORRECT_PASSWORD_ENTERED_MSG, fmt.Sprintf("Incorrect ResetToken provided for user: %s", userId), http.StatusBadRequest, errtrace.New(""))
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(requestData.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return response.InternalServerError(fmt.Sprintf("Error while generating hash of the password: %s", err.Error()), errtrace.Wrap(err))
	}

	result := db.DB.Model(&dao.MasterUserRecordDao{}).Where("id=?", userId).Updates(map[string]interface{}{
		"password":    hashedNewPassword,
		"reset_token": "",
	})
	if result.Error != nil {
		return response.InternalServerError(fmt.Sprintf("Error in updating user's state and password: %s", result.Error.Error()), errtrace.Wrap(result.Error))
	} else if result.RowsAffected <= 0 {
		logger.Error("Error in changing user's password. No record found")
		return c.NoContent(http.StatusNotFound)
	}

	logger.Info("Password updated successfully for user", "userId", userId)

	return c.NoContent(http.StatusOK)
}

type ChangePasswordResponse struct {
	IsPasswordChanged bool `json:"isPasswordChanged"`
}
