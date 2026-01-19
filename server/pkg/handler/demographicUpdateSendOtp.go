package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @summary DemographicUpdateSendOtp
// @description Generates and sends an OTP code for the current user before initiating a demographic update
// @tags Demographic Updates
// @accept json
// @produce json
// @param payload body request.DemographicUpdateSendOtpRequest true "Payload indicating OTP type"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} response.DemographicUpdateSendOtpResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/customer/demographic-update/mobile [post]
func DemographicUpdateSendOtp(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	sendOtpRequest := new(request.DemographicUpdateSendOtpRequest)
	if err := c.Bind(sendOtpRequest); err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(sendOtpRequest); err != nil {
		return err
	}

	otp, err := utils.GenerateOTP(nil)
	if err != nil {
		logger.Error(constant.ERROR_IN_GENERATING_OTP)
		return response.GenerateErrResponse(constant.ERROR_IN_GENERATING_OTP, constant.OTP_GENERATING_ERROR_MSG, err.Error(), http.StatusInternalServerError, errtrace.Wrap(err))
	}

	err = utils.SendOTPWithType(sendOtpRequest.Type, otp, *user, logger)
	if err != nil {
		return err
	}

	apiPath := "/account/customer/demographic-update/mobile"
	otpId, err := updateOtpRecord(c.RealIP(), otp, user.MobileNo, user.Email, apiPath, userId, sendOtpRequest.Type)
	if err != nil {
		logger.Error("Error in updating/creating record in user_otp table", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error in updating/creating record in user_otp table: %s", err.Error()), errtrace.Wrap(err))
	}

	logger.Info("Demographic OTP sent successfully for user", "userId", userId)

	return c.JSON(http.StatusOK, response.DemographicUpdateSendOtpResponse{
		OtpId:             otpId,
		OtpExpiryDuration: config.Config.Otp.OtpExpiryDuration,
		MaskedMobileNo:    "XXX-XXX-" + user.MobileNo[len(user.MobileNo)-4:],
		MaskedEmail:       MaskEmail(user.Email),
	})
}
