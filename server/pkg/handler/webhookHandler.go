package handler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/sardine"
	"process-api/pkg/utils"
	"time"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/riverqueue/river"
)

type AccountDetails struct {
	AccountNumber string `json:"accountNumber"`
	HolderName    string `json:"holderName"`
	InstitutionId string `json:"institutionId"`
}

type TransacationNewPayload struct {
	Channel            string         `json:"channel"`
	TransactionType    string         `json:"transactionType"`
	TransactionNumber  string         `json:"transactionNumber"`
	BinNumber          string         `json:"binNumber"`
	CardId             string         `json:"cardID"`
	CreditorAccount    AccountDetails `json:"creditorAccount"`
	DebtorAccount      AccountDetails `json:"debtorAccount"`
	InstructedAmount   int            `json:"instructedAmount"`
	InstructedCurrency string         `json:"instructedCurrency"`
	Mcc                string         `json:"mcc"`
}

type StatementWebhookPayload struct {
	AccountIds []string `json:"accountIds"`
	Timestamp  string   `json:"timestamp"`
}

type SardineCounterparty = struct {
	Address         *sardine.Address                                           `json:"address,omitempty"`
	BusinessName    *string                                                    `json:"businessName,omitempty"`
	DateOfBirth     *string                                                    `json:"dateOfBirth,omitempty"`
	EmailAddress    *string                                                    `json:"emailAddress,omitempty"`
	FirstName       *string                                                    `json:"firstName,omitempty"`
	Id              string                                                     `json:"id"`
	IdDocument      *sardine.IdDocument                                        `json:"idDocument,omitempty"`
	IdentityObject  *string                                                    `json:"identityObject,omitempty"`
	IsEmailVerified *bool                                                      `json:"isEmailVerified,omitempty"`
	IsKycVerified   *bool                                                      `json:"isKycVerified,omitempty"`
	IsPhoneVerified *bool                                                      `json:"isPhoneVerified,omitempty"`
	LastName        *string                                                    `json:"lastName,omitempty"`
	MiddleName      *string                                                    `json:"middleName,omitempty"`
	Nationality     *string                                                    `json:"nationality,omitempty"`
	PaymentMethod   *sardine.PaymentMethod                                     `json:"paymentMethod,omitempty"`
	Phone           *string                                                    `json:"phone,omitempty"`
	Source          *sardine.PostCustomerInformationJSONBodyCounterpartySource `json:"source,omitempty"`
	Tags            *[]sardine.Tags                                            `json:"tags,omitempty"`
	Type            sardine.PostCustomerInformationJSONBodyCounterpartyType    `json:"type"`
}

