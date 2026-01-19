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
	"process-api/pkg/ledger"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestSetCardPin() {
	unfreeze := clock.FreezeNow()
	defer unfreeze()
	defer SetupMockForLedger(suite).Close()

	config.Config.Aws.KmsEncryptionKeyId = "test-kms-encryption-key-id"

	user := suite.createUser()

	// Create cardHolder Record
	cardHolder := dao.UserAccountCardDao{
		CardHolderId:  "CH000006009011111",
		CardId:        "6f586be7bf1c44b8b4ea11b2e2510e25",
		UserId:        user.Id,
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
	}

	err := suite.TestDB.Create(&cardHolder).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	reference := uuid.New().String()
	payloadData := ledger.ChangePinRequest{
		Reference:       reference,
		TransactionType: "PIN_CHANGE",
		CustomerId:      "100000000012345",
		AccountNumber:   "123456789012345",
		CardId:          "6f586be7bf1c44b8b4ea11b2e2510e25",
		NewPIN:          "1234",
		Product:         "DEFAULT",
		Channel:         "VISA_DPS",
		Program:         "DEFAULT",
		IsEncrypt:       true,
	}

	jsonPayloadBytes, err := json.Marshal(payloadData)
	suite.Require().NoError(err, "Failed to marshal test payload")

	jsonPayload := string(jsonPayloadBytes)

	payloadId := uuid.New().String()
	payloadRecord := dao.SignablePayloadDao{
		Id:      payloadId,
		Payload: jsonPayload,
	}

	err = suite.TestDB.Create(&payloadRecord).Error
	suite.Require().NoError(err, "Failed to insert test payload")

	otpSessionId := uuid.New().String()

	currentTime := clock.Now()

	userOtpRecord := dao.MasterUserOtpDao{
		OtpId:     otpSessionId,
		Otp:       "123456",
		OtpStatus: constant.OTP_VERIFIED,
		Email:     "email@example.com",
		ApiPath:   "/account/cards/pin/otp",
		UserId:    user.Id,
		IP:        "1.1.1.1",
		CreatedAt: currentTime,
		UsedAt:    &currentTime,
	}

	err = suite.TestDB.Select("otp_id", "otp", "otp_status", "email", "api_path", "user_id", "ip", "created_at", "used_at").Create(&userOtpRecord).Error
	suite.Require().NoError(err, "Failed to insert otp record")

	setCardPinRequest := request.SetCardPinRequest{
		Signature: "example_signature",
		PayloadId: payloadId,
		OtpId:     otpSessionId,
		Otp:       "123456",
	}
	requestBody, err := json.Marshal(setCardPinRequest)
	suite.Require().NoError(err, "Failed to marshall card pin request")

	encryptedApiKey, err := utils.EncryptKmsBinary("c077ad8b3d6f40c9896f5fb475f738d6")
	suite.Require().NoError(err, "Failed to encrypt example apiKey")
	suite.T().Logf("Encrypted API Key: %v", encryptedApiKey)

	userPublicKey := dao.UserPublicKey{
		UserId:             user.Id,
		KmsEncryptedApiKey: []byte(encryptedApiKey),
		KeyId:              "exampleKeyId",
		PublicKey:          "examplePublicKey",
	}

	err = suite.TestDB.Create(&userPublicKey).Error
	suite.Require().NoError(err, "Failed to insert user public key record")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/pin", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/pin")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, "examplePublicKey", c)

	err = handler.SetCardPin(customContext)

	suite.Require().NoError(err, "Handler should not return an error")

	suite.Require().Equal(http.StatusOK, rec.Code, "Expected status code 200 OK")

	var responseBody response.SetCardPinResponse
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().True(responseBody.PinSet, "Expected PinSet to be true")

	var otpRecord dao.MasterUserOtpDao
	err = suite.TestDB.Model(&dao.MasterUserOtpDao{}).Where("otp_id = ?", userOtpRecord.OtpId).First(&otpRecord).Error
	suite.Require().NoError(err, "Failed to get otp record from db")
	suite.WithinDuration(*otpRecord.ChallengeExpiredAt, clock.Now(), 0, "OTP record challenge expired at set to now when otp challenge is used successfully")
}

