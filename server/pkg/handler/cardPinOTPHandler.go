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
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"strconv"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// @summary Get set card pin OTP
// @description Issues OTP for setting the user's primary card pin
// @tags cards
// @accept json
// @produce json
// @param createCardPinOTPRequest body CreateCardSetPinOTPRequest true "Create card set pin OTP payload"
// @success 200 {object} response.OtpResponseWithMaskedNumber "Successful otp issuance for logged in user"
// @Failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @router /account/cards/pin/otp [post]
func CreateSetCardPinOTP(c echo.Context) error {
	logger := logging.GetEchoContextLogger(c)

	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		logger.Error("CreateCardSetPinOTP failed to cast!")
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	requestData := new(CreateCardSetPinOTPRequest)
	var user dao.MasterUserRecordDao

	err := c.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	err = db.DB.Where("id = ?", userId).First(&user).Error
	if err != nil {
		logger.Error("User record not found")
		return c.NoContent(http.StatusNotFound)
	}

	otp, err := utils.GenerateOTP(nil)
	if err != nil {
		logger.Error(constant.ERROR_IN_GENERATING_OTP)
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
		UserId:    user.Id,
		OtpType:   requestData.Type,
		Otp:       otp,
		OtpStatus: constant.OTP_SENT,
		ApiPath:   "/account/cards/pin/otp",
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

// @summary ChallengeSetCardPinOtp
// @description Challenges the set card pin OTP
// @tags cards
// @accept json
// @produce json
// @param challengeSetCardPinOtpRequest body ChallengeSetCardPinOtpRequest true "payload"
// @Success 200 "OK"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/cards/pin/otp/verify [post]
func ChallengeSetCardPinOtp(c echo.Context) error {
	logger := logging.GetEchoContextLogger(c)

	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	var requestData ChallengeSetCardPinOtpRequest

	if err := c.Bind(&requestData); err != nil {
		logger.Error("Invalid request", "error", err.Error())
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	apiPath := "/account/cards/pin/otp"
	_, err := utils.VerifyOTP(requestData.OtpId, requestData.OtpValue, apiPath)
	if err != nil {
		logging.Logger.Error("error verifying otp for card pin", "error", err.Error())
		return response.GenerateOTPErrResponse(errtrace.Wrap(err))
	}

	logger.Info("Set Card Pin OTP verified successfully for user", "userId", userId)
	return c.NoContent(http.StatusOK)
}

type CreateCardSetPinOTPRequest struct {
	Type string `json:"type" validate:"required,oneof=SMS CALL"`
}

type ChallengeSetCardPinOtpRequest struct {
	OtpId    string `json:"otpId" validate:"required"`
	OtpValue string `json:"otpValue" validate:"required"`
}