// Note: the ledger will mark a webhook subscription as "offline" for multiple
// non-200 responses. Even if we fail, we probably should return 200. The
// malformed and unauthenticated requests are the exception to the rule.
func (h *Handler) LedgerWebhookHandler(c echo.Context) error {
	logger := logging.GetEchoContextLogger(c)

	var payload request.LedgerEventPayload
	if err := c.Bind(&payload); err != nil {
		logging.Logger.Error("Invalid request", "Error", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	if utils.VerifyLedgerWebhookSignature(payload) {
		logging.Logger.Debug("Webhook signature verified successfully")
	} else {
		logging.Logger.Error("Webhook signature verification failed")
		return c.NoContent(http.StatusUnauthorized)
	}

	// Check if the payload has data
	if len(payload.Payload) == 0 {
		logger.Error("Payload data is nil")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Payload data is missing"})
	}

	var internalTransactionPayload TransacationNewPayload
	if payload.EventName == "Transaction.NEW" {
		if err := json.Unmarshal(payload.Payload, &internalTransactionPayload); err != nil {
			logger.Error("Failed to unmarshal for transaction.new payload", "err", err)
			return c.NoContent(http.StatusBadRequest)
		}

		ledgerTransactionEventRecord := MapTransactionNewPayloadToLedgerEventDao(payload, internalTransactionPayload)
		if ledgerTransactionEventRecord == nil {
			logger.Error("failed to map event payload to ledger transaction event record")
			return c.NoContent(http.StatusOK)
		}

		err := SaveTransactionEventRecord(*ledgerTransactionEventRecord)
		if err != nil {
			logger.Error("Failed to save transaction event record", "err", err)
			return c.NoContent(http.StatusOK)
		}

		ctx := context.Background()
		_, err = h.RiverClient.Insert(ctx, TransactionMonitoringArgs{
			EventId: ledgerTransactionEventRecord.EventId,
		}, nil)
		if err != nil {
			logger.Error("Failed to start river job", "err", err)
		}
	}

	var internalStatementPayload StatementWebhookPayload
	if payload.EventName == "MONTHLY STATEMENT GENERATION" {
		if err := json.Unmarshal(payload.Payload, &internalStatementPayload); err != nil {
			logger.Error("Failed to unmarshal accountIds from payload", "error", err.Error())
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid userIds payload"})
		}

		if len(internalStatementPayload.AccountIds) == 0 {
			logger.Error("Payload data does not contain any user IDs")
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Payload data is missing"})
		}

		parsedTime, err := time.Parse(time.RFC3339, internalStatementPayload.Timestamp)
		if err != nil {
			logger.Error("Failed to parse timestamp", "timestamp", internalStatementPayload.Timestamp, "error", err.Error())
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid timestamp format"})
		}
		statementDate := parsedTime.AddDate(0, -1, 0)

		year := fmt.Sprint(statementDate.Year())
		monthName := statementDate.Month().String()

		baseUrl := "https://middleware.production.dreamfi.com"
		if h.Env != constant.PROD {
			baseUrl = "https://middleware.sandbox.dreamfi.com"
		}

		ctx := context.Background()
		_, err = h.RiverClient.Insert(ctx, StatementNotificationEmailEnqueueBatchJobArgs{
			AccountIds: internalStatementPayload.AccountIds,
			Month:      monthName,
			Year:       year,
			BaseUrl:    baseUrl,
		}, nil)
		if err != nil {
			logger.Error("Failed to start river job", "err", err)
		}
	}

	return c.NoContent(http.StatusOK)
}

func MapTransactionNewPayloadToLedgerEventDao(eventPayload request.LedgerEventPayload, internalPayload TransacationNewPayload) *dao.LedgerTransactionEventDao {
	// TODO: Need to potentially report as fradulent to sardine in instances where user account record is marked as closed
	// but we are still receiving ledger events for said account
	ledgerTransactionEventRecord := dao.LedgerTransactionEventDao{
		EventId:            eventPayload.EventId,
		Channel:            internalPayload.Channel,
		TransactionType:    internalPayload.TransactionType,
		TransactionNumber:  internalPayload.TransactionNumber,
		BinNumber:          internalPayload.BinNumber,
		CardId:             internalPayload.CardId,
		InstructedAmount:   internalPayload.InstructedAmount,
		InstructedCurrency: internalPayload.InstructedCurrency,
		Mcc:                internalPayload.Mcc,
		RawPayload:         eventPayload.Payload,
	}

	switch internalPayload.TransactionType {
	case "ACH_OUT":
		ledgerTransactionEventRecord.IsOutward = true
		encryptedBankAccountNumber, err := utils.EncryptKmsBinary(internalPayload.CreditorAccount.AccountNumber)
		if err != nil {
			logging.Logger.Error("Failed to encrypt bank account number to binary for ach out", "err", err)
		}
		ledgerTransactionEventRecord.AccountRoutingNumber = internalPayload.DebtorAccount.InstitutionId
		ledgerTransactionEventRecord.ExternalBankAccountNumber = encryptedBankAccountNumber
		ledgerTransactionEventRecord.ExternalBankAccountName = internalPayload.CreditorAccount.HolderName
		ledgerTransactionEventRecord.ExternalBankAccountRoutingNumber = internalPayload.CreditorAccount.InstitutionId
		ledgerTransactionEventRecord.AccountNumber = internalPayload.DebtorAccount.AccountNumber

		return &ledgerTransactionEventRecord

	case "ACH_PULL":
		ledgerTransactionEventRecord.IsOutward = false
		encryptedBankAccountNumber, err := utils.EncryptKmsBinary(internalPayload.DebtorAccount.AccountNumber)
		if err != nil {
			logging.Logger.Error("Failed to encrypt bank account number to binary for ach pull", "err", err)
		}
		ledgerTransactionEventRecord.AccountRoutingNumber = internalPayload.CreditorAccount.InstitutionId
		ledgerTransactionEventRecord.ExternalBankAccountNumber = encryptedBankAccountNumber
		ledgerTransactionEventRecord.ExternalBankAccountName = internalPayload.DebtorAccount.HolderName
		ledgerTransactionEventRecord.ExternalBankAccountRoutingNumber = internalPayload.DebtorAccount.InstitutionId
		ledgerTransactionEventRecord.AccountNumber = internalPayload.CreditorAccount.AccountNumber

		return &ledgerTransactionEventRecord

	case "PRE_AUTH":
		ledgerTransactionEventRecord.IsOutward = true
		ledgerTransactionEventRecord.AccountNumber = internalPayload.DebtorAccount.AccountNumber
		ledgerTransactionEventRecord.AccountRoutingNumber = internalPayload.DebtorAccount.InstitutionId

		return &ledgerTransactionEventRecord

	case "COMPLETION", "WITHDRAWAL", "PURCHASE", "BILLPAY_DEBIT":
		ledgerTransactionEventRecord.IsOutward = true

		ledgerTransactionEventRecord.CardPayeeId = internalPayload.CreditorAccount.AccountNumber
		ledgerTransactionEventRecord.CardPayeeName = internalPayload.CreditorAccount.HolderName

		ledgerTransactionEventRecord.AccountNumber = internalPayload.DebtorAccount.AccountNumber
		ledgerTransactionEventRecord.AccountRoutingNumber = internalPayload.DebtorAccount.InstitutionId

		return &ledgerTransactionEventRecord

	case "BILLPAY_CREDIT", "ATM_DEPOSIT", "RETURN":
		ledgerTransactionEventRecord.IsOutward = false

		ledgerTransactionEventRecord.CardPayeeId = internalPayload.CreditorAccount.AccountNumber
		ledgerTransactionEventRecord.CardPayeeName = internalPayload.CreditorAccount.HolderName

		ledgerTransactionEventRecord.AccountNumber = internalPayload.DebtorAccount.AccountNumber
		ledgerTransactionEventRecord.AccountRoutingNumber = internalPayload.DebtorAccount.InstitutionId

		return &ledgerTransactionEventRecord

	default:
		logging.Logger.Warn("ledger event is not of handled transcation type", "transactionType", internalPayload.TransactionType)
		return nil

		// TODO: Need to handle ATM deposit/withdrawal, cashapp in/out, and direct deposit when examples are available
	}
}

func SaveTransactionEventRecord(ledgerTransactionEventRecord dao.LedgerTransactionEventDao) error {
	accountRecord, err := dao.UserAccountCardDao{}.FindOneByAccountNumber(db.DB, ledgerTransactionEventRecord.AccountNumber)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to find user account record by account number: %w", err))
	}
	if accountRecord == nil {
		return errtrace.Wrap(fmt.Errorf("failed to find user account record by account number"))
	}

	ledgerTransactionEventRecord.UserId = accountRecord.UserId
	ledgerTransactionEventRecord.AccountNumber = accountRecord.AccountNumber
	encryptedPayload, err := utils.EncryptKmsBinary(string(ledgerTransactionEventRecord.RawPayload))
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to encrypt raw payload: %w", err))
	}
	ledgerTransactionEventRecord.RawPayload = encryptedPayload

	err = db.DB.Create(&ledgerTransactionEventRecord).Error
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("unable to save ledger transaction event record for ach out: %w", err))
	}

	return nil
}

