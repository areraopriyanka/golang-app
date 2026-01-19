package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"time"

	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

func (suite *IntegrationTestSuite) TestLoginWithKeyId() {
	unfreeze := clock.FreezeNow()
	defer unfreeze()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	suite.Require().NoError(err, "Failed to encrypt password")

	sessionId := uuid.New().String()

	userRecord := dao.MasterUserRecordDao{
		Id:                   sessionId,
		FirstName:            "Foo",
		LastName:             "Bar",
		Email:                "user@email.com",
		Password:             hashedPassword,
		LedgerCustomerNumber: "1234567890",
		UserStatus:           constant.ACTIVE,
	}

	config.Config.Jwt.SecreteKey = "example_key"

	err = suite.TestDB.Select("id", "first_name", "last_name", "email", "password", "ledger_customer_number", "user_status").Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	userPublicKey := dao.UserPublicKey{
		UserId:    sessionId,
		KeyId:     "exampleKeyId",
		PublicKey: "examplePublicKey",
	}
	err = suite.TestDB.Select("user_id", "key_id", "public_key").Create(&userPublicKey).Error
	suite.Require().NoError(err, "Failed to insert user public key record")

	updateRequest := request.LoginRequest{
		Username:  "user@email.com",
		Password:  "Password123!",
		PublicKey: "examplePublicKey",
	}
	requestBody, err := json.Marshal(updateRequest)
	suite.Require().NoError(err, "Failed to marshall request")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/login")

	err = handler.Login(c)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var user dao.MasterUserRecordDao

	err = suite.TestDB.Model(&dao.MasterUserRecordDao{}).Where("id=?", sessionId).Find(&user).Error
	suite.Require().NoError(err, "Failed to fetch user data")

	var loginResponse response.LoginResponse
	err = json.Unmarshal(rec.Body.Bytes(), &loginResponse)
	suite.Require().NoError(err, "Failed to unmarshal login response")

	suite.Equal(user.LedgerCustomerNumber, loginResponse.CustomerNo, "Customer number should match")
	suite.Equal(user.UserStatus, loginResponse.Status, "User status should be ACTIVE")
	suite.True(loginResponse.IsUserRegistered, "User is considered registered")
	suite.False(loginResponse.IsUserOnboarding, "User is not considered onboarding")

	tokenWithBearer := rec.Header().Get("Authorization")
	suite.NotEmpty(tokenWithBearer, "Authorization token should be present in response headers")

	claims := security.GetClaimsFromToken(tokenWithBearer)

	suite.Equal(claims.Subject, sessionId, "JWT claim subject should match userId from user record")
	suite.Equal("onboarded", claims.Type, "JWT claim type should be 'onboarded'")
	suite.WithinDuration(clock.Now().Add(30*time.Minute), claims.ExpiresAt.Time, 0, "Expiration should match expectation")

	var loginRecord dao.MasterUserLoginDao
	err = suite.TestDB.Model(&dao.MasterUserLoginDao{}).Where("user_id = ?", sessionId).First(&loginRecord).Error
	suite.Require().NoError(err, "Failed to fetch user data")
	suite.Equal(sessionId, loginRecord.UserId, "Login Record UserId must match user record")
	suite.NotEmpty(loginRecord.IP, "Login record must contain IP address from request")
	suite.Require().Empty(loginResponse.Email, "User email should not be sent if keyId is present")
	suite.Require().Empty(loginResponse.MobileNo, "User mobile no should not be sent if keyId is present")
}

