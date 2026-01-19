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
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

// @summary VerifyResetPasswordOTP
// @description Verifies the reset password OTP
// @tags resetPassword
// @accept json
// @produce json
// @param verifyResetPasswordOtpRequest body request.ChallengeOtpRequest true "payload"
// @Success 200 {object} response.VerifyResetPasswordOtpResponse
// @Failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /reset-password/verify-otp [post]
func VerifyResetPasswordOTP(c echo.Context) error {
	logger := logging.GetEchoContextLogger(c)

	request := new(request.ChallengeOtpRequest)

	err := c.Bind(request)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	// validations
	if err := c.Validate(request); err != nil {
		return err
	}

	var otpRecord dao.MasterUserOtpDao
	result := db.DB.Where("otp_id=?", request.OtpId).Find(&otpRecord)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return response.BadRequestErrors{
				Errors: []response.BadRequestError{
					{
						FieldName: "invalid_otp",
						Error:     constant.INVALID_OTP_MSG,
					},
				},
			}
		}
		logger.Error("DB error", "error", result.Error.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("DB error: %s", result.Error.Error()),
			MaybeInnerError: errtrace.Wrap(result.Error),
		}
	}

	_, errResponse := dao.RequireUserWithState(otpRecord.UserId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	apiPath := "/reset-password/send-otp"
	userOtp, err := utils.VerifyOTP(request.OtpId, request.Otp, apiPath)
	if err != nil {
		logging.Logger.Error("error verifying otp for forgot password", "error", err.Error())
		return response.GenerateOTPErrResponse(errtrace.Wrap(err))
	}

	resetToken, err := resetToken(userOtp.UserId)
	if err != nil {
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("DB Error: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	logger.Info("OTP verified successfully", "otpId", request.OtpId)
	return c.JSON(http.StatusOK, response.VerifyResetPasswordOtpResponse{ResetToken: resetToken})
}

func resetToken(userId string) (string, error) {
	// update ResetToken in master_user_records table
	resetToken := uuid.New().String()
	updateUser := map[string]interface{}{
		"reset_token": resetToken,
		"password":    nil,
	}

	result := db.DB.Model(dao.MasterUserRecordDao{}).Where("id=?", userId).Updates(updateUser)
	// Handle other DB error's
	if result.Error != nil {
		return "", errtrace.Wrap(result.Error)
	}
	// Record not present error
	if result.RowsAffected == 0 {
		return "", errtrace.Wrap(fmt.Errorf("user record not found"))
	}

	return resetToken, nil
}
