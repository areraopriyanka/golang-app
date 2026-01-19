package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"strconv"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// @summary SendChangePasswordOTP
// @description Sends a change password OTP to the user via sms
// @tags sendChangePasswordOtp
// @accept json
// @produce json
// @param CreateChangePasswordOtpRequest body CreateChangePasswordOtpRequest true "Create change password pin OTP payload"
// @success 200 {object} response.OtpResponseWithMaskedNumber "Successful otp issuance for logged in user"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @router /account/change-password/send-otp [post]
func SendChangePasswordOTP(c echo.Context) error {
	logger := logging.GetEchoContextLogger(c)

	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	var requestData CreateChangePasswordOtpRequest
	var user dao.MasterUserRecordDao

	if err := c.Bind(&requestData); err != nil {
		logging.Logger.Error("Invalid request", "error", err.Error())
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	err := db.DB.Where("id = ?", userId).First(&user).Error
	if err != nil {
		logger.Error("User record not found")
		return c.NoContent(http.StatusNotFound)
	}

	otp, err := utils.GenerateOTP(nil)
	if err != nil {
		logger.Error("Error generating OTP", "error", err.Error())
		return response.GenerateErrResponse(constant.ERROR_IN_GENERATING_OTP, constant.OTP_GENERATING_ERROR_MSG, err.Error(), http.StatusInternalServerError, errtrace.Wrap(err))
	}

	switch requestData.Type {
	case constant.CALL:
		twimlUrl := fmt.Sprintf("%svoice-xml?otp=%s", config.Config.Server.BaseUrl, otp)
		err = utils.MakeCall(user.MobileNo, config.Config.Twilio.From, twimlUrl)
		if err != nil {
			return utils.HandleTwilioError(errtrace.Wrap(err))
		}

	case constant.SMS:
		expirationInMinutes := strconv.Itoa(config.Config.Otp.OtpExpiryDuration / 60000)
		body := fmt.Sprintf("Use this One Time Password %s to verify your phone number. This One Time Password will be valid for the next %s minute(s). Do not share this with anyone.", otp, expirationInMinutes)

		err = utils.SendSMS(user.MobileNo, body, config.Config.Twilio.From)
		if err != nil {
			return utils.HandleTwilioError(errtrace.Wrap(err))
		}

	default:
		logger.Error("user requested an OTP type that is not supported", "type", requestData.Type)
		return c.NoContent(http.StatusBadRequest)
	}

	userOtpRecord := dao.MasterUserOtpDao{
		OtpId:     uuid.New().String(),
		UserId:    userId,
		OtpType:   requestData.Type,
		Otp:       otp,
		OtpStatus: constant.OTP_SENT,
		ApiPath:   "/account/change-password/send-otp",
		MobileNo:  user.MobileNo,
		Email:     user.Email,
		IP:        c.RealIP(),
		CreatedAt: clock.Now(),
	}

	otpResult := db.DB.Select("otp_id", "user_id", "otp_type", "otp", "otp_status", "api_path", "mobile_no", "email", "ip", "created_at").Create(&userOtpRecord)
	if otpResult.Error != nil {
		return response.InternalServerError("Failed to generate OTP", errtrace.Wrap(otpResult.Error))
	}

	return c.JSON(http.StatusOK, response.OtpResponseWithMaskedNumber{
		OtpId:             userOtpRecord.OtpId,
		OtpExpiryDuration: config.Config.Otp.OtpExpiryDuration,
		MaskedMobileNo:    "XXX-XXX-" + user.MobileNo[len(user.MobileNo)-4:],
	})
}

// @summary ChallengeChangePasswordOTP
// @description Challenges the change password OTP
// @tags changePasswordOtp
// @accept json
// @produce json
// @param challengeChangePasswordOtpRequest body request.ChallengeOtpRequest true "payload"
// @Success 200 {object} ChallengeChangePasswordOtpResponse
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/change-password/verify-otp [post]
func ChallengeChangePasswordOTP(c echo.Context) error {
	logger := logging.GetEchoContextLogger(c)

	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	var requestData request.ChallengeOtpRequest

	if err := c.Bind(&requestData); err != nil {
		logging.Logger.Error("Invalid request", "error", err.Error())
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	apiPath := "/account/change-password/send-otp"
	userOtp, err := utils.VerifyOTP(requestData.OtpId, requestData.Otp, apiPath)
	if err != nil {
		logging.Logger.Error("error verifying otp for change password", "error", err.Error())
		return response.GenerateOTPErrResponse(errtrace.Wrap(err))
	}

	resetToken, err := changePasswordResetToken(userOtp.UserId)
	if err != nil {
		return response.InternalServerError(fmt.Sprintf("DB Error: %s", err.Error()), errtrace.Wrap(err))
	}

	logger.Info("Change Password OTP verified successfully for user", "userId", userId)
	return c.JSON(http.StatusOK, ChallengeChangePasswordOtpResponse{ResetToken: resetToken})
}

type CreateChangePasswordOtpRequest struct {
	Type string `json:"type" validate:"required,oneof=SMS CALL"`
}

type ChallengeChangePasswordOtpResponse struct {
	ResetToken string `json:"resetToken" validate:"required"`
}

func changePasswordResetToken(userId string) (string, error) {
	// update ResetToken in master_user_records table
	resetToken := uuid.New().String()
	updateUser := map[string]interface{}{
		"reset_token": resetToken,
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