func SendTransactionEventToSardine(ledgerTransactionEventRecord dao.LedgerTransactionEventDao) error {
	client, err := utils.NewSardineClient(config.Config.Sardine)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create sardine client: %w", err))
	}

	requestBody, err := MapLedgerEventRecordToSardineRequest(ledgerTransactionEventRecord)
	if err != nil {
		logging.Logger.Error("Error occurred while mapping ledger event record to sardine request", "err", err)
	}

	if requestBody == nil {
		logging.Logger.Info("Skipping empty SardineRequest", "eventId", ledgerTransactionEventRecord.EventId, "transactionType", ledgerTransactionEventRecord.TransactionType)
		return nil
	}

	sardineResponse, err := client.PostCustomerInformationWithResponse(context.Background(), *requestBody)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("error occurred while calling sardine API: %w", err))
	}

	switch {
	case sardineResponse.JSON200 != nil:
		var res bytes.Buffer
		if err := json.Indent(&res, sardineResponse.Body, "", "  "); err != nil {
			return errtrace.Wrap(fmt.Errorf("error occurred while formatting sardine response body: %w", err))
		}
		logging.Logger.Debug("Received 200 status code from Sardine", "successResponse", res.String())

	case sardineResponse.JSON400 != nil:
		return errtrace.Wrap(fmt.Errorf("received 400 response from sardine: %s", *sardineResponse.JSON400.Message))
	case sardineResponse.JSON401 != nil:
		return errtrace.Wrap(fmt.Errorf("received 401 response from sardine: %s", *sardineResponse.JSON401.Reason))
	case sardineResponse.JSON422 != nil:
		return errtrace.Wrap(fmt.Errorf("received 422 response from sardine: %s", *sardineResponse.JSON422.Message))
	default:
		return errtrace.Wrap(fmt.Errorf("received unexpected error response from sardine API"))
	}

	return nil
}