func (suite *IntegrationTestSuite) TestLoginWithOutKeyId() {
	unfreeze := clock.FreezeNow()
	defer unfreeze()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	suite.Require().NoError(err, "Failed to encrypt password")

	sessionId := uuid.New().String()

	userRecord := dao.MasterUserRecordDao{
		Id:                   sessionId,
		FirstName:            "Foo",
		LastName:             "Bar",
		Email:                "user@email.com",
		Password:             hashedPassword,
		LedgerCustomerNumber: "1234567890",
		UserStatus:           constant.ACTIVE,
		MobileNo:             "1234567890",
	}

	config.Config.Jwt.SecreteKey = "example_key"

	err = suite.TestDB.Select("id", "first_name", "last_name", "email", "password", "ledger_customer_number", "user_status", "mobile_no").Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	userPublicKey := dao.UserPublicKey{
		UserId:    sessionId,
		KeyId:     "",
		PublicKey: "examplePublicKey",
	}
	err = suite.TestDB.Select("user_id", "key_id", "public_key").Create(&userPublicKey).Error
	suite.Require().NoError(err, "Failed to insert user public key record")

	updateRequest := request.LoginRequest{
		Username: "user@email.com",
		Password: "Password123!",
	}
	requestBody, err := json.Marshal(updateRequest)
	suite.Require().NoError(err, "Failed to marshall request")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/login")

	err = handler.Login(c)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	tokenWithBearer := rec.Header().Get("Authorization")
	suite.NotEmpty(tokenWithBearer, "Authorization token should be present in response headers")

	claims := security.GetClaimsFromToken(tokenWithBearer)

	suite.Equal(claims.Subject, sessionId, "JWT claim subject should match userId from user record")
	suite.Equal("unregistered-onboarded", claims.Type, "JWT claim type should be 'unregistered-onboarded'")
	suite.WithinDuration(clock.Now().Add(30*time.Minute), claims.ExpiresAt.Time, 0, "Expiration should match expectation")

	var user dao.MasterUserRecordDao

	err = suite.TestDB.Model(&dao.MasterUserRecordDao{}).Where("id=?", sessionId).Find(&user).Error
	suite.Require().NoError(err, "Failed to fetch user data")

	var loginResponse response.LoginResponse
	err = json.Unmarshal(rec.Body.Bytes(), &loginResponse)
	suite.Require().NoError(err, "Failed to unmarshal login response")

	suite.Equal(user.LedgerCustomerNumber, loginResponse.CustomerNo, "Customer number should match")
	suite.Equal(user.UserStatus, loginResponse.Status, "User status should be ACTIVE")
	suite.Equal("XXX-XXX-7890", loginResponse.MobileNo, "User phone number should be masked")
	suite.Equal("us**@email.com", loginResponse.Email, "User email should match")
	suite.False(loginResponse.IsUserRegistered, "User is not registered")
	suite.False(loginResponse.IsUserOnboarding, "User is not considered onboarding")

	var loginRecord dao.MasterUserLoginDao
	err = suite.TestDB.Model(&dao.MasterUserLoginDao{}).Where("user_id = ?", sessionId).First(&loginRecord).Error
	suite.Require().NoError(err, "Failed to fetch user data")
	suite.Equal(sessionId, loginRecord.UserId, "Login Record UserId must match user record")
	suite.NotEmpty(loginRecord.IP, "Login record must contain IP address from request")
	suite.Require().Equal(loginResponse.Email, "us**@email.com", "Masked user email should be sent if keyId is not present")
	suite.Require().Equal(loginResponse.MobileNo, "XXX-XXX-7890", "Masked mobile no should be sent if keyId is not present")
}

func (suite *IntegrationTestSuite) TestLoginForUnexpectedStatusUser() {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	suite.Require().NoError(err, "Failed to encrypt password")

	sessionId := uuid.New().String()

	userRecord := dao.MasterUserRecordDao{
		Id:                   sessionId,
		FirstName:            "Foo",
		LastName:             "Bar",
		Email:                "user@email.com",
		Password:             hashedPassword,
		LedgerCustomerNumber: "1234567890",
		UserStatus:           "INACTIVE",
	}

	config.Config.Jwt.SecreteKey = "example_key"

	err = suite.TestDB.Select("id", "first_name", "last_name", "email", "password", "ledger_customer_number", "user_status").Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	updateRequest := request.LoginRequest{
		Username: "user@email.com",
		Password: "Password123!",
	}
	requestBody, err := json.Marshal(updateRequest)
	suite.Require().NoError(err, "Failed to marshall request")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/login")

	err = handler.Login(c)
	typedErr, ok := err.(response.ErrorResponse)
	suite.Equal(ok, true)
	suite.Equal(typedErr.Message, "User must have active ledger to login")
	suite.Equal(http.StatusPreconditionFailed, typedErr.StatusCode)
}

