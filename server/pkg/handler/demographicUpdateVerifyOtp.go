package handler

import (
	"net/http"
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

// @summary DemographicUpdateVerifyOtp
// @description Verifies demographic update otp
// @tags Demographic Updates
// @accept json
// @param payload body request.ChallengeOtpRequest true "Payload with otpId and otp"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 "OK"
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/customer/demographic-update/mobile/verify [post]
func DemographicUpdateVerifyOtp(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	_, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	verifyOtpRequest := new(request.ChallengeOtpRequest)
	if err := c.Bind(verifyOtpRequest); err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(verifyOtpRequest); err != nil {
		return err
	}

	apiPath := "/account/customer/demographic-update/mobile"
	_, err := utils.VerifyOTP(verifyOtpRequest.OtpId, verifyOtpRequest.Otp, apiPath)
	if err != nil {
		return response.GenerateOTPErrResponse(errtrace.Wrap(err))
	}

	logger.Info("Demographic OTP verified successfully for user", "userId", userId)
	return c.NoContent(http.StatusOK)
}