func getBankIdHash(routingNumber, accountNumber string) (*string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(routingNumber + accountNumber))
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	result := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	return &result, nil
}

func MapLedgerEventRecordToSardineRequest(ledgerTransactionEventRecord dao.LedgerTransactionEventDao) (*sardine.PostCustomerInformationJSONRequestBody, error) {
	if ledgerTransactionEventRecord.TransactionType == "PRE_AUTH" && ledgerTransactionEventRecord.CardPayeeId == "" {
		logging.Logger.Warn("skipping PRE_AUTH transaction. webhook payload does not contain CardPayeeId/Creditor AccountNumber")
		return nil, nil
	}
	return mapLedgerEventRecordToSardineRequestWithCardFetcher(ledgerTransactionEventRecord, GetCardDetailsForSardine)
}

func mapLedgerEventRecordToSardineRequestWithCardFetcher(
	ledgerTransactionEventRecord dao.LedgerTransactionEventDao,
	fetchCard func(dao.LedgerTransactionEventDao) (*ledger.GetCardDetailsResult, error),
) (*sardine.PostCustomerInformationJSONRequestBody, error) {
	switch ledgerTransactionEventRecord.TransactionType {
	case "ACH_PULL":
		return createSardineAchRequest(
			ledgerTransactionEventRecord,
			[]sardine.PostCustomerInformationJSONBodyCheckpoints{
				"customer", "ach", "counterparty", "aml",
			},
			"ach-inward-transfer",
		)
	case "ACH_OUT":
		return createSardineAchRequest(
			ledgerTransactionEventRecord,
			[]sardine.PostCustomerInformationJSONBodyCheckpoints{
				"customer", "ach", "counterparty", "aml",
			},
			"ach-outward-transfer",
		)
	case "PRE_AUTH", "COMPLETION", "PURCHASE", "BILLPAY_DEBIT":
		return createSardineCardPaymentRequest(
			ledgerTransactionEventRecord,
			[]sardine.PostCustomerInformationJSONBodyCheckpoints{
				"customer", "payment", "aml",
			},
			"debit-card-merchant",
			sardine.TransactionActionTypeBuy,
			fetchCard,
		)
	case "BILLPAY_CREDIT", "RETURN":
		return createSardineCardPaymentRequest(
			ledgerTransactionEventRecord,
			[]sardine.PostCustomerInformationJSONBodyCheckpoints{
				"customer", "return", "aml",
			},
			"debit-card-merchant",
			sardine.TransactionActionTypeRefund,
			fetchCard,
		)
	case "WITHDRAWAL":
		return createSardineAtmRequest(
			ledgerTransactionEventRecord,
			[]sardine.PostCustomerInformationJSONBodyCheckpoints{
				"customer", "atm", "aml",
			},
			"atm-withdrawal",
			sardine.TransactionActionTypeWithdraw,
		)
	case "ATM_DEPOSIT":
		return createSardineAtmRequest(
			ledgerTransactionEventRecord,
			[]sardine.PostCustomerInformationJSONBodyCheckpoints{
				"customer", "atm", "aml",
			},
			"atm-deposit",
			sardine.TransactionActionTypeDeposit,
		)
	default:
		return nil, fmt.Errorf("unrecognized or unhandled transaction type when formatting sardine request")
	}
}

