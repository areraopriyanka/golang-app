package handler

import (
	"encoding/base64"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/sardine"
	"process-api/pkg/utils"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockKMSBinaryClient struct {
	kmsiface.KMSAPI
}

func (mockKMSBinaryClient) Decrypt(input *kms.DecryptInput) (*kms.DecryptOutput, error) {
	inputToDecode := input.CiphertextBlob
	result, err := base64.StdEncoding.DecodeString(string(inputToDecode))
	if err != nil {
		return nil, err
	}
	return &kms.DecryptOutput{Plaintext: result}, nil
}

func (mockKMSBinaryClient) Encrypt(input *kms.EncryptInput) (*kms.EncryptOutput, error) {
	result := base64.StdEncoding.EncodeToString(input.Plaintext)
	return &kms.EncryptOutput{
		CiphertextBlob: []byte(result),
	}, nil
}

func TestMapLedgerEventRecordToSardineRequest_ACH_PULL(t *testing.T) {
	utils.SetKmsClient(mockKMSBinaryClient{})

	externalBankAccountNumber, err := utils.EncryptKmsBinary("987546218371925")
	require.NoError(t, err)

	record := dao.LedgerTransactionEventDao{
		EventId:                          "EVT62669",
		UserId:                           "test-user-123",
		TransactionType:                  "ACH_PULL",
		TransactionNumber:                "QA00000000048624",
		AccountNumber:                    "500400039328683",
		AccountRoutingNumber:             "124303298",
		ExternalBankAccountNumber:        externalBankAccountNumber,
		ExternalBankAccountName:          "DONOR_FIRSTNAME",
		ExternalBankAccountRoutingNumber: "011002550",
		InstructedAmount:                 1300,
		InstructedCurrency:               "USD",
		IsOutward:                        false,
		CreatedAt:                        time.Unix(1735171039, 0),
	}

	stubFetcher := func(dao.LedgerTransactionEventDao) (*ledger.GetCardDetailsResult, error) {
		return nil, nil
	}

	result, err := mapLedgerEventRecordToSardineRequestWithCardFetcher(record, stubFetcher)
	require.NoError(t, err)
	require.NotNil(t, result)

	userBankId, err := getBankIdHash("124303298", "500400039328683")
	require.NoError(t, err)
	externalBankId, err := getBankIdHash("011002550", "987546218371925")
	require.NoError(t, err)

	deviceFalse := false
	transferAction := sardine.TransactionActionType("transfer")
	internalIdSource := "internal"
	usdCurrency := "USD"
	achInFlow := "ach-inward-transfer"
	amount := float32(13.00)
	createdAt := int64(1735171039)
	isOutward := false
	externalName := "DONOR_FIRSTNAME"
	userAccountNum := "500400039328683"
	userRoutingNum := "124303298"
	externalAccountNum := "987546218371925"
	externalRoutingNum := "011002550"

	expected := &sardine.PostCustomerInformationJSONRequestBody{
		SessionKey: result.SessionKey,
		Flow:       sardine.Flow{Name: &achInFlow},
		Config:     &sardine.Config{Device: &deviceFalse},
		Customer:   sardine.Customer{Id: "test-user-123"},
		Checkpoints: &[]sardine.PostCustomerInformationJSONBodyCheckpoints{
			"customer", "ach", "counterparty", "aml",
		},
		Transaction: &sardine.Transaction{
			Id:              "QA00000000048624",
			ActionType:      &transferAction,
			IsOutward:       &isOutward,
			Amount:          &amount,
			CurrencyCode:    &usdCurrency,
			CreatedAtMillis: &createdAt,
			PaymentMethod: &sardine.PaymentMethod{
				Type: "bank",
				Bank: &sardine.Bank{
					Id:                  userBankId,
					IdSource:            &internalIdSource,
					AccountNumber:       &userAccountNum,
					RoutingNumber:       &userRoutingNum,
					BalanceCurrencyCode: &usdCurrency,
				},
			},
		},
		Counterparty: &SardineCounterparty{
			Id:         "011002550",
			FirstName:  &externalName,
			MiddleName: &externalName,
			LastName:   &externalName,
			PaymentMethod: &sardine.PaymentMethod{
				Type: "bank",
				Bank: &sardine.Bank{
					Id:                  externalBankId,
					IdSource:            &internalIdSource,
					AccountNumber:       &externalAccountNum,
					RoutingNumber:       &externalRoutingNum,
					BalanceCurrencyCode: &usdCurrency,
				},
			},
		},
	}

	assert.Equal(t, expected, result)
}

func TestMapLedgerEventRecordToSardineRequest_ACH_OUT(t *testing.T) {
	utils.SetKmsClient(mockKMSBinaryClient{})

	externalBankAccountNumber, err := utils.EncryptKmsBinary("987546218371925")
	require.NoError(t, err)

	record := dao.LedgerTransactionEventDao{
		EventId:                          "EVT62670",
		UserId:                           "test-user-456",
		TransactionType:                  "ACH_OUT",
		TransactionNumber:                "QA00000000048625",
		AccountNumber:                    "500400039328683",
		AccountRoutingNumber:             "124303298",
		ExternalBankAccountNumber:        externalBankAccountNumber,
		ExternalBankAccountName:          "BENEFACTOR_FIRSTNAME",
		ExternalBankAccountRoutingNumber: "011002550",
		InstructedAmount:                 110,
		InstructedCurrency:               "USD",
		IsOutward:                        true,
		CreatedAt:                        time.Unix(1735171039, 0),
	}

	stubFetcher := func(dao.LedgerTransactionEventDao) (*ledger.GetCardDetailsResult, error) {
		return nil, nil
	}

	result, err := mapLedgerEventRecordToSardineRequestWithCardFetcher(record, stubFetcher)
	require.NoError(t, err)
	require.NotNil(t, result)

	userBankId, err := getBankIdHash("124303298", "500400039328683")
	require.NoError(t, err)
	externalBankId, err := getBankIdHash("011002550", "987546218371925")
	require.NoError(t, err)

	deviceFalse := false
	transferAction := sardine.TransactionActionType("transfer")
	internalIdSource := "internal"
	usdCurrency := "USD"
	achOutFlow := "ach-outward-transfer"
	amount := float32(1.10)
	createdAt := int64(1735171039)
	isOutward := true
	externalName := "BENEFACTOR_FIRSTNAME"
	userAccountNum := "500400039328683"
	userRoutingNum := "124303298"
	externalAccountNum := "987546218371925"
	externalRoutingNum := "011002550"

	expected := &sardine.PostCustomerInformationJSONRequestBody{
		SessionKey: result.SessionKey,
		Flow:       sardine.Flow{Name: &achOutFlow},
		Config:     &sardine.Config{Device: &deviceFalse},
		Customer:   sardine.Customer{Id: "test-user-456"},
		Checkpoints: &[]sardine.PostCustomerInformationJSONBodyCheckpoints{
			"customer", "ach", "counterparty", "aml",
		},
		Transaction: &sardine.Transaction{
			Id:              "QA00000000048625",
			ActionType:      &transferAction,
			IsOutward:       &isOutward,
			Amount:          &amount,
			CurrencyCode:    &usdCurrency,
			CreatedAtMillis: &createdAt,
			PaymentMethod: &sardine.PaymentMethod{
				Type: "bank",
				Bank: &sardine.Bank{
					Id:                  userBankId,
					IdSource:            &internalIdSource,
					AccountNumber:       &userAccountNum,
					RoutingNumber:       &userRoutingNum,
					BalanceCurrencyCode: &usdCurrency,
				},
			},
		},
		Counterparty: &SardineCounterparty{
			Id:         "011002550",
			FirstName:  &externalName,
			MiddleName: &externalName,
			LastName:   &externalName,
			PaymentMethod: &sardine.PaymentMethod{
				Type: "bank",
				Bank: &sardine.Bank{
					Id:                  externalBankId,
					IdSource:            &internalIdSource,
					AccountNumber:       &externalAccountNum,
					RoutingNumber:       &externalRoutingNum,
					BalanceCurrencyCode: &usdCurrency,
				},
			},
		},
	}

	assert.Equal(t, expected, result)
}

func TestMapLedgerEventRecordToSardineRequest_PRE_AUTH(t *testing.T) {
	utils.SetKmsClient(mockKMSBinaryClient{})

	record := dao.LedgerTransactionEventDao{
		EventId:              "EVT62671",
		UserId:               "test-user-789",
		TransactionType:      "PRE_AUTH",
		TransactionNumber:    "QA00000000048626",
		AccountNumber:        "500400039328683",
		AccountRoutingNumber: "124303298",
		CardId:               "9be76d9f293844f38ca2de5a6c26918a",
		BinNumber:            "446614422",
		InstructedAmount:     1272,
		InstructedCurrency:   "USD",
		Mcc:                  "5965",
		IsOutward:            true,
		CreatedAt:            time.Unix(1735171039, 0),
	}

	stubFetcher := func(dao.LedgerTransactionEventDao) (*ledger.GetCardDetailsResult, error) {
		result := &ledger.GetCardDetailsResult{}
		result.Card.CardMaskNumber = "************7324"
		return result, nil
	}

	result, err := mapLedgerEventRecordToSardineRequestWithCardFetcher(record, stubFetcher)
	require.NoError(t, err)
	require.NotNil(t, result)

	deviceFalse := false
	buyAction := sardine.TransactionActionType("buy")
	usdCurrency := "USD"
	debitFlow := "debit-card-merchant"
	amount := float32(12.72)
	createdAt := int64(1735171039)
	isOutward := true
	mcc := "5965"
	lastFour := "7324"
	cardHash := "9be76d9f293844f38ca2de5a6c26918a"
	bin := "446614422"

	expected := &sardine.PostCustomerInformationJSONRequestBody{
		SessionKey: result.SessionKey,
		Flow:       sardine.Flow{Name: &debitFlow},
		Config:     &sardine.Config{Device: &deviceFalse},
		Customer:   sardine.Customer{Id: "test-user-789"},
		Checkpoints: &[]sardine.PostCustomerInformationJSONBodyCheckpoints{
			"customer", "payment", "aml",
		},
		Transaction: &sardine.Transaction{
			Id:              "QA00000000048626",
			ActionType:      &buyAction,
			IsOutward:       &isOutward,
			Amount:          &amount,
			CurrencyCode:    &usdCurrency,
			Mcc:             &mcc,
			CreatedAtMillis: &createdAt,
			PaymentMethod: &sardine.PaymentMethod{
				Type: "card",
				Card: &sardine.Card{
					Last4: &lastFour,
					Hash:  &cardHash,
					Bin:   &bin,
				},
			},
		},
		Counterparty: &SardineCounterparty{
			Id:   "",
			Type: "vendor",
		},
	}

	assert.Equal(t, expected, result)
}

func TestMapLedgerEventRecordToSardineRequest_COMPLETION(t *testing.T) {
	utils.SetKmsClient(mockKMSBinaryClient{})

	record := dao.LedgerTransactionEventDao{
		EventId:              "EVT62672",
		UserId:               "test-user-999",
		TransactionType:      "COMPLETION",
		TransactionNumber:    "QA00000000048627",
		AccountNumber:        "100010746502013",
		AccountRoutingNumber: "124303298",
		CardId:               "9be76d9f293844f38ca2de5a6c26918a",
		CardPayeeId:          "100010746502013",
		CardPayeeName:        "ACQUIRER NAME          CITY NAME      US",
		BinNumber:            "446614422",
		InstructedAmount:     1347,
		InstructedCurrency:   "USD",
		Mcc:                  "5969",
		IsOutward:            true,
		CreatedAt:            time.Unix(1735171039, 0),
	}

	stubFetcher := func(dao.LedgerTransactionEventDao) (*ledger.GetCardDetailsResult, error) {
		result := &ledger.GetCardDetailsResult{}
		result.Card.CardMaskNumber = "************7324"
		return result, nil
	}

	result, err := mapLedgerEventRecordToSardineRequestWithCardFetcher(record, stubFetcher)
	require.NoError(t, err)
	require.NotNil(t, result)

	deviceFalse := false
	buyAction := sardine.TransactionActionType("buy")
	usdCurrency := "USD"
	debitFlow := "debit-card-merchant"
	amount := float32(13.47)
	createdAt := int64(1735171039)
	isOutward := true
	mcc := "5969"
	lastFour := "7324"
	cardHash := "9be76d9f293844f38ca2de5a6c26918a"
	bin := "446614422"
	businessName := "ACQUIRER NAME          CITY NAME      US"

	expected := &sardine.PostCustomerInformationJSONRequestBody{
		SessionKey: result.SessionKey,
		Flow:       sardine.Flow{Name: &debitFlow},
		Config:     &sardine.Config{Device: &deviceFalse},
		Customer:   sardine.Customer{Id: "test-user-999"},
		Checkpoints: &[]sardine.PostCustomerInformationJSONBodyCheckpoints{
			"customer", "payment", "aml",
		},
		Transaction: &sardine.Transaction{
			Id:              "QA00000000048627",
			ActionType:      &buyAction,
			IsOutward:       &isOutward,
			Amount:          &amount,
			CurrencyCode:    &usdCurrency,
			Mcc:             &mcc,
			CreatedAtMillis: &createdAt,
			PaymentMethod: &sardine.PaymentMethod{
				Type: "card",
				Card: &sardine.Card{
					Last4: &lastFour,
					Hash:  &cardHash,
					Bin:   &bin,
				},
			},
		},
		Counterparty: &SardineCounterparty{
			Id:           "100010746502013",
			Type:         "vendor",
			BusinessName: &businessName,
		},
	}

	assert.Equal(t, expected, result)
}

func TestMapLedgerEventRecordToSardineRequest_WITHDRAWAL(t *testing.T) {
	utils.SetKmsClient(mockKMSBinaryClient{})

	record := dao.LedgerTransactionEventDao{
		EventId:              "EVT62673",
		UserId:               "test-user-111",
		TransactionType:      "WITHDRAWAL",
		TransactionNumber:    "QA00000000048628",
		AccountNumber:        "500400039328683",
		AccountRoutingNumber: "124303298",
		InstructedAmount:     5000,
		InstructedCurrency:   "USD",
		IsOutward:            true,
		CreatedAt:            time.Unix(1735171039, 0),
	}

	stubFetcher := func(dao.LedgerTransactionEventDao) (*ledger.GetCardDetailsResult, error) {
		return nil, nil
	}

	result, err := mapLedgerEventRecordToSardineRequestWithCardFetcher(record, stubFetcher)
	require.NoError(t, err)
	require.NotNil(t, result)

	userBankId, err := getBankIdHash("124303298", "500400039328683")
	require.NoError(t, err)

	deviceFalse := false
	withdrawAction := sardine.TransactionActionType("withdraw")
	internalIdSource := "internal"
	usdCurrency := "USD"
	atmWithdrawalFlow := "atm-withdrawal"
	amount := float32(50.00)
	createdAt := int64(1735171039)
	isOutward := true
	userAccountNum := "500400039328683"
	userRoutingNum := "124303298"
	mcc := ""

	expected := &sardine.PostCustomerInformationJSONRequestBody{
		SessionKey: result.SessionKey,
		Flow:       sardine.Flow{Name: &atmWithdrawalFlow},
		Config:     &sardine.Config{Device: &deviceFalse},
		Customer:   sardine.Customer{Id: "test-user-111"},
		Checkpoints: &[]sardine.PostCustomerInformationJSONBodyCheckpoints{
			"customer", "atm", "aml",
		},
		Transaction: &sardine.Transaction{
			Id:              "QA00000000048628",
			ActionType:      &withdrawAction,
			IsOutward:       &isOutward,
			Amount:          &amount,
			CurrencyCode:    &usdCurrency,
			Mcc:             &mcc,
			CreatedAtMillis: &createdAt,
			PaymentMethod: &sardine.PaymentMethod{
				Type: "bank",
				Bank: &sardine.Bank{
					Id:                  userBankId,
					IdSource:            &internalIdSource,
					AccountNumber:       &userAccountNum,
					RoutingNumber:       &userRoutingNum,
					BalanceCurrencyCode: &usdCurrency,
				},
			},
		},
		Counterparty: &SardineCounterparty{},
	}

	assert.Equal(t, expected, result)
}

func TestMapLedgerEventRecordToSardineRequest_ATM_DEPOSIT(t *testing.T) {
	utils.SetKmsClient(mockKMSBinaryClient{})

	record := dao.LedgerTransactionEventDao{
		EventId:              "EVT62674",
		UserId:               "test-user-222",
		TransactionType:      "ATM_DEPOSIT",
		TransactionNumber:    "QA00000000048629",
		AccountNumber:        "500400039328683",
		AccountRoutingNumber: "124303298",
		InstructedAmount:     10000,
		InstructedCurrency:   "USD",
		IsOutward:            false,
		CreatedAt:            time.Unix(1735171039, 0),
	}

	stubFetcher := func(dao.LedgerTransactionEventDao) (*ledger.GetCardDetailsResult, error) {
		return nil, nil
	}

	result, err := mapLedgerEventRecordToSardineRequestWithCardFetcher(record, stubFetcher)
	require.NoError(t, err)
	require.NotNil(t, result)

	userBankId, err := getBankIdHash("124303298", "500400039328683")
	require.NoError(t, err)

	deviceFalse := false
	depositAction := sardine.TransactionActionType("deposit")
	internalIdSource := "internal"
	usdCurrency := "USD"
	atmDepositFlow := "atm-deposit"
	amount := float32(100.00)
	createdAt := int64(1735171039)
	isOutward := false
	userAccountNum := "500400039328683"
	userRoutingNum := "124303298"
	mcc := ""

	expected := &sardine.PostCustomerInformationJSONRequestBody{
		SessionKey: result.SessionKey,
		Flow:       sardine.Flow{Name: &atmDepositFlow},
		Config:     &sardine.Config{Device: &deviceFalse},
		Customer:   sardine.Customer{Id: "test-user-222"},
		Checkpoints: &[]sardine.PostCustomerInformationJSONBodyCheckpoints{
			"customer", "atm", "aml",
		},
		Transaction: &sardine.Transaction{
			Id:              "QA00000000048629",
			ActionType:      &depositAction,
			IsOutward:       &isOutward,
			Amount:          &amount,
			CurrencyCode:    &usdCurrency,
			Mcc:             &mcc,
			CreatedAtMillis: &createdAt,
			PaymentMethod: &sardine.PaymentMethod{
				Type: "bank",
				Bank: &sardine.Bank{
					Id:                  userBankId,
					IdSource:            &internalIdSource,
					AccountNumber:       &userAccountNum,
					RoutingNumber:       &userRoutingNum,
					BalanceCurrencyCode: &usdCurrency,
				},
			},
		},
		Counterparty: &SardineCounterparty{},
	}

	assert.Equal(t, expected, result)
}
