package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/clock"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// @summary CreateUserAccount
// @description Creates a user in the middleware DB and sends a JWT containing the user ID in the response header.
// @tags onboarding
// @accept json
// @produce json
// @param CreateUserAccountRequest body request.CreateUserAccountRequest true "CreateUserAccount payload"
// @Success 200 "OK"
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @failure 409 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /onboarding/customer/account [post]
func CreateUserAccount(c echo.Context) error {
	request := new(request.CreateUserAccountRequest)

	logger := logging.GetEchoContextLogger(c)

	err := c.Bind(request)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(request); err != nil {
		return err
	}

	// Ensure the email is not already registered, either in middleware or in the ledger
	isEmailExists, err := IsUserEmailRegistered(c, request.Email)
	if err != nil {
		logger.Error("isUserEmailRegistered failed", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("isUserEmailRegistered failed: %s", err.Error()), errtrace.Wrap(err))
	}

	if isEmailExists {
		logger.Error("Error creating user account", "error", "Email address is already registered")
		return response.GenerateErrResponse(constant.EMAIL_ALREADY_REGISTERED, constant.EMAIL_ALREADY_REGISTERED_MSG, "", http.StatusConflict, errtrace.New(""))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.InternalServerError(fmt.Sprintf("Error while generating hash of the password: %s", err.Error()), errtrace.Wrap(err))
	}

	// Generate long lived non-guessable id
	sessionId := uuid.New().String()

	userRecord := dao.MasterUserRecordDao{
		Id:         sessionId,
		Email:      request.Email,
		Password:   hashedPassword,
		UserStatus: constant.USER_CREATED,
	}

	result := db.DB.Select("id", "email", "user_status", "password", "created_at").Create(&userRecord)
	if result.Error != nil {
		return response.InternalServerError(fmt.Sprintf("DB error: %s", result.Error.Error()), errtrace.Wrap(err))
	}

	now := clock.Now()
	token, err := security.GenerateOnboardingJwt(sessionId, &now)
	if err != nil {
		logging.Logger.Error("Error generating JWT", "error", err.Error())
		return response.InternalServerError("Error while generating jwt", errtrace.Wrap(err))
	}

	c.Response().Header().Set("Authorization", "Bearer "+token)

	return c.NoContent(http.StatusOK)
}