func createSardineRequest(ledgerTransactionEventRecord dao.LedgerTransactionEventDao) sardine.PostCustomerInformationJSONRequestBody {
	sardineSessionKey := uuid.New().String()
	deviceFalse := false
	requestBody := sardine.PostCustomerInformationJSONRequestBody{
		SessionKey: sardineSessionKey,
		Flow:       sardine.Flow{},
		Config: &sardine.Config{
			Device: &deviceFalse,
		},
		Customer: sardine.Customer{
			Id: ledgerTransactionEventRecord.UserId,
		},
		Counterparty: &SardineCounterparty{},
	}
	return requestBody
}

func createSardineAtmRequest(ledgerTransactionEventRecord dao.LedgerTransactionEventDao, checkpoints []sardine.PostCustomerInformationJSONBodyCheckpoints, debitPreAuthFlowName string, actionType sardine.TransactionActionType) (*sardine.PostCustomerInformationJSONRequestBody, error) {
	requestBody := createSardineRequest(ledgerTransactionEventRecord)
	amount := float32(utils.CentsToUSD(int64(ledgerTransactionEventRecord.InstructedAmount)))

	internalBank, err := createSardineInternalBankPaymentMethod(ledgerTransactionEventRecord)
	if err != nil {
		return nil, fmt.Errorf("error createing SardineInternalBank: %w", err)
	}
	requestBody.Checkpoints = &checkpoints
	requestBody.Flow.Name = &debitPreAuthFlowName
	requestBody.Transaction = &sardine.Transaction{
		Id:              ledgerTransactionEventRecord.TransactionNumber,
		ActionType:      &actionType,
		IsOutward:       &ledgerTransactionEventRecord.IsOutward,
		Amount:          &amount,
		CreatedAtMillis: utils.Pointer(ledgerTransactionEventRecord.CreatedAt.Unix()),
		CurrencyCode:    &ledgerTransactionEventRecord.InstructedCurrency,
		Mcc:             &ledgerTransactionEventRecord.Mcc,
		PaymentMethod:   internalBank,
	}
	return &requestBody, nil
}

func createSardineCardPaymentRequest(
	ledgerTransactionEventRecord dao.LedgerTransactionEventDao,
	checkpoints []sardine.PostCustomerInformationJSONBodyCheckpoints,
	debitPreAuthFlowName string,
	actionType sardine.TransactionActionType,
	fetchCard func(dao.LedgerTransactionEventDao) (*ledger.GetCardDetailsResult, error),
) (*sardine.PostCustomerInformationJSONRequestBody, error) {
	getCardResult, err := fetchCard(ledgerTransactionEventRecord)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to get card details from call to ledger: %w", err))
	}

	lastFour := getCardResult.Card.CardMaskNumber[len(getCardResult.Card.CardMaskNumber)-4:]

	amount := float32(utils.CentsToUSD(int64(ledgerTransactionEventRecord.InstructedAmount)))

	requestBody := createSardineRequest(ledgerTransactionEventRecord)

	requestBody.Checkpoints = &checkpoints
	requestBody.Flow.Name = &debitPreAuthFlowName
	requestBody.Transaction = &sardine.Transaction{
		Id:              ledgerTransactionEventRecord.TransactionNumber,
		ActionType:      &actionType,
		IsOutward:       &ledgerTransactionEventRecord.IsOutward,
		Amount:          &amount,
		CreatedAtMillis: utils.Pointer(ledgerTransactionEventRecord.CreatedAt.Unix()),
		CurrencyCode:    &ledgerTransactionEventRecord.InstructedCurrency,
		Mcc:             &ledgerTransactionEventRecord.Mcc,
		PaymentMethod: &sardine.PaymentMethod{
			Type: "card",
			Card: &sardine.Card{
				Last4: &lastFour,
				Hash:  &ledgerTransactionEventRecord.CardId,
				Bin:   &ledgerTransactionEventRecord.BinNumber,
			},
		},
	}

	requestBody.Counterparty = &SardineCounterparty{}
	requestBody.Counterparty.Type = "vendor"
	if ledgerTransactionEventRecord.CardPayeeId != "" {
		requestBody.Counterparty.Id = ledgerTransactionEventRecord.CardPayeeId
	}
	if ledgerTransactionEventRecord.CardPayeeName != "" {
		requestBody.Counterparty.BusinessName = &ledgerTransactionEventRecord.CardPayeeName
	}
	return &requestBody, nil
}

