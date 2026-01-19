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
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

// @summary VerifyMobileVerificationOtp
// @description Verifies mobile verification OTP code for a current user session
// @tags onboarding
// @accept json
// @Param Authorization header string true "Bearer token for user authentication"
// @param verifyMobileVerificationOtpRequest body request.ChallengeOtpRequest true "VerifyMobileVerificationOtpRequest payload"
// @Param userId path string true "User ID"
// @success 200 "OK"
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /onboarding/customer/{userId}/mobile [post]
// @router /onboarding/customer/mobile [post]
func VerifyMobileVerificationOtp(c echo.Context) error {
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

	request := new(request.ChallengeOtpRequest)

	err := c.Bind(request)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(request); err != nil {
		return err
	}

	_, errResponse := dao.RequireUserWithState(userId, constant.PHONE_VERIFICATION_OTP_SENT)
	if errResponse != nil {
		return errResponse
	}

	apiPath := "/onboarding/customer/mobile"
	userOtp, err := utils.VerifyOTP(request.OtpId, request.Otp, apiPath)
	if err != nil {
		logging.Logger.Error("error verifying otp for mobile verification", "error", err.Error())
		return response.GenerateOTPErrResponse(errtrace.Wrap(err))
	}

	err = updateUser(userId, userOtp.MobileNo)
	var pgErr *pq.Error
	if err != nil {
		// The top-level error is now of type errtrace.Error, not *pq.Error
		// Therefore, we use errors.As instead of a direct type assertion.
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return response.BadRequestErrors{
					Errors: []response.BadRequestError{
						{
							FieldName: "mobile_number",
							Error:     constant.MOBILE_NUMBER_ALREADY_EXISTS_MSG,
						},
					},
				}
			}
		}
		logger.Error("Other DB Error", "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Other DB Error: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	logger.Info("OTP verified successfully for user", "userId", userId)
	return c.NoContent(http.StatusOK)
}

func updateUser(userId string, mobileNo string) error {
	data := dao.MasterUserRecordDao{
		UserStatus: constant.PHONE_NUMBER_VERIFIED,
		MobileNo:   mobileNo,
	}

	result := db.DB.Model(dao.MasterUserRecordDao{}).Where("id=?", userId).Updates(data)
	if result.Error != nil {
		return errtrace.Wrap(result.Error)
	}
	return nil
}
