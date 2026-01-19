package handler

import (
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

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var OnboardingUserStatuses = []string{
	constant.USER_CREATED,
	constant.AGREEMENTS_REVIEWED,
	constant.AGE_VERIFICATION_PASSED,
	constant.PHONE_VERIFICATION_OTP_SENT,
	constant.PHONE_NUMBER_VERIFIED,
	constant.ADDRESS_CONFIRMED,
	constant.KYC_PASS,
	constant.KYC_FAIL,
	constant.MEMBERSHIP_ACCEPTED,
	constant.CARD_AGREEMENTS_REVIEWED,
}

// @summary RecoverOnboardingOTP
// @description Sends an onboarding recovery OTP to the user
// @accept json
// @produce json
// @param recoverOnboardingOtpRequest body RecoverOnboardingOtpRequest true "RecoverOnboardingOtpRequest"
// @param Authorization header string true "Bearer token for user authentication"
// @success 200 {object} RecoverOnboardingOtpResponse "RecoverOnboardingOtpResponse"
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 409 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @router /recover-onboarding/send-otp [post]
func RecoverOnboardingOTP(c echo.Context) error {
	cc, ok := c.(*security.RecoverOnboardingUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	var requestData RecoverOnboardingOtpRequest

	err := c.Bind(&requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	user, errResponse := dao.RequireUserWithState(userId, OnboardingUserStatuses...)
	if errResponse != nil {
		return errResponse
	}

	otp, err := utils.GenerateOTP(nil)
	if err != nil {
		logger.Error("Error generating OTP", "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.ERROR_IN_GENERATING_OTP, Message: constant.OTP_GENERATING_ERROR_MSG, StatusCode: http.StatusInternalServerError, MaybeInnerError: errtrace.Wrap(err)}
	}

	err = utils.SendOTPWithType(requestData.Type, otp, *user, logger)
	if err != nil {
		return err
	}

	userOtpRecord := dao.MasterUserOtpDao{
		OtpId:     uuid.New().String(),
		UserId:    userId,
		OtpType:   requestData.Type,
		Otp:       otp,
		OtpStatus: constant.OTP_SENT,
		ApiPath:   "/recover-onboarding/send-otp",
		MobileNo:  user.MobileNo,
		Email:     user.Email,
		IP:        c.RealIP(),
		CreatedAt: clock.Now(),
	}

	otpResult := db.DB.Select("otp_id", "user_id", "otp_type", "otp", "otp_status", "api_path", "mobile_no", "email", "ip", "created_at").Create(&userOtpRecord)
	if otpResult.Error != nil {
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			Message:         "Failed to generate OTP",
			StatusCode:      http.StatusInternalServerError,
			MaybeInnerError: errtrace.Wrap(otpResult.Error),
		}
	}

	maskedMobileNo := ""
	if len(user.MobileNo) >= 4 {
		maskedMobileNo = "XXX-XXX-" + user.MobileNo[len(user.MobileNo)-4:]
	}

	return c.JSON(http.StatusOK, RecoverOnboardingOtpResponse{
		OtpId:             userOtpRecord.OtpId,
		OtpExpiryDuration: config.Config.Otp.OtpExpiryDuration,
		MaskedMobileNo:    maskedMobileNo,
		MaskedEmail:       MaskEmail(user.Email),
	})
}

// @summary ChallengeRecoverOnboardingOTP
// @description Challenges the onboarding recovery OTP
// @accept json
// @produce json
// @param challengeRecoverOnboardingOtpRequest body ChallengeRecoverOnboardingOtpRequest true "ChallengeRecoverOnboardingOtpRequest"
// @success 200 {object} ChallengeRecoverOnboardingOtpResponse "ChallengeRecoverOnboardingOtpResponse"
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /recover-onboarding/verify-otp [post]
func ChallengeRecoverOnboardingOTP(c echo.Context) error {
	cc, ok := c.(*security.RecoverOnboardingUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	var requestData ChallengeRecoverOnboardingOtpRequest

	err := c.Bind(&requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	user, errResponse := dao.RequireUserWithState(userId, OnboardingUserStatuses...)
	if errResponse != nil {
		return errResponse
	}

	apiPath := "/recover-onboarding/send-otp"
	_, err = utils.VerifyOTP(requestData.OtpId, requestData.Otp, apiPath)
	if err != nil {
		logging.Logger.Error("error verifying otp for onboarding recovery", "error", err.Error())
		return response.GenerateOTPErrResponse(errtrace.Wrap(err))
	}

	logger.Info("Onboarding recovery OTP verified successfully for user", "userId", userId)

	now := clock.Now()
	token, err := security.GenerateOnboardingJwt(userId, &now)
	if err != nil {
		logger.Error("Error generating JWT", "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, Message: "Error while generating jwt", StatusCode: http.StatusInternalServerError, MaybeInnerError: errtrace.Wrap(err)}
	}
	c.Set("SkipTokenReset", true)
	c.Response().Header().Set("Authorization", "Bearer "+token)

	var userData *ChallengeRecoverOnboardingOtpResponseUserData = nil
	if len(user.FirstName) != 0 && len(user.LastName) != 0 {
		userData = &ChallengeRecoverOnboardingOtpResponseUserData{
			DOB:       user.DOB.Format("01/02/2006"),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Suffix:    user.Suffix,
		}
	}

	return c.JSON(http.StatusOK, ChallengeRecoverOnboardingOtpResponse{
		UserStatus: user.UserStatus,
		UserData:   userData,
	})
}

type RecoverOnboardingOtpRequest struct {
	Type string `json:"type" validate:"required,otpType"`
}

type RecoverOnboardingOtpResponse struct {
	OtpExpiryDuration int    `json:"otpExpiryDuration" validate:"required"`
	OtpId             string `json:"otpId" validate:"required" mask:"true"`
	MaskedMobileNo    string `json:"maskedMobileNo" validate:"required" mask:"true"`
	MaskedEmail       string `json:"maskedEmail" validate:"required" mask:"true"`
}

type ChallengeRecoverOnboardingOtpRequest struct {
	OtpId string `json:"otpId" validate:"required"`
	Otp   string `json:"otp" validate:"required"`
}

type ChallengeRecoverOnboardingOtpResponseUserData struct {
	DOB       string `json:"dob" validate:"required" mask:"true"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Suffix    string `json:"suffix" validate:"required"`
}

type ChallengeRecoverOnboardingOtpResponse struct {
	UserStatus string                                         `json:"userStatus" validate:"required" enums:"USER_CREATED,AGREEMENTS_REVIEWED,AGE_VERIFICATION_PASSED,PHONE_VERIFICATION_OTP_SENT,PHONE_NUMBER_VERIFIED,ADDRESS_CONFIRMED,KYC_PASS,KYC_FAIL,MEMBERSHIP_ACCEPTED,CARD_AGREEMENTS_REVIEWED"`
	UserData   *ChallengeRecoverOnboardingOtpResponseUserData `json:"userData"`
}