func (suite *IntegrationTestSuite) TestSetCardPinInvalidOtpFailure() {
	defer SetupMockForLedger(suite).Close()

	config.Config.Aws.KmsEncryptionKeyId = "test-kms-encryption-key-id"

	user := suite.createUser()

	// Create cardHolder sRecord
	cardHolder := dao.UserAccountCardDao{
		CardHolderId:  "CH000006009011111",
		CardId:        "6f586be7bf1c44b8b4ea11b2e2510e25",
		UserId:        user.Id,
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
	}

	err := suite.TestDB.Create(&cardHolder).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	otpSessionId := uuid.New().String()

	userOtpRecord := dao.MasterUserOtpDao{
		OtpId:     otpSessionId,
		Otp:       "123456",
		OtpStatus: constant.OTP_SENT,
		Email:     "user@example.com",
		ApiPath:   "/account/cards/pin/otp",
		UserId:    user.Id,
		IP:        "1.1.1.1",
		CreatedAt: clock.Now(),
	}

	err = suite.TestDB.Select("otp_id", "otp", "otp_status", "email", "api_path", "user_id", "ip", "created_at").Create(&userOtpRecord).Error
	suite.Require().NoError(err, "Failed to insert otp record")

	reference := uuid.New().String()
	payloadData := ledger.ChangePinRequest{
		Reference:       reference,
		TransactionType: "PIN_CHANGE",
		CustomerId:      "100000000012345",
		AccountNumber:   "123456789012345",
		CardId:          "6f586be7bf1c44b8b4ea11b2e2510e25",
		NewPIN:          "1234",
		Product:         "DEFAULT",
		Channel:         "VISA_DPS",
		Program:         "DEFAULT",
		IsEncrypt:       true,
	}

	jsonPayloadBytes, err := json.Marshal(payloadData)
	suite.Require().NoError(err, "Failed to marshal test payload")

	jsonPayload := string(jsonPayloadBytes)

	payloadId := uuid.New().String()
	payloadRecord := dao.SignablePayloadDao{
		Id:      payloadId,
		Payload: jsonPayload,
	}

	err = suite.TestDB.Create(&payloadRecord).Error
	suite.Require().NoError(err, "Failed to insert test payload")

	setCardPinRequest := request.SetCardPinRequest{
		Signature: "example_signature",
		PayloadId: payloadId,
		OtpId:     otpSessionId,
		Otp:       "123456",
	}
	requestBody, err := json.Marshal(setCardPinRequest)
	suite.Require().NoError(err, "Failed to marshall card pin request")

	encryptedApiKey, err := utils.EncryptKmsBinary("c077ad8b3d6f40c9896f5fb475f738d6")
	suite.Require().NoError(err, "Failed to encrypt example apiKey")
	suite.T().Logf("Encrypted API Key: %v", encryptedApiKey)

	userPublicKey := dao.UserPublicKey{
		UserId:             user.Id,
		KmsEncryptedApiKey: []byte(encryptedApiKey),
		KeyId:              "exampleKeyId",
		PublicKey:          "examplePublicKey",
	}

	err = suite.TestDB.Create(&userPublicKey).Error
	suite.Require().NoError(err, "Failed to insert user public key record")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/pin", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/pin")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, "examplePublicKey", c)

	err = handler.SetCardPin(customContext)
	suite.Require().NotNil(err, "Handler should not return an error")

	e.HTTPErrorHandler(err, c)

	var responseBody response.BadRequestErrors
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(http.StatusBadRequest, rec.Code, "Expected status code 400 badRequest")
	suite.Equal("invalid_otp", responseBody.Errors[0].FieldName, "Invalid field name")
	suite.Equal("Please enter valid one time password.", responseBody.Errors[0].Error, "Invalid error message")
}