func createSardineAchRequest(ledgerTransactionEventRecord dao.LedgerTransactionEventDao, checkpoints []sardine.PostCustomerInformationJSONBodyCheckpoints, achOutFlowName string) (*sardine.PostCustomerInformationJSONRequestBody, error) {
	requestBody := createSardineRequest(ledgerTransactionEventRecord)
	amount := float32(utils.CentsToUSD(int64(ledgerTransactionEventRecord.InstructedAmount)))

	externalBankAccountName := ledgerTransactionEventRecord.ExternalBankAccountName
	externalBankAccountRoutingNumber := ledgerTransactionEventRecord.ExternalBankAccountRoutingNumber

	internalBankPaymentMethod, err := createSardineInternalBankPaymentMethod(ledgerTransactionEventRecord)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("error calling createSardineInternalBankPaymentMethod: %w", err))
	}
	externalBankPaymentMethod, err := createSardineExternalBankPaymentMethod(ledgerTransactionEventRecord)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("error calling createSardineExternalBankPaymentMethod: %w", err))
	}

	requestBody.Checkpoints = &checkpoints
	requestBody.Flow.Name = &achOutFlowName
	requestBody.Transaction = &sardine.Transaction{
		Id:              ledgerTransactionEventRecord.TransactionNumber,
		ActionType:      utils.Pointer(sardine.TransactionActionTypeTransfer),
		IsOutward:       &ledgerTransactionEventRecord.IsOutward,
		Amount:          &amount,
		CreatedAtMillis: utils.Pointer(ledgerTransactionEventRecord.CreatedAt.Unix()),
		CurrencyCode:    &ledgerTransactionEventRecord.InstructedCurrency,
		PaymentMethod:   internalBankPaymentMethod,
	}
	requestBody.Counterparty = &SardineCounterparty{
		Id:            externalBankAccountRoutingNumber,
		FirstName:     &externalBankAccountName,
		MiddleName:    &externalBankAccountName,
		LastName:      &externalBankAccountName,
		PaymentMethod: externalBankPaymentMethod,
	}
	return &requestBody, nil
}

func createSardineExternalBankPaymentMethod(ledgerTransactionEventRecord dao.LedgerTransactionEventDao) (*sardine.PaymentMethod, error) {
	var externalBankAccountNumber string
	if len(ledgerTransactionEventRecord.ExternalBankAccountNumber) > 0 {
		var err error
		externalBankAccountNumber, err = utils.DecryptKmsBinary(ledgerTransactionEventRecord.ExternalBankAccountNumber)
		if err != nil {
			return nil, errtrace.Wrap(fmt.Errorf("error occurred while decrypting external bank account number: %w", err))
		}
	}

	externalBankAccountRoutingNumber := ledgerTransactionEventRecord.ExternalBankAccountRoutingNumber
	externalBankId, err := getBankIdHash(externalBankAccountRoutingNumber, externalBankAccountNumber)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("error occurred while hashing user bank account id: %w", err))
	}
	return &sardine.PaymentMethod{
		Type: sardine.PaymentMethodTypeBank,
		Bank: &sardine.Bank{
			Id:                  externalBankId,
			IdSource:            utils.Pointer("internal"),
			AccountNumber:       &externalBankAccountNumber,
			RoutingNumber:       &externalBankAccountRoutingNumber,
			BalanceCurrencyCode: utils.Pointer("USD"),
		},
	}, nil
}

