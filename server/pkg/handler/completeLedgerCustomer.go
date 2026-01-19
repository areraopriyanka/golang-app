package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/sardine"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/google/uuid"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

// @Summary CompleteLedgerCustomer
// @Description Adds a checking account, registers a public key to the ledger, and adds a cardholder and card to complete user onboarding
// @Tags onboarding
// @Accept json
// @Produce json
// @Param userId path string true "user id of onboarding customer"
// @Param Authorization header string true "Bearer token for user authentication"
// @Param payload body CompleteLedgerCustomerRequest true "CompleteLedgerCustomerRequest payload"
// @Success 201 {object} CompleteLedgerCustomerResponse
// @header 201 {string} Authorization "Bearer token for user authentication"
// @header 201 {string} X-Token-Expiration "Token expiration timestamp"
// @Failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse "Public key is associated with a different user"
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /onboarding/customer/{userId}/complete-ledger-customer [put]
// @Router /onboarding/customer/complete-ledger-customer [put]
func CompleteLedgerCustomer(c echo.Context) error {
	// TODO: Remove path param logic later and use userId from JWT only.
	userId := c.Param("userId")
	var cc *security.OnboardingUserContext
	var ok bool
	if userId == "" {
		cc, ok = c.(*security.OnboardingUserContext)
		if !ok {
			return response.UnauthorizedError("Failed to get user Id from custom context")
		}
		userId = cc.UserId
	}

	logger := logging.GetEchoContextLogger(c)

	// Get the request
	payload := new(CompleteLedgerCustomerRequest)
	bindErr := c.Bind(payload)
	if bindErr != nil {
		logger.Error("Error decoding request data", "error", bindErr.Error())
		return response.BadRequestInvalidBody
	}

	// Handle validation errors
	if err := c.Validate(payload); err != nil {
		return err
	}

	user, errResponse := dao.RequireUserWithState(
		userId, constant.CARD_AGREEMENTS_REVIEWED,
	)
	if errResponse != nil {
		return errResponse
	}

	userAccountCard, err := dao.UserAccountCardDao{}.FindOneActiveByUserId(db.DB, userId)
	if err != nil {
		logger.Error("Error while fetching userAccountCard record", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error while fetching userAccountCard record: %s", err.Error()), errtrace.Wrap(err))
	}

	if userAccountCard == nil {
		userAccountCard = &dao.UserAccountCardDao{}
	}

	err = setCustomerSettingPciCheckTrue(*user, logger)
	if err != nil {
		return err
	}

	// Create the initial checking account if missing
	if userAccountCard.AccountNumber == "" {
		logger.Info("Creating initial checking account")
		accountNumber, err := createLedgerCheckingAccount(*user, logger)
		if err != nil {
			return err
		}
		userAccountCard.AccountNumber = accountNumber
	} else {
		logger.Info("User already has an initial checking account", "accountNumber", userAccountCard.AccountNumber)
	}

	// Get the public key
	var existingPublicKey dao.UserPublicKey
	keyExists := db.DB.Model(&dao.UserPublicKey{}).Where("public_key=?", payload.PublicKey).First(&existingPublicKey)

	// Create the public key if missing
	if keyExists.Error == nil {
		if existingPublicKey.UserId == userId {
			logger.Info("Public key already registered for userId", "userId", userId)
		} else {
			logger.Error("Public key already associated with a different userId. Ledger will not permit re-use.", "userId", existingPublicKey.UserId)
			return response.GenerateErrResponse(constant.PUBLIC_KEY_IN_USE, "", "Public key already associated with a different userId", http.StatusForbidden, errtrace.New(""))
		}
	} else if errors.Is(keyExists.Error, gorm.ErrRecordNotFound) {
		// Not finding the public key is good - they must be unique (even across users). Any other errors should 500.
		logger.Info("Registering public key")
		err := registerPublicKey(*user, payload.PublicKey, logger, userId)
		if err != nil {
			return err
		}
	} else {
		logger.Error("Error checking for public key", "error", keyExists.Error.Error())
		return response.InternalServerError(fmt.Sprintf("Error checking for public key: %s", keyExists.Error.Error()), errtrace.Wrap(keyExists.Error))
	}

	// Get the CardHolderId
	var cardHolderId string

	// Add the card holder if missing
	if userAccountCard.CardHolderId == "" {
		logger.Info("Adding card holder")
		var err error
		cardHolderId, err = addCardHolder(*user, userAccountCard.AccountNumber, logger)
		if err != nil {
			return err
		}
		userAccountCard.CardHolderId = cardHolderId
	} else {
		logger.Info("CardHolderId already set for user", "userId", user.Id, "cardHolderId", userAccountCard.CardHolderId)
	}

	// Add the card if missing
	var cardId string
	if userAccountCard.CardId == "" {
		logger.Info("Adding card")
		var err error
		cardId, err = addCard(*user, logger, *userAccountCard)
		if err != nil {
			return err
		}
		userAccountCard.CardId = cardId
	} else {
		logger.Warn("Card already added for the user", "userId", user.Id, "cardId", userAccountCard.CardId)
		return response.GenerateErrResponse(constant.CARD_ALREADY_ADDED, constant.CARD_ALREADY_ADDED_MSG, "", http.StatusConflict, errtrace.New(""))
	}

	// Mark user status as active
	if user.UserStatus == constant.CARD_AGREEMENTS_REVIEWED {
		updateResult := db.DB.Model(&user).Update("user_status", constant.ACTIVE)
		if updateResult.Error != nil {
			logger.Error("Error saving user's status", "error", updateResult.Error.Error())
			return response.InternalServerError(fmt.Sprintf("Error while fetching userAccountCard record: %s", updateResult.Error.Error()), errtrace.Wrap(updateResult.Error))
		}

		if updateResult.RowsAffected == 0 {
			logger.Error("The user record not updated to status ACTIVE")
			return response.InternalServerError("The user record not updated to status ACTIVE", errtrace.New(""))
		}
	}

	err = reportFeedbackToSardine(*user, logger, payload.SardineSessionKey, sardine.FeedbackStatusApproved)
	if err != nil {
		// Sardine errors should not fail the request
		logging.Logger.Warn("Failed to report feedback to Sardine", "error", err.Error())
	}

	// Update the user's membership status once onboarding is completed
	err = createMembershipRecord(userId)
	if err != nil {
		logger.Error("error while creating membership record", "error", err)
		return response.InternalServerError(fmt.Sprintf("error while creating membership record: error: %s", err.Error()), errtrace.Wrap(err))
	}

	now := clock.Now()

	token, err := security.GenerateOnboardedJwt(user.Id, payload.PublicKey, &now)
	if err != nil {
		logging.Logger.Error("Error generating JWT", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error generating JWT: %s", err.Error()), errtrace.Wrap(err))
	}
	c.Set("SkipTokenReset", true)
	c.Response().Header().Set("Authorization", "Bearer "+token)

	return c.JSON(http.StatusCreated, CompleteLedgerCustomerResponse{CardHolderId: cardHolderId, CardId: cardId})
}

func createLedgerCheckingAccount(user dao.MasterUserRecordDao, logger *slog.Logger) (string, error) {
	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDLedgerApiClient(config.Config.Ledger, ledgerParamsBuilder)

	// `accountType` could accept other enums (e.g., SAVINGS; WALLET; etc.), but for now, we're just creating
	// the initial checking account for the user so that we can proceed with device registration (`CustomerService.AddUserKey`)
	// See https://apidocs.netxd.com/developers/docs/account_apis/AddAccount%20Consumer for more info
	accountType := "CHECKING"
	// This default accountName value came from the current FE config 725c4ab7ee:src/res/strings/index.tsx
	accountName := "Default DreamFi Account"
	req := ledgerClient.BuildAddConsumerAccountRequest(user.LedgerCustomerNumber, accountType, accountName)
	resp, addAccountErr := ledgerClient.AddConsumerAccount(req)
	if addAccountErr != nil {
		logger.Error("Error adding account for customer to ledger", "error", addAccountErr.Error())
		return "", response.InternalServerError(fmt.Sprintf("Error adding account for customer to ledger: %s", addAccountErr.Error()), errtrace.Wrap(addAccountErr))

	}
	if resp.Error != nil {
		logger.Error("error message from AddConsumerAccount", "code", resp.Error.Code, "error", resp.Error.Message)
		return "", response.InternalServerError(fmt.Sprintf("error message from AddConsumerAccount, error: %s", resp.Error.Message), errtrace.New(""))
	}
	if resp.Result == nil {
		logger.Error("result from AddConsumerAccount is missing")
		return "", response.InternalServerError("result from AddConsumerAccount is missing", errtrace.New(""))
	}

	accountNumber := resp.Result.AccountNumber
	accountId := resp.Result.ID
	logger.Info("Created ledger customer checking account with number", "accountNumber", accountNumber)

	userAccountCardRecord := dao.UserAccountCardDao{
		UserId:        user.Id,
		AccountId:     accountId,
		AccountNumber: accountNumber,
		AccountStatus: "ACTIVE",
	}

	result := db.DB.Create(&userAccountCardRecord)
	if result.Error != nil {
		logger.Error(constant.INTERNAL_SERVER_ERROR, "error", result.Error)
		return "", response.InternalServerError(fmt.Sprintf("Error while creating userAccountCard record: %s", result.Error), errtrace.Wrap(result.Error))

	}

	return accountNumber, nil
}

func setCustomerSettingPciCheckTrue(user dao.MasterUserRecordDao, logger *slog.Logger) error {
	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDLedgerApiClient(config.Config.Ledger, ledgerParamsBuilder)

	resp, err := ledgerClient.UpdateCustomerSettings(ledger.UpdateCustomerSettingsRequest{
		CustomerId: user.LedgerCustomerNumber,
		PciCheck:   true,
	})
	if err != nil {
		logger.Error("error setting PCICheck to true", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("error setting PCICheck to true: %s", err.Error()), errtrace.Wrap(err))
	}

	if resp.Error != nil {
		logger.Error("error message from setting PCICheck to true", "code", resp.Error.Code, "msg", resp.Error.Message)
		return response.InternalServerError(fmt.Sprintf("error message from setting PCICheck to true: %s", resp.Error.Message), errtrace.New(""))
	}

	logger.Info("Received successful message form UpdateCustomerSettings PCICheck", "msg", resp.Result.Message)

	return nil
}

func registerPublicKey(user dao.MasterUserRecordDao, publicKey string, logger *slog.Logger, userId string) error {
	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDLedgerApiClient(config.Config.Ledger, ledgerParamsBuilder)

	request := ledger.BuildAddUserKeyRequest(user.Email, publicKey)
	resp, err := ledgerClient.AddUserKey(request)
	if err != nil {
		logger.Error("Error adding user's public key to ledger", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error adding user's public key to ledger: %s", err.Error()), errtrace.Wrap(err))
	}

	if resp.Error != nil {
		logger.Error("error message from ledger AddUserKey", "code", resp.Error.Code, "message", resp.Error.Message)
		return response.InternalServerError(fmt.Sprintf("error message from ledger AddUserKey. Error: %s", resp.Error.Message), errtrace.New(""))
	}

	encryptedApiKey, err := utils.EncryptKmsBinary(resp.Result.ApiKey)
	if err != nil {
		logger.Error("failed to encrypt ledger AddUserKey apiKey", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("failed to encrypt ledger AddUserKey apiKey: %s", err.Error()), errtrace.Wrap(err))
	}

	userPublicKey := dao.UserPublicKey{
		UserId:             userId,
		KeyId:              resp.Result.KeyID,
		KmsEncryptedApiKey: []byte(encryptedApiKey),
		PublicKey:          publicKey,
	}

	createResult := db.DB.Select("user_id", "kms_encrypted_api_key", "key_id", "public_key").Create(&userPublicKey)
	if createResult.Error != nil {
		logger.Error("Error saving user's public key", "error", createResult.Error.Error())
		return response.InternalServerError(fmt.Sprintf("Error saving user's public key: %s", createResult.Error.Error()), errtrace.Wrap(createResult.Error))
	}
	return nil
}

func addCardHolder(user dao.MasterUserRecordDao, accountNumber string, logger *slog.Logger) (string, error) {
	payload := generateAddCardHolderPayload(user, accountNumber)

	payloadDataJSON, marshallErr := json.Marshal(payload)
	if marshallErr != nil {
		logger.Error("An error occurred while marshalling the payload", "error", marshallErr.Error())
		return "", response.InternalServerError(fmt.Sprintf("An error occurred while marshalling the payload: %s", marshallErr.Error()), errtrace.Wrap(marshallErr))
	}

	request := &ledger.Request{
		Id:     constant.LEDGER_REQUEST_ID,
		Method: "ledger.CARD.request",
		Params: ledger.Params{
			Payload: payloadDataJSON,
		},
	}

	signatureErr := request.SignRequestWithMiddlewareKey()
	if signatureErr != nil {
		logger.Error("failed to sign json payload", "error", signatureErr.Error())
		return "", response.InternalServerError(fmt.Sprintf("failed to sign json payload: %s", signatureErr.Error()), errtrace.Wrap(signatureErr))
	}

	var responseData ledger.AddCardHolderLedgerResponse
	err := ledger.CallLedgerAPIWithUrlAndGetTypedResponse(request, config.Config.Ledger.CardsEndpoint, &responseData, logger)
	if err != nil {
		logger.Error("Error while calling ledger API", "error", err.Error())
		return "", response.InternalServerError(fmt.Sprintf("Error while calling ledger API: %s", err.Error()), errtrace.Wrap(err))
	}

	if responseData.Error != nil {
		logger.Error("The ledger responded with an error", "code", responseData.Error.Code, "msg", responseData.Error.Message)
		return "", response.InternalServerError(fmt.Sprintf("The ledger responded with an error, msg: %s", responseData.Error.Message), errtrace.New(""))
	}

	// Store cardHolderId in DB
	err = updateCardHolderIdInDB(user.Id, responseData.Result.CardHolderId)
	if err != nil {
		logger.Error("Error while inserting cardHolderId", "error", err.Error())
		return "", response.InternalServerError(fmt.Sprintf("Error while inserting cardHolderId: %s", err.Error()), errtrace.Wrap(err))
	}

	logger.Info("Card holder added successfully. CardHolderId", "cardHolderId", responseData.Result.CardHolderId)
	return responseData.Result.CardHolderId, nil
}

func addCard(user dao.MasterUserRecordDao, logger *slog.Logger, cardHolder dao.UserAccountCardDao) (string, error) {
	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	cardsClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, ledgerParamsBuilder)

	payload := cardsClient.BuildAddCardRequest(user.LedgerCustomerNumber, cardHolder.AccountNumber, cardHolder.CardHolderId)

	responseData, err := cardsClient.AddCard(payload)
	if err != nil {
		logger.Error("Error while calling ledger API", "error", err.Error())
		return "", response.InternalServerError(fmt.Sprintf("Error while calling ledger API: %s", err.Error()), errtrace.Wrap(err))
	}

	if responseData.Error != nil {
		logger.Error("The ledger responded with an error", "code", responseData.Error.Code, "msg", responseData.Error.Message)
		return "", response.InternalServerError(fmt.Sprintf("The ledger responded with an error: %s", responseData.Error.Message), errtrace.New(""))
	}

	// Update cardId in DB
	err = updateCardIdInDB(user.Id, responseData.Result.Card.CardId)
	if err != nil {
		logger.Error("Error while inserting cardId", "error", err.Error())
		return "", response.InternalServerError(fmt.Sprintf("Error while inserting cardId: %s", err.Error()), errtrace.Wrap(err))
	}

	logger.Info("Card added successfully", "cardId", responseData.Result.Card.CardId)
	return responseData.Result.Card.CardId, nil
}

func reportFeedbackToSardine(user dao.MasterUserRecordDao, logger *slog.Logger, sardineSessionKey string, sardineFeedbackStatus sardine.FeedbackStatus) error {
	// Create sardine client
	client, err := utils.NewSardineClient(config.Config.Sardine)
	if err != nil {
		logger.Error("Error occurred while creating sardine client", "error", err.Error())
		return errtrace.Wrap(err)
	}

	requestBody := generateSardineFeedbackPayload(user, sardineSessionKey, sardineFeedbackStatus)

	sardineResponse, err := client.PostCustomerFeedbackWithResponse(context.Background(), requestBody)
	if err != nil {
		logger.Error("Error occurred while calling sardine API", "error", err.Error())
		return errtrace.Wrap(err)
	}

	var res bytes.Buffer
	if err := json.Indent(&res, sardineResponse.Body, "", "  "); err != nil {
		logger.Error("Error occurred while formatting sardine response body", "error", err.Error())
		return errtrace.Wrap(err)
	}
	logger.Info("Received from Sardine", "statusCode", sardineResponse.StatusCode(), "response", res.String())

	return nil
}

func setLedgerPasswordForUser(password string, user dao.MasterUserRecordDao) (err error) {
	encryptedPassword, encryptionErr := utils.EncryptKmsBinary(password)
	if encryptionErr != nil {
		return errtrace.Wrap(encryptionErr)
	}
	updateResult := db.DB.Model(&user).Update("kms_encrypted_ledger_password", encryptedPassword)
	if updateResult.Error != nil {
		return errtrace.Wrap(updateResult.Error)
	}
	return nil
}

func generateAddCardHolderPayload(user dao.MasterUserRecordDao, accountNumber string) ledger.AddCardHolderPayload {
	payload := ledger.AddCardHolderPayload{
		Reference:       utils.GenerateReferenceNumber(),
		TransactionType: "ADD_CARD_HOLDER_WITH_PRIMARY_ADDRESS",
		CustomerId:      user.LedgerCustomerNumber,
		AccountNumber:   accountNumber,
		Product:         config.Config.Ledger.CardsProduct,
		Channel:         config.Config.Ledger.CardsChannel,
		Program:         config.Config.Ledger.CardsProgram,
	}
	return payload
}

func generateSardineFeedbackPayload(user dao.MasterUserRecordDao, sardineSessionKey string, sardineFeedbackStatus sardine.FeedbackStatus) sardine.PostCustomerFeedbackJSONRequestBody {
	sessionKey := sardineSessionKey
	feedbackScope := (sardine.FeedbackScope)("user")
	feedbackType := sardine.FeedbackTypeOnboarding
	kind := sardine.PostCustomerFeedbackJSONBodyKindFlow
	feedbackId := fmt.Sprintf("sardine.feedback_%d", clock.Now().UnixNano())
	timeMillis := clock.Now().UnixMilli()

	requestBody := sardine.PostCustomerFeedbackJSONRequestBody{
		Customer: &struct {
			Id *string "json:\"id,omitempty\""
		}{
			Id: &user.Id,
		},
		Feedback: &sardine.Feedback{
			Id:         &feedbackId,
			Scope:      &feedbackScope,
			Status:     &sardineFeedbackStatus,
			Type:       &feedbackType,
			TimeMillis: &timeMillis,
		},
		Kind:       &kind,
		SessionKey: sessionKey,
	}
	return requestBody
}

func updateCardHolderIdInDB(userId, cardHolderId string) error {
	result := db.DB.Model(dao.UserAccountCardDao{}).Where("user_id=? AND account_status=?", userId, "ACTIVE").Update("card_holder_id", cardHolderId)
	if result.Error != nil {
		return errtrace.Wrap(result.Error)
	}

	return nil
}

func updateCardIdInDB(userId, cardId string) error {
	result := db.DB.Model(dao.UserAccountCardDao{}).Where("user_id=? AND account_status=?", userId, "ACTIVE").Update("card_id", cardId)
	if result.Error != nil {
		return errtrace.Wrap(result.Error)
	}

	return nil
}

func createMembershipRecord(userId string) error {
	membershipRecord := dao.UserMembershipDao{
		Id:               uuid.New().String(),
		UserID:           userId,
		MembershipStatus: constant.SUBSCIBED,
	}
	result := db.DB.Create(&membershipRecord)
	return errtrace.Wrap(result.Error)
}

// TODO: mfp? It doesn't seem to be needed, and the expected format is not documented and can error easily.
type CompleteLedgerCustomerRequest struct {
	PublicKey         string `json:"publicKey" validate:"required,max=255"`
	SardineSessionKey string `json:"sardineSessionKey" validate:"required"`
}

type CompleteLedgerCustomerResponse struct {
	CardHolderId string `json:"cardHolderId" validate:"required"`
	CardId       string `json:"cardId" validate:"required"`
}
