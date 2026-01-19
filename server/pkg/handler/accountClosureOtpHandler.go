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

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// @summary SendAccountClosureOTP
// @description Sends an account closure OTP to the user via sms email or call
// @tags accountClosure
// @accept json
// @produce json
// @param createAccountClosureOtpRequest body CreateAccountClosureOtpRequest true "Payload for account closure OTP"
// @param Authorization header string true "Bearer token for user authentication"
// @success 200 {object} CreateAccountClosureOtpResponse "Successful otp issuance for logged in user"
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 409 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @router /account/close/send-otp [post]
func SendAccountClosureOTP(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	var requestData CreateAccountClosureOtpRequest

	err := c.Bind(&requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	otp, err := utils.GenerateOTP(nil)
	if err != nil {
		logger.Error("Error generating OTP", "error", err.Error())
		return response.GenerateErrResponse(constant.ERROR_IN_GENERATING_OTP, constant.OTP_GENERATING_ERROR_MSG, err.Error(), http.StatusInternalServerError, errtrace.Wrap(err))
	}

	err = utils.SendOTPWithType(requestData.Type, otp, *user, logger)
	if err != nil {
		return errtrace.Wrap(err)
	}

	userOtpRecord := dao.MasterUserOtpDao{
		OtpId:     uuid.New().String(),
		UserId:    userId,
		OtpType:   requestData.Type,
		Otp:       otp,
		OtpStatus: constant.OTP_SENT,
		ApiPath:   "/account/close/send-otp",
		MobileNo:  user.MobileNo,
		Email:     user.Email,
		IP:        c.RealIP(),
		CreatedAt: clock.Now(),
	}

	otpResult := db.DB.Select("otp_id", "user_id", "otp_type", "otp", "otp_status", "api_path", "mobile_no", "email", "ip", "created_at").Create(&userOtpRecord)
	if otpResult.Error != nil {
		return response.InternalServerError(fmt.Sprintf("Failed to generate OTP: %s", otpResult.Error), errtrace.Wrap(otpResult.Error))
	}

	return c.JSON(http.StatusOK, CreateAccountClosureOtpResponse{
		OtpId:             userOtpRecord.OtpId,
		OtpExpiryDuration: config.Config.Otp.OtpExpiryDuration,
		MaskedMobileNo:    "XXX-XXX-" + user.MobileNo[len(user.MobileNo)-4:],
		MaskedEmail:       MaskEmail(user.Email),
	})
}

// @summary ChallengeAccountClosureOTP
// @description Challenges the account closure OTP
// @tags accountClosure
// @accept json
// @produce json
// @param ChallengeAccountClosureOtpRequest body request.ChallengeOtpRequest true "payload"
// @param Authorization header string true "Bearer token for user authentication"
// @Success 200 "OK"
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/close/verify-otp [post]
func ChallengeAccountClosureOTP(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	var requestData request.ChallengeOtpRequest

	err := c.Bind(&requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	apiPath := "/account/close/send-otp"
	_, err = utils.VerifyOTP(requestData.OtpId, requestData.Otp, apiPath)
	if err != nil {
		logging.Logger.Error("error verifying otp for account closure", "error", err.Error())
		return response.GenerateOTPErrResponse(errtrace.Wrap(err))
	}

	logger.Info("Account closure OTP verified successfully for user", "userId", userId)
	return c.NoContent(http.StatusOK)
}

type CreateAccountClosureOtpRequest struct {
	Type string `json:"type" validate:"required,otpType"`
}

type CreateAccountClosureOtpResponse struct {
	OtpExpiryDuration int    `json:"otpExpiryDuration" validate:"required"`
	OtpId             string `json:"otpId" validate:"required" mask:"true"`
	MaskedMobileNo    string `json:"maskedMobileNo" validate:"required" mask:"true"`
	MaskedEmail       string `json:"maskedEmail" validate:"required" mask:"true"`
}