func createSardineInternalBankPaymentMethod(ledgerTransactionEventRecord dao.LedgerTransactionEventDao) (*sardine.PaymentMethod, error) {
	userAccountRoutingNumber := ledgerTransactionEventRecord.AccountRoutingNumber
	userAccountNumber := ledgerTransactionEventRecord.AccountNumber
	userBankId, err := getBankIdHash(userAccountRoutingNumber, userAccountNumber)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("error occurred while hashing user bank account id: %w", err))
	}
	return &sardine.PaymentMethod{
		Type: sardine.PaymentMethodTypeBank,
		Bank: &sardine.Bank{
			Id:                  userBankId,
			IdSource:            utils.Pointer("internal"),
			AccountNumber:       &userAccountNumber,
			RoutingNumber:       &userAccountRoutingNumber,
			BalanceCurrencyCode: utils.Pointer("USD"),
		},
	}, nil
}

func GetCardDetailsForSardine(ledgerTransactionEventRecord dao.LedgerTransactionEventDao) (*ledger.GetCardDetailsResult, error) {
	userAccountCardRecord, err := dao.UserAccountCardDao{}.FindOneByAccountNumber(db.DB, ledgerTransactionEventRecord.AccountNumber)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to get user account card record for retrieving card details: %w", err))
	}
	if userAccountCardRecord == nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to get user account card record for retrieving card details"))
	}

	userRecord, errResponse := dao.RequireUserWithState(userAccountCardRecord.UserId, constant.ACTIVE)
	if errResponse != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to get user record for retrieving card details: %w", err))
	}

	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerSignedClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, ledgerParamsBuilder)

	getCardRequest := ledgerSignedClient.BuildGetCardDetailsRequest(userRecord.LedgerCustomerNumber, userAccountCardRecord.AccountNumber, userAccountCardRecord.CardId)
	getCardResponse, err := ledgerSignedClient.GetCardDetails(getCardRequest)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("received error from callLedgerGetCardDetails: %w", err))
	}

	if getCardResponse.Error != nil {
		return nil, errtrace.Wrap(fmt.Errorf("received error in getCardResponse: %w", err))
	}

	if getCardResponse.Result == nil {
		logging.Logger.Error("The ledger responded with an empty result object", "response", getCardResponse)
		return nil, errtrace.Wrap(fmt.Errorf("getCardResponse result object is unexpectedly nil: %w", err))
	}

	return getCardResponse.Result, nil
}

type TransactionMonitoringArgs struct {
	EventId string `json:"eventId"`
}

func (TransactionMonitoringArgs) Kind() string { return "transactionMonitoring" }

type TransactionMonitoringWorker struct {
	river.WorkerDefaults[TransactionMonitoringArgs]
}

func RegisterTransactionMonitoringWorker(workers *river.Workers) {
	river.AddWorker(workers, &TransactionMonitoringWorker{})
}

func (w *TransactionMonitoringWorker) Work(ctx context.Context, job *river.Job[TransactionMonitoringArgs]) error {
	ledgerTransactionEventRecord, err := dao.LedgerTransactionEventDao{}.FindOneByEventId(db.DB, job.Args.EventId)
	if err != nil {
		logging.Logger.Warn("Unable to find ledger event record for job", "err", err)
		return err
	}
	if ledgerTransactionEventRecord == nil {
		logging.Logger.Warn("Unable to find ledger event record for job")
		return fmt.Errorf("failed to find ledger event record for job")
	}

	if !config.Config.Sardine.SendTransactions {
		logging.Logger.Warn("SendTransactions is disabled for Sardine. Skipping delivery")
		return nil
	}

	err = SendTransactionEventToSardine(*ledgerTransactionEventRecord)
	if err != nil {
		logging.Logger.Warn("Received error calling Sardine from job", "err", err.Error())
	}

	return nil
}

func notifyUserOfStatement(notificationEmailArgs StatementNotificationEmailJobArgs) error {
	appLink := fmt.Sprintf("%s/web/account/%s/statements", notificationEmailArgs.BaseUrl, notificationEmailArgs.AccountId)
	emailData := response.StatementEmailTemplateData{
		FirstName: notificationEmailArgs.FirstName,
		Month:     notificationEmailArgs.Month,
		Year:      notificationEmailArgs.Year,
		AppLink:   appLink,
	}

	templateName := "../email-templates/statementNotificationTemplate.html"
	htmlBody, err := utils.GenerateEmailBody(templateName, emailData)
	if err != nil {
		return errtrace.Wrap(err)
	}

	emailSubject := "Your Monthly DreamFi Statement is Ready"
	err = utils.SendEmail(notificationEmailArgs.FirstName, notificationEmailArgs.Email, emailSubject, htmlBody)
	if err != nil {
		return errtrace.Wrap(err)
	}

	return nil
}

