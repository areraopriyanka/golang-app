package handler

import (
	"errors"
	"fmt"
	"net/http"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"time"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

// @summary Get Device Registration OTP
// @description Issues otp for unregistered device after login
// @tags auth
// @accept json
// @produce json
// @param getDeviceRegistrationOtpRequest body request.GetDeviceRegistrationOtpRequest true "Get Device Registration Otp payload"
// @success 200 {object} response.OtpResponse "Successful otp issuance for logged in user"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 409 {object} response.ErrorResponse
// @router /register-device/otp [post]
func GetDeviceRegistrationOtp(c echo.Context) error {
	cc, ok := c.(*security.LoggedInUnregisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	var requestData request.GetDeviceRegistrationOtpRequest
	var user dao.MasterUserRecordDao

	if err := c.Bind(&requestData); err != nil {
		logger.Error("Invalid request", "error", err.Error())
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
		logger.Error(constant.ERROR_IN_GENERATING_OTP)
		return response.GenerateErrResponse(constant.ERROR_IN_GENERATING_OTP, constant.OTP_GENERATING_ERROR_MSG, err.Error(), http.StatusInternalServerError, errtrace.Wrap(err))
	}

	err = utils.SendOTPWithType(requestData.Type, otp, user, logger)
	if err != nil {
		return err
	}

	sessionId := uuid.New().String()

	userOtpRecord := dao.MasterUserOtpDao{
		OtpId:     sessionId,
		UserId:    user.Id,
		OtpType:   requestData.Type,
		Otp:       otp,
		OtpStatus: constant.OTP_SENT,
		ApiPath:   "/register-device/otp",
		MobileNo:  user.MobileNo,
		Email:     user.Email,
		IP:        c.RealIP(),
		CreatedAt: clock.Now(),
	}

	otpResult := db.DB.Select("otp_id", "user_id", "otp_type", "otp", "otp_status", "api_path", "mobile_no", "email", "ip", "created_at").Create(&userOtpRecord)
	if otpResult.Error != nil {
		return response.InternalServerError("Failed to generate OTP", errtrace.Wrap(otpResult.Error))
	}

	return c.JSON(http.StatusOK, response.OtpResponse{
		OtpId:             userOtpRecord.OtpId,
		OtpExpiryDuration: config.Config.Otp.OtpExpiryDuration,
	})
}

// @summary Challenge Device Registration OTP
// @description Verifies the user-provided OTP and registers the device by generating a key if successful
// @tags auth
// @accept json
// @produce json
// @param challengeDeviceRegistrationOtpRequest body request.ChallengeDeviceRegistrationOtpRequest true "Challenge Device Registration Otp payload"
// @success 200 {object} response.ChallengeDeviceRegistrationOtpResponse "Successful otp verification for device registration"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /register-device/otp/verify [post]
func ChallengeDeviceRegistrationOtp(c echo.Context) error {
	var otpRequest request.ChallengeDeviceRegistrationOtpRequest
	var otpRecord dao.MasterUserOtpDao
	var userRecord dao.MasterUserRecordDao

	cc, ok := c.(*security.LoggedInUnregisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	if err := c.Bind(&otpRequest); err != nil {
		logger.Error("Invalid request", "error", err.Error())
		return response.BadRequestInvalidBody
	}

	userResult := db.DB.Where("id = ?", userId).First(&userRecord)
	if userResult.Error != nil {
		logger.Error("User record not found", "userId", userId, "error", userResult.Error.Error())
		return response.ErrorResponse{ErrorCode: constant.USER_NOT_FOUND, StatusCode: http.StatusUnauthorized, LogMessage: constant.EMAIL_ALREADY_REGISTERED_MSG, MaybeInnerError: errtrace.Wrap(userResult.Error)}
	}

	if userResult.RowsAffected == 0 {
		logger.Error("User not found")
		return response.ErrorResponse{ErrorCode: constant.INVALID_CREDENTIALS_ERROR, StatusCode: http.StatusUnauthorized, LogMessage: "User not found", MaybeInnerError: errtrace.New("")}
	}

	otpResult := db.DB.Where("otp_id = ?", otpRequest.OtpId).First(&otpRecord)
	if otpResult.Error != nil {
		if errors.Is(otpResult.Error, gorm.ErrRecordNotFound) {
			logger.Error("OTP record not found")
			return response.ErrorResponse{ErrorCode: constant.INVALID_OTP, Message: "OTP record not found", StatusCode: http.StatusNotFound, MaybeInnerError: errtrace.Wrap(otpResult.Error)}
		}
		logger.Error("DB Error", "error", otpResult.Error.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			Message:         constant.INTERNAL_SERVER_ERROR_MSG,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("DB Error: %s", otpResult.Error.Error()),
			MaybeInnerError: errtrace.Wrap(otpResult.Error),
		}
	}

	expiryTime := otpRecord.CreatedAt.Add(time.Duration(config.Config.Otp.OtpExpiryDuration * int(time.Millisecond)))
	logger.Info("OTP Expiry Check", "createdAt", otpRecord.CreatedAt, "expiryTime", expiryTime, "currentTime", clock.Now())
	if clock.Now().After(expiryTime) {
		logger.Error("OTP has expired for user", "email", userRecord.Email)

		err := db.DB.Model(&otpRecord).Update("otp_status", constant.OTP_EXPIRED).Error
		if err != nil {
			logger.Error("OTP record not found")
			return c.NoContent(http.StatusNotFound)
		}
		return response.ErrorResponse{ErrorCode: constant.OTP_EXPIRED, StatusCode: http.StatusBadRequest, Message: constant.OTP_EXPIRED_ERROR_MSG, MaybeInnerError: errtrace.New("")}
	}

	if otpRecord.OtpStatus == constant.OTP_EXPIRED || otpRecord.OtpStatus == constant.OTP_VERIFIED {
		logger.Error("OTP record not valid")
		return response.ErrorResponse{ErrorCode: constant.INVALID_OTP, StatusCode: http.StatusBadRequest, Message: constant.INVALID_OTP_MSG, MaybeInnerError: errtrace.New("")}
	}

	if otpRecord.Otp != otpRequest.OtpValue {
		logger.Error("Incorrect OTP entered for user", "email", userRecord.Email)
		return response.ErrorResponse{ErrorCode: constant.INVALID_OTP, StatusCode: http.StatusBadRequest, LogMessage: constant.INVALID_OTP_MSG, MaybeInnerError: errtrace.New("")}
	}

	if otpRecord.ApiPath != "/register-device/otp" {
		logger.Error("Incorrect API Path of otp record for user", "email", userRecord.Email)
		return response.ErrorResponse{ErrorCode: constant.INVALID_OTP, StatusCode: http.StatusBadRequest, LogMessage: "Incorrect API path for OTP", MaybeInnerError: errtrace.New("")}
	}

	result := db.DB.Model(&otpRecord).Update("otp_status", constant.OTP_VERIFIED)
	if result.Error != nil {
		logger.Error("Error updating OTP status", "error", result.Error.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			Message:         "Error updating OTP status",
			StatusCode:      http.StatusInternalServerError,
			MaybeInnerError: errtrace.Wrap(result.Error),
		}
	}

	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDLedgerApiClient(config.Config.Ledger, ledgerParamsBuilder)

	request := ledger.BuildAddUserKeyRequest(userRecord.Email, otpRequest.PublicKey)
	resp, err := ledgerClient.AddUserKey(request)
	if err != nil {
		logger.Error("Error adding user's public key to ledger", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error adding user's public key to ledger: error: %s", err.Error()), errtrace.Wrap(err))
	}

	if resp.Error != nil {
		if resp.Error.Code == "USER_ALREADY_REGISTERED_WITH_GIVEN_PUBLICKEY" {
			logger.Warn("User resubmitted existing publicKey. Asking client to regenerate", "code", resp.Error.Code, "msg", resp.Error.Message)
			return response.ErrorResponse{
				ErrorCode:       "USER_ALREADY_REGISTERED_WITH_GIVEN_PUBLICKEY",
				Message:         "User already registed with this publicKey. Generate a new public/private pair.",
				StatusCode:      http.StatusConflict,
				MaybeInnerError: errtrace.New(""),
			}
		}
		logger.Error("error message from ledger AddUserKey", "code", resp.Error.Code, "msg", resp.Error.Message)
		return response.InternalServerError(fmt.Sprintf("error message from ledger AddUserKey: error: %s", resp.Error.Message), errtrace.Wrap(err))
	}

	encryptedApiKey, err := utils.EncryptKmsBinary(resp.Result.ApiKey)
	if err != nil {
		logger.Error("error encrypting ledger api key", "apiKey", resp.Result.ApiKey)
		return response.InternalServerError("error encrypting ledger api key", errtrace.Wrap(err))
	}

	userPublicKey := dao.UserPublicKey{
		UserId:             userId,
		KmsEncryptedApiKey: []byte(encryptedApiKey),
		KeyId:              resp.Result.KeyID,
		PublicKey:          otpRequest.PublicKey,
	}

	userPublicKeyResult := db.DB.Select("user_id", "kms_encrypted_api_key", "key_id", "public_key").Create(&userPublicKey)
	if userPublicKeyResult.Error != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			Message:         "Failed to insert public key record",
			StatusCode:      http.StatusInternalServerError,
			MaybeInnerError: errtrace.Wrap(userPublicKeyResult.Error),
		})
	}

	now := clock.Now()
	token, err := security.GenerateOnboardedJwt(userRecord.Id, otpRequest.PublicKey, &now)
	if err != nil {
		logger.Error("Error generating JWT", "error", err.Error())
		return response.InternalServerError("Could not generate token", errtrace.Wrap(err))
	}

	err = db.DB.Model(&otpRecord).Update("used_at", clock.Now()).Error
	if err != nil {
		logger.Error("OTP record not found")
		return c.NoContent(http.StatusNotFound)
	}
	c.Set("SkipTokenReset", true)
	c.Response().Header().Set("Authorization", "Bearer "+token)

	return c.JSON(http.StatusOK, response.ChallengeDeviceRegistrationOtpResponse{KeyId: resp.Result.KeyID})
}
