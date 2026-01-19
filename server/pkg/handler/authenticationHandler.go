package handler

import (
	"bytes"
	"context"
	"encoding/json"
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
	"process-api/pkg/sardine"
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"slices"
	"strings"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// @summary Login
// @description Logs in a user and returns a session token
// @tags auth
// @accept json
// @produce json
// @param login body request.LoginRequest true "Login payload"
// @success 200 {object} response.LoginResponse "Successful login"
// @header 200 {string} Authorization "Bearer token for user authentication"
// @header 200 {string} X-Token-Expiration "Token expiration timestamp"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /login [post]
func Login(c echo.Context) error {
	var credentials request.LoginRequest
	var user *dao.MasterUserRecordDao

	err := c.Bind(&credentials)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(credentials); err != nil {
		return err
	}

	user, err = dao.MasterUserRecordDao{}.FindUserByEmail(credentials.Username)
	if err != nil {
		// Handle Other DB error's
		logging.Logger.Error("Error while fetching user record", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error while fetching user record: %s", err.Error()), errtrace.Wrap(err))
	}

	if user == nil {
		logging.Logger.Error("No user record found with provided email", "username", credentials.Username)
		return response.GenerateErrResponse(constant.INVALID_CREDENTIALS_ERROR, constant.INVALID_CREDENTIALS_ERROR_MSG, "No user record found with provided email", http.StatusUnauthorized, errtrace.New(""))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		logging.Logger.Warn("Incorrect password entered for user", "username", credentials.Username)
		return response.GenerateErrResponse(constant.INVALID_CREDENTIALS_ERROR, constant.INVALID_CREDENTIALS_ERROR_MSG, "Incorrect password entered for user", http.StatusUnauthorized, errtrace.Wrap(err))
	}

	if slices.Contains(OnboardingUserStatuses, user.UserStatus) {
		now := clock.Now()
		token, err := security.GenerateRecoverOnboardingJwt(user.Id, &now)
		if err != nil {
			logging.Logger.Error("Error generating JWT", "error", err.Error())
			return response.InternalServerError("Error while generating jwt", errtrace.Wrap(err))
		}

		c.Response().Header().Set("Authorization", "Bearer "+token)

		mobileNo := ""
		if len(user.MobileNo) >= 4 {
			mobileNo = "XXX-XXX-" + user.MobileNo[len(user.MobileNo)-4:]
		}
		onboardingLoginResponse := response.LoginResponse{
			Id:               user.Id,
			Status:           user.UserStatus,
			IsUserOnboarding: true,
			MobileNo:         mobileNo,
			Email:            MaskEmail(user.Email),
		}

		if err := sendLoginEventToSardine(user.Id, credentials.SardineSessionKey); err != nil {
			logging.Logger.Warn("Failed to send login event to Sardine", "error", err)
		}
		return c.JSON(http.StatusOK, onboardingLoginResponse)
	}

	if user.UserStatus != constant.ACTIVE {
		return response.GenerateErrResponse("USER_NOT_ACTIVE", "User must have active ledger to login", "", http.StatusPreconditionFailed, errtrace.New(""))
	}
	now := clock.Now()

	sessionId := uuid.New().String()

	userLoginRecord := dao.MasterUserLoginDao{
		Id:     sessionId,
		UserId: user.Id,
		IP:     c.RealIP(),
	}

	err = db.DB.Select("id", "user_id", "ip", "created_at").Create(&userLoginRecord).Error
	if err != nil {
		logging.Logger.Error("User login record could not be created", "error", err)
		return response.InternalServerError(fmt.Sprintf("User login record could not be created: %s", err.Error()), errtrace.Wrap(err))
	}

	userPublicKey, err := dao.UserPublicKey{}.FindUserPublicRecord(user.Id, credentials.PublicKey)
	if err != nil {
		logging.Logger.Error("Failed to find UserPublicKey", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Failed to find UserPublicKey: %s", err.Error()), errtrace.Wrap(err))
	}

	if userPublicKey == nil {
		token, err := security.GenerateUnregisteredOnboardedJwt(user.Id, &now)
		if err != nil {
			logging.Logger.Error("Error generating JWT", "error", err.Error())
			return response.InternalServerError(fmt.Sprintf("Error generating JWT: %s", err.Error()), errtrace.Wrap(err))
		}
		unregisteredLoginResponse := response.LoginResponse{
			Id:               user.Id,
			CustomerNo:       user.LedgerCustomerNumber,
			Status:           user.UserStatus,
			IsUserRegistered: false,
			MobileNo:         "XXX-XXX-" + user.MobileNo[len(user.MobileNo)-4:],
			Email:            MaskEmail(user.Email),
		}

		c.Response().Header().Set("Authorization", "Bearer "+token)

		if err := sendLoginEventToSardine(user.Id, credentials.SardineSessionKey); err != nil {
			logging.Logger.Warn("Failed to send login event to Sardine", "error", err)
		}
		return c.JSON(http.StatusOK, unregisteredLoginResponse)
	}

	// NOTE: only the userPublicKey record can be trusted for assigning the JWT claims
	token, err := security.GenerateOnboardedJwt(user.Id, userPublicKey.PublicKey, &now)
	if err != nil {
		logging.Logger.Error("Error generating JWT", "error", err.Error())
		return response.InternalServerError("Could not generate token", errtrace.Wrap(err))
	}

	c.Response().Header().Set("Authorization", "Bearer "+token)

	registeredLoginResponse := response.LoginResponse{
		Id:               user.Id,
		CustomerNo:       user.LedgerCustomerNumber,
		Status:           user.UserStatus,
		IsUserRegistered: true,
	}

	if err := sendLoginEventToSardine(user.Id, credentials.SardineSessionKey); err != nil {
		logging.Logger.Warn("Failed to send login event to Sardine", "error", err)
	}
	return c.JSON(http.StatusOK, registeredLoginResponse)
}

func MaskEmail(email string) string {
	domainIndex := strings.Index(email, "@")
	if domainIndex == -1 || domainIndex < 1 {
		return email
	}

	visiblePrefix := 1
	if domainIndex > 2 {
		visiblePrefix = 2
	}

	maskedPart := strings.Repeat("*", domainIndex-visiblePrefix)
	domain := email[domainIndex:]

	return email[:visiblePrefix] + maskedPart + domain
}

// Many handlers that will soon be removed are using this function. New implementation of ResetToken in security
func DeprecatedResetToken(c echo.Context, customerNo string) (int, *response.ErrorResponse, *dao.UserDao) {
	return 0, nil, nil
}

func sendLoginEventToSardine(userId string, sardineSessionKey string) error {
	if sardineSessionKey == "" {
		return fmt.Errorf("missing sardine session key")
	}

	client, err := utils.NewSardineClient(config.Config.Sardine)
	if err != nil {
		return fmt.Errorf("failed to create sardine client: %w", err)
	}

	flowName := "login"
	flowType := sardine.FlowTypeLogin
	checkpoints := []sardine.PostCustomerInformationJSONBodyCheckpoints{
		"login",
	}

	requestBody := sardine.PostCustomerInformationJSONRequestBody{
		Flow: sardine.Flow{
			Name: &flowName,
			Type: &flowType,
		},
		SessionKey: sardineSessionKey,
		Customer: sardine.Customer{
			Id: userId,
		},
		Checkpoints: &checkpoints,
	}

	sardineResponse, err := client.PostCustomerInformationWithResponse(context.Background(), requestBody)
	if err != nil {
		return fmt.Errorf("error occurred while calling sardine API: %w", err)
	}

	switch {
	case sardineResponse.JSON200 != nil:
		var res bytes.Buffer
		if err := json.Indent(&res, sardineResponse.Body, "", "  "); err != nil {
			return fmt.Errorf("error occurred while formatting sardine response body: %w", err)
		}
		logging.Logger.Debug("Received 200 status code from Sardine", "successResponse", res.String())

	case sardineResponse.JSON400 != nil:
		return fmt.Errorf("received 400 response from sardine: %s", *sardineResponse.JSON400.Message)
	case sardineResponse.JSON401 != nil:
		return fmt.Errorf("received 401 response from sardine: %s", *sardineResponse.JSON401.Reason)
	case sardineResponse.JSON422 != nil:
		return fmt.Errorf("received 422 response from sardine: %s", *sardineResponse.JSON422.Message)
	default:
		return fmt.Errorf("received unexpected error response from sardine API")
	}

	return nil
}