type StatementNotificationEmailEnqueueBatchJobArgs struct {
	AccountIds []string `json:"accountIds"`
	Month      string   `json:"month"`
	Year       string   `json:"year"`
	BaseUrl    string   `json:"baseUrl"`
}

func (StatementNotificationEmailEnqueueBatchJobArgs) Kind() string {
	return "statement_notification_email_enqueue_batch_job"
}

func (StatementNotificationEmailEnqueueBatchJobArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: "sendgrid",
	}
}

type StatementNotificationEmailEnqueueBatchWorker struct {
	river.WorkerDefaults[StatementNotificationEmailEnqueueBatchJobArgs]
	RiverClient *river.Client[*sql.Tx]
}

func (w *StatementNotificationEmailEnqueueBatchWorker) SetRiverClientForBatchWorker(client *river.Client[*sql.Tx]) {
	w.RiverClient = client
}

func RegisterStatementNotificationEmailEnqueueBatchWorker(workers *river.Workers, riverClient *river.Client[*sql.Tx]) *StatementNotificationEmailEnqueueBatchWorker {
	worker := &StatementNotificationEmailEnqueueBatchWorker{
		RiverClient: riverClient,
	}
	river.AddWorker(workers, worker)
	return worker
}

func (w *StatementNotificationEmailEnqueueBatchWorker) Work(ctx context.Context, job *river.Job[StatementNotificationEmailEnqueueBatchJobArgs]) error {
	var batchParams []river.InsertManyParams
	cardsWithUsers, err := dao.UserAccountCardDao{}.FindUserAccountCardWithUserByAccountIds(job.Args.AccountIds)
	if err != nil {
		logging.Logger.Error("Failed to lookup users/accounts for statement notification batch", "accountIds", job.Args.AccountIds, "error", err.Error())
		return err
	}
	if cardsWithUsers == nil {
		logging.Logger.Warn("No users/accounts found for statement notification batch", "accountIds", job.Args.AccountIds)
		return fmt.Errorf("no users/accounts found for statement notification batch")
	}

	for _, cardWithUser := range *cardsWithUsers {
		batchParams = append(batchParams, river.InsertManyParams{
			Args: StatementNotificationEmailJobArgs{
				FirstName: cardWithUser.FirstName,
				Email:     cardWithUser.Email,
				Month:     job.Args.Month,
				Year:      job.Args.Year,
				BaseUrl:   job.Args.BaseUrl,
				AccountId: cardWithUser.AccountId,
			},
		})
	}

	if len(batchParams) == 0 {
		logging.Logger.Error("Empty batch params for statement notification email worker")
		return nil
	}

	_, err = w.RiverClient.InsertMany(ctx, batchParams)
	if err != nil {
		logging.Logger.Error("Failed to enqueue statement notification jobs", "error", err.Error())
		return err
	}

	return nil
}

type StatementNotificationEmailJobArgs struct {
	FirstName string `json:"firstName"`
	Email     string `json:"email"`
	Month     string `json:"month"`
	Year      string `json:"year"`
	BaseUrl   string `json:"baseUrl"`
	AccountId string `json:"accountId"`
}

func (StatementNotificationEmailJobArgs) Kind() string { return "statement_notification_email" }

func (StatementNotificationEmailJobArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: "sendgrid",
	}
}

type StatementNotificationWorker struct {
	river.WorkerDefaults[StatementNotificationEmailJobArgs]
}

func RegisterStatementNotificationWorker(workers *river.Workers) {
	river.AddWorker(workers, &StatementNotificationWorker{})
}

func (w *StatementNotificationWorker) Work(ctx context.Context, job *river.Job[StatementNotificationEmailJobArgs]) error {
	err := notifyUserOfStatement(job.Args)
	if err != nil {
		logging.Logger.Error("Error sending new statement notification email", "err", err)
		return err
	}
	return nil
}
