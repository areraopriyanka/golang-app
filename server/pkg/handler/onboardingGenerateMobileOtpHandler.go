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
	"process-api/pkg/validators"
	"strconv"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// @summary SendMobileVerificationOtp
// @description Generates and sends mobile verification OTP code for a current user session to a specified phone number
// @tags onboarding
// @accept json
// @produce json
// @param sendMobileVerificationOtpRequest body request.SendMobileVerificationOtpRequest true "SendMobileVerificationOtp payload"
// @Param userId path string true "User ID"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 201 {object} response.OtpResponse
// @header 201 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /onboarding/customer/{userId}/mobile [put]
// @router /onboarding/customer/mobile [put]
func SendMobileVerificationOtp(c echo.Context) error {
	// TODO: Remove path param logic later and use userId from JWT only.
	userId := c.Param("userId")
	if userId == "" {
		cc, ok := c.(*security.OnboardingUserContext)
		if !ok {
			return response.ErrorResponse{ErrorCode: constant.UNAUTHORIZED_ACCESS_ERROR, Message: constant.UNAUTHORIZED_ACCESS_ERROR_MSG, StatusCode: http.StatusUnauthorized, LogMessage: "Invalid type of custom context", MaybeInnerError: errtrace.New("")}
		}

		userId = cc.UserId
	}

	logger := logging.GetEchoContextLogger(c)

	requestData := new(request.SendMobileVerificationOtpRequest)

	err := c.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	user, errResponse := dao.RequireUserWithState(
		userId, constant.AGE_VERIFICATION_PASSED, constant.PHONE_VERIFICATION_OTP_SENT,
	)
	if errResponse != nil {
		return errResponse
	}

	// Validate and convert mobile number to E.164 formate before sending sms
	mobileNo, err := validators.VerifyAndFormatMobileNumber(requestData.MobileNo)
	if err != nil {
		logger.Error("Error while validating and converting mobile number to E.164 formate", "error", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	otp, err := utils.GenerateOTP(nil)
	if err != nil {
		return response.ErrorResponse{ErrorCode: constant.ERROR_IN_GENERATING_OTP, Message: constant.OTP_GENERATING_ERROR_MSG, StatusCode: http.StatusInternalServerError, MaybeInnerError: errtrace.Wrap(err)}
	}

	switch requestData.Type {
	case constant.SMS:
		body := fmt.Sprintf("Use this One Time Password %s to verify your phone number. This One Time Password will be valid for the next %s minute(s). Do not share this with anyone.", otp, strconv.Itoa(config.Config.Otp.OtpExpiryDuration/60000))

		err = utils.SendSMS(mobileNo, body, config.Config.Twilio.From)
		if err != nil {
			return utils.HandleTwilioError(err)
		}
	case constant.CALL:
		twimlUrl := fmt.Sprintf("%svoice-xml?otp=%s", config.Config.Server.BaseUrl, otp)
		err = utils.MakeCall(mobileNo, config.Config.Twilio.From, twimlUrl)
		if err != nil {
			return utils.HandleTwilioError(err)
		}
	default:
		logger.Error("user requested an OTP type that is not supported", "type", requestData.Type)
		return c.NoContent(http.StatusBadRequest)
	}

	apiPath := "/onboarding/customer/mobile"
	otpId, err := updateOtpRecord(c.RealIP(), otp, mobileNo, user.Email, apiPath, userId, requestData.Type)
	if err != nil {
		logger.Error("Error in updating/creating record in user_otp table", "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Error in updating/creating record in user_otp table: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	if user.UserStatus == constant.AGE_VERIFICATION_PASSED {
		err = updateUserStatus(userId, constant.PHONE_VERIFICATION_OTP_SENT)
		if err != nil {
			return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, Message: constant.INTERNAL_SERVER_ERROR_MSG, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error in updating user's status: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
		}
	}

	response := response.OtpResponse{
		OtpExpiryDuration: config.Config.Otp.OtpExpiryDuration,
		OtpId:             otpId,
	}

	logger.Info("OTP generated and sent SMS successfully for user", "userId", userId)
	return c.JSON(http.StatusCreated, response)
}

func updateOtpRecord(ip, otp, mobileNo, email, apiPath, userId, otpType string) (string, error) {
	// TODO: Need to implement retry otp logic. Currently not considered

	otpId := uuid.New().String()
	userOtp := dao.MasterUserOtpDao{
		OtpId:     otpId,
		Otp:       otp,
		OtpStatus: constant.OTP_SENT,
		OtpType:   otpType,
		IP:        ip,
		MobileNo:  mobileNo,
		Email:     email,
		UserId:    userId,
		UsedAt:    nil,
		ApiPath:   apiPath,
		CreatedAt: clock.Now(),
		UpdatedAt: clock.Now(),
	}

	// Update the record if the userId exists and the OTP is not used.
	// Otherwise, create a new record.
	result := db.DB.Where("user_id=? AND used_at IS NULL AND api_path=?", userId, apiPath).Assign(userOtp).FirstOrCreate(&userOtp)
	if result.RowsAffected <= 0 {
		return "", errtrace.Wrap(result.Error)
	}
	return otpId, nil
}

func updateUserStatus(userId string, userStatus string) error {
	data := dao.MasterUserRecordDao{
		UserStatus: userStatus,
	}

	result := db.DB.Model(dao.MasterUserRecordDao{}).Where("id=?", userId).Update(data)
	if result.Error != nil {
		return errtrace.Wrap(result.Error)
	}

	if result.RowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("user's status not updated"))
	}

	return nil
}
