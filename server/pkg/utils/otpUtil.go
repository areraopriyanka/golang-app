package utils

import (
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/big"
	"net/http"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/model"
	"process-api/pkg/model/response"
	"strconv"
	"time"

	"braces.dev/errtrace"
)

func GenerateOTP(randReader *io.Reader) (string, error) {
	if config.Config.Otp.UseHardcodedOtp {
		return config.Config.Otp.HardcodedOtp, nil
	}

	digits := config.Config.Otp.OtpDigits

	if randReader == nil {
		randReader = &rand.Reader
	}

	otp, err := rand.Int(
		*randReader,
		big.NewInt(int64(math.Pow(10, float64(digits)))),
	)
	if err != nil {
		return "", errtrace.Wrap(err)
	}
	otpString := fmt.Sprintf("%0*d", digits, otp)
	return otpString, nil
}

func VerifyOTP(otpId, otp, apiPath string) (*dao.MasterUserOtpDao, error) {
	var userOtp dao.MasterUserOtpDao
	// Checking if the 'usedAt' value is null to ensure the OTP is used only once.
	result := db.DB.Where("otp_id=? AND otp=? AND used_at IS NULL AND api_path=?", otpId, otp, apiPath).Find(&userOtp)
	if result.Error != nil {
		return nil, errtrace.Wrap(fmt.Errorf("could not find unused OTP record for api path %s: %w", apiPath, result.Error))
	}

	isValid := otpIsNotExpired(userOtp.CreatedAt)
	if !isValid {
		result := db.DB.Model(&userOtp).Updates(dao.MasterUserOtpDao{OtpStatus: constant.OTP_EXPIRED})
		if result.Error != nil {
			return nil, errtrace.Wrap(fmt.Errorf("error in updating expired OTP status: %w", result.Error))
		}
		return nil, errtrace.Wrap(model.ErrOtpExpired)
	}
	now := clock.Now()
	result = db.DB.Model(&userOtp).Updates(dao.MasterUserOtpDao{OtpStatus: constant.OTP_VERIFIED, UsedAt: &now})
	if result.Error != nil {
		return nil, errtrace.Wrap(fmt.Errorf("error in updating verified OTP status: %w", result.Error))
	}
	return &userOtp, nil
}

func otpIsNotExpired(createdAt time.Time) bool {
	otpExpiryTime := createdAt.Add(time.Millisecond * time.Duration(config.Config.Otp.OtpExpiryDuration))
	// if otp expired return false else true
	return !clock.Now().After(otpExpiryTime)
}

func CheckOtpChallengeIsNotExpired(otpId, otp, apiPath, userId string) error {
	var userOtp dao.MasterUserOtpDao
	result := db.DB.Where("otp_id=? AND otp=? AND used_at IS NOT NULL AND api_path=? AND user_id=?", otpId, otp, apiPath, userId).Find(&userOtp)
	if result.Error != nil {
		return errtrace.Wrap(fmt.Errorf("could not find OTP record %s: %w", apiPath, result.Error))
	}

	if userOtp.ChallengeExpiredAt != nil {
		return errtrace.Wrap(model.ErrOtpExpired)
	}

	return nil
}

func ExpireOtpChallenge(otpId string) error {
	now := clock.Now()
	result := db.DB.Model(dao.MasterUserOtpDao{}).
		Where("otp_id=? AND used_at IS NOT NULL AND challenge_expired_at IS NULL", otpId).
		Update(dao.MasterUserOtpDao{ChallengeExpiredAt: &now})

	return errtrace.Wrap(result.Error)
}

func SendOTPWithType(otpType string, otp string, user dao.MasterUserRecordDao, logger *slog.Logger) error {
	switch otpType {
	case constant.EMAIL:
		if len(user.Email) == 0 {
			return &response.ErrorResponse{
				ErrorCode:       "EMAIL_MISSING",
				Message:         "Email is missing.",
				StatusCode:      http.StatusConflict,
				LogMessage:      "Attempted to send OTP email without email address",
				MaybeInnerError: errtrace.New(""),
			}
		}
		templateName := config.Config.Email.TemplateDirectory + constant.EMAIL_VERIFICATION_TEMPLATE_NAME
		emailData := response.OtpEmailTemplateData{
			FirstName:         user.FirstName,
			LastName:          user.LastName,
			OtpExpTimeMinutes: strconv.Itoa(config.Config.Otp.OtpExpiryDuration / 60000),
			Otp:               otp,
		}
		htmlBody, err := GenerateEmailBody(templateName, emailData)
		if err != nil {
			logger.Error("Error while generating device registration email body", "error", err.Error())
			return &response.ErrorResponse{
				ErrorCode:       constant.ERROR_GENERATING_EMAIL_BODY,
				Message:         constant.ERROR_GENERATING_EMAIL_BODY_MSG,
				StatusCode:      http.StatusInternalServerError,
				MaybeInnerError: errtrace.Wrap(err),
			}
		}
		err = SendEmail(user.FullName(), user.Email, "Your DreamFi One Time Password", htmlBody)
		if err != nil {
			logger.Error("Error while sending email", "error", err.Error())
			return &response.ErrorResponse{
				ErrorCode:       constant.ERROR_IN_SENDING_EMAIL,
				Message:         constant.ERROR_IN_SENDING_EMAIL_MSG,
				StatusCode:      http.StatusInternalServerError,
				MaybeInnerError: errtrace.Wrap(err),
			}
		}
	case constant.CALL:
		if len(user.MobileNo) == 0 {
			return &response.ErrorResponse{
				ErrorCode:       "MOBILE_NUMBER_MISSING",
				Message:         "Mobile number is missing.",
				StatusCode:      http.StatusConflict,
				LogMessage:      "Attempted to send OTP call without mobile number",
				MaybeInnerError: errtrace.New(""),
			}
		}
		twimlUrl := fmt.Sprintf("%svoice-xml?otp=%s", config.Config.Server.BaseUrl, otp)
		err := MakeCall(user.MobileNo, config.Config.Twilio.From, twimlUrl)
		if err != nil {
			return HandleTwilioError(errtrace.Wrap(err))
		}

	case constant.SMS:
		if len(user.MobileNo) == 0 {
			return &response.ErrorResponse{
				ErrorCode:       "MOBILE_NUMBER_MISSING",
				Message:         "Mobile number is missing.",
				StatusCode:      http.StatusConflict,
				LogMessage:      "Attempted to send OTP text without mobile number",
				MaybeInnerError: errtrace.New(""),
			}
		}
		expirationInMinutes := strconv.Itoa(config.Config.Otp.OtpExpiryDuration / 60000)
		body := fmt.Sprintf("Use this One Time Password %s to verify your phone number. This One Time Password will be valid for the next %s minute(s). Do not share this with anyone.", otp, expirationInMinutes)

		err := SendSMS(user.MobileNo, body, config.Config.Twilio.From)
		if err != nil {
			return HandleTwilioError(errtrace.Wrap(err))
		}

	default:
		logger.Error("user requested an OTP type that is not supported", "type", otpType)
		return &response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			Message:         constant.INTERNAL_SERVER_ERROR_MSG,
			StatusCode:      http.StatusInternalServerError,
			MaybeInnerError: errtrace.New(""),
		}
	}

	return nil
}