func (suite *IntegrationTestSuite) TestLoginForUnexpectedStatusUserWithIncorrectPassword() {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	suite.Require().NoError(err, "Failed to encrypt password")

	sessionId := uuid.New().String()

	userRecord := dao.MasterUserRecordDao{
		Id:                   sessionId,
		FirstName:            "Foo",
		LastName:             "Bar",
		Email:                "user@email.com",
		Password:             hashedPassword,
		LedgerCustomerNumber: "1234567890",
		UserStatus:           "ONBOARDING_IN_PROGRESS",
	}

	config.Config.Jwt.SecreteKey = "example_key"

	err = suite.TestDB.Select("id", "first_name", "last_name", "email", "password", "ledger_customer_number", "user_status").Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	updateRequest := request.LoginRequest{
		Username: "user@email.com",
		Password: "WrongPassword456!",
	}
	requestBody, err := json.Marshal(updateRequest)
	suite.Require().NoError(err, "Failed to marshall request")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/login")

	err = handler.Login(c)
	typedErr, ok := err.(response.ErrorResponse)
	suite.Equal(true, ok)
	suite.Equal("Invalid Credentials.", typedErr.Message)
	suite.Equal(typedErr.StatusCode, http.StatusUnauthorized)
}

func (suite *IntegrationTestSuite) TestLoginForOnboardingUser() {
	unfreeze := clock.FreezeNow()
	defer unfreeze()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	suite.Require().NoError(err, "Failed to encrypt password")

	sessionId := uuid.New().String()

	userRecord := dao.MasterUserRecordDao{
		Id:         sessionId,
		FirstName:  "Foo",
		LastName:   "Bar",
		Email:      "user@email.com",
		Password:   hashedPassword,
		UserStatus: "USER_CREATED",
	}

	config.Config.Jwt.SecreteKey = "example_key"

	err = suite.TestDB.Select("id", "first_name", "last_name", "email", "password", "user_status").Create(&userRecord).Error
	suite.Require().NoError(err, "Failed to insert test user")

	updateRequest := request.LoginRequest{
		Username: "user@email.com",
		Password: "Password123!",
	}
	requestBody, err := json.Marshal(updateRequest)
	suite.Require().NoError(err, "Failed to marshall request")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/login")

	err = handler.Login(c)
	suite.NoError(err, "Handler should not return an error")
	suite.Equal(http.StatusOK, rec.Code)

	var user dao.MasterUserRecordDao

	err = suite.TestDB.Model(&dao.MasterUserRecordDao{}).Where("id=?", sessionId).Find(&user).Error
	suite.Require().NoError(err, "Failed to fetch user data")

	var loginResponse response.LoginResponse
	err = json.Unmarshal(rec.Body.Bytes(), &loginResponse)
	suite.Require().NoError(err, "Failed to unmarshal login response")

	suite.Equal(user.UserStatus, loginResponse.Status, "User status should match")
	suite.False(loginResponse.IsUserRegistered, "User is not considered registered")
	suite.True(loginResponse.IsUserOnboarding, "User is considered onboarding")

	tokenWithBearer := rec.Header().Get("Authorization")
	suite.NotEmpty(tokenWithBearer, "Authorization token should be present in response headers")

	claims := security.GetClaimsFromToken(tokenWithBearer)

	suite.Equal(sessionId, claims.Subject, "JWT claim subject should match userId from user record")
	suite.Equal("recover-onboarding", claims.Type, "JWT claim type should be 'recover-onboarding")
	suite.WithinDuration(clock.Now().Add(30*time.Minute), claims.ExpiresAt.Time, 0, "Expiration should match expectation")
}
