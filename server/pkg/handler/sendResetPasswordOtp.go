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
	"process-api/pkg/model/response"
	"process-api/pkg/utils"
	"strconv"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// @summary SendResetPasswordOTP
// @description Sends a reset password OTP to the user via email
// @tags resetPassword
// @accept json
// @produce json
// @Param email query string true "User's email to send reset password otp"
// @Success 200 {object} response.GenerateEmailOtpResponse
// @failure 500 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @router /reset-password/send-otp [put]
func SendResetPasswordOTP(c echo.Context) error {
	email := c.QueryParam("email")

	logger := logging.GetEchoContextLogger(c)

	user, err := dao.MasterUserRecordDao{}.FindUserByEmail(email)

	if user == nil {
		logger.Warn("User record not found for email", "email", email)

		// Generate a fake OTP ID and expiration time
		fakeOtpId := generateFakeOtpId()
		expirationTime := config.Config.Otp.OtpExpiryDuration

		response := response.GenerateEmailOtpResponse{
			OtpId:             fakeOtpId,
			OtpExpiryDuration: expirationTime,
		}
		return c.JSON(http.StatusOK, response)
	}

	if err != nil {
		logging.Logger.Error("Error while fetching user record", "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("DB Error: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	// Send reset password OTP only for ACTIVE user
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

	// TODO: Need to implement retry otp logic.
	// Currently not considered, User can resend otp multiple times
	otp, err := utils.GenerateOTP(nil)
	// Only sending error code in the response
	if err != nil {
		return response.ErrorResponse{ErrorCode: constant.ERROR_IN_GENERATING_OTP, Message: constant.OTP_GENERATING_ERROR_MSG, StatusCode: http.StatusInternalServerError, MaybeInnerError: errtrace.Wrap(err)}
	}

	// Send otp on specified emailId
	err = sendOtpViaEmail(*user, email, otp)
	// Only sending error code in the response
	if err != nil {
		return response.ErrorResponse{ErrorCode: constant.ERROR_IN_SENDING_EMAIL, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error while sending otp via email: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	// Create or update otp record
	otpId, err := createOrUpdateOtpRecord(c.RealIP(), otp, user.Id, email)
	if err != nil {
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error in updating/creating record in user_otp table: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	response := response.GenerateEmailOtpResponse{
		OtpId:             otpId,
		OtpExpiryDuration: config.Config.Otp.OtpExpiryDuration,
	}

	logging.Logger.Info("OTP is generated and sent on email successfully", "customer", user.Id, "emailId", email)
	return c.JSON(http.StatusOK, response)
}

func generateFakeOtpId() string {
	return uuid.New().String()
}

func sendOtpViaEmail(user dao.MasterUserRecordDao, email, otp string) error {
	emailData := response.OtpEmailTemplateData{
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		OtpExpTimeMinutes: strconv.Itoa(config.Config.Otp.OtpExpiryDuration / 60000),
		Otp:               otp,
	}

	templateName := config.Config.Email.TemplateDirectory + constant.RESET_PASSWORD_EMAIL_TEMPLATE_NAME
	htmlBody, err := utils.GenerateEmailBody(templateName, emailData)
	if err != nil {
		return errtrace.Wrap(err)
	}

	emailSubject := "Password Reset"
	err = utils.SendEmail(user.FullName(), email, emailSubject, htmlBody)
	if err != nil {
		return errtrace.Wrap(err)
	}

	return nil
}

func createOrUpdateOtpRecord(ip, otp, userId, email string) (string, error) {
	otpId := uuid.New().String()
	apiPath := "/reset-password/send-otp"
	userOtp := dao.MasterUserOtpDao{
		OtpId:     otpId,
		Otp:       otp,
		OtpType:   constant.EMAIL,
		OtpStatus: constant.OTP_SENT,
		Email:     email,
		IP:        ip,
		UserId:    userId,
		UsedAt:    nil,
		ApiPath:   apiPath,
		CreatedAt: clock.Now(),
		UpdatedAt: clock.Now(),
	}

	// Update the record and count if the userId exists and the OTP is not used.(Currently count not considered)
	// Otherwise, create a new record.
	result := db.DB.Where("user_id=? AND used_at IS NULL AND api_path=?", userId, apiPath).Assign(userOtp).FirstOrCreate(&userOtp)
	if result.RowsAffected <= 0 {
		return "", errtrace.Wrap(fmt.Errorf("no record created or updated"))
	}

	return userOtp.OtpId, nil
}