func (suite *IntegrationTestSuite) TestSetCardPinChallengeExpiredOtpFailure() {
	defer SetupMockForLedger(suite).Close()

	config.Config.Aws.KmsEncryptionKeyId = "test-kms-encryption-key-id"

	user := suite.createUser()

	// Create userAccountCard Record
	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH000006009011111",
		CardId:        "6f586be7bf1c44b8b4ea11b2e2510e25",
		UserId:        user.Id,
		AccountNumber: "123456789012345",
		AccountStatus: "ACTIVE",
	}

	err := suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")

	otpSessionId := uuid.New().String()

	currentTime := clock.Now()

	userOtpRecord := dao.MasterUserOtpDao{
		OtpId:              otpSessionId,
		Otp:                "123456",
		OtpStatus:          constant.OTP_VERIFIED,
		Email:              "user@example.com",
		ApiPath:            "/account/cards/pin/otp",
		UserId:             user.Id,
		IP:                 "1.1.1.1",
		CreatedAt:          currentTime,
		UsedAt:             &currentTime,
		ChallengeExpiredAt: &currentTime,
	}

	err = suite.TestDB.Select("otp_id", "otp", "otp_status", "email", "api_path", "user_id", "ip", "created_at", "used_at", "challenge_expired_at").Create(&userOtpRecord).Error
	suite.Require().NoError(err, "Failed to insert otp record")

	reference := uuid.New().String()
	payloadData := ledger.ChangePinRequest{
		Reference:       reference,
		TransactionType: "PIN_CHANGE",
		CustomerId:      "100000000012345",
		AccountNumber:   "123456789012345",
		CardId:          "6f586be7bf1c44b8b4ea11b2e2510e25",
		NewPIN:          "1234",
		Product:         "DEFAULT",
		Channel:         "VISA_DPS",
		Program:         "DEFAULT",
		IsEncrypt:       true,
	}

	jsonPayloadBytes, err := json.Marshal(payloadData)
	suite.Require().NoError(err, "Failed to marshal test payload")

	jsonPayload := string(jsonPayloadBytes)

	payloadId := uuid.New().String()
	payloadRecord := dao.SignablePayloadDao{
		Id:      payloadId,
		Payload: jsonPayload,
	}

	err = suite.TestDB.Create(&payloadRecord).Error
	suite.Require().NoError(err, "Failed to insert test payload")

	setCardPinRequest := request.SetCardPinRequest{
		Signature: "example_signature",
		PayloadId: payloadId,
		OtpId:     otpSessionId,
		Otp:       "123456",
	}
	requestBody, err := json.Marshal(setCardPinRequest)
	suite.Require().NoError(err, "Failed to marshall card pin request")

	encryptedApiKey, err := utils.EncryptKmsBinary("c077ad8b3d6f40c9896f5fb475f738d6")
	suite.Require().NoError(err, "Failed to encrypt example apiKey")

	suite.T().Logf("Encrypted API Key: %v", encryptedApiKey)

	userPublicKey := dao.UserPublicKey{
		UserId:             user.Id,
		KmsEncryptedApiKey: []byte(encryptedApiKey),
		KeyId:              "exampleKeyId",
		PublicKey:          "examplePublicKey",
	}

	err = suite.TestDB.Create(&userPublicKey).Error
	suite.Require().NoError(err, "Failed to insert user public key record")

	e := handler.NewEcho()
	req := httptest.NewRequest(http.MethodPost, "/cards/pin", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/cards/pin")

	customContext := security.GenerateLoggedInRegisteredUserContext(user.Id, "examplePublicKey", c)

	err = handler.SetCardPin(customContext)
	suite.Require().NotNil(err, "Handler should not return an error")

	e.HTTPErrorHandler(err, c)

	var responseBody response.BadRequestErrors
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	suite.Require().NoError(err, "Failed to unmarshal response")

	suite.Require().Equal(http.StatusBadRequest, rec.Code, "Expected status code 400 badRequest")
	suite.Equal("invalid_otp", responseBody.Errors[0].FieldName, "Invalid field name")
	suite.Equal("One Time Password is expired.", responseBody.Errors[0].Error, "Invalid error message")
}
