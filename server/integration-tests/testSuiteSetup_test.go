package test

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"log/slog"
	"os"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/handler"
	"process-api/pkg/logging"
	"process-api/pkg/plaid"
	"process-api/pkg/resource/agreements"
	"process-api/pkg/utils"
	"process-api/pkg/validators"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	initialDB   *gorm.DB
	TestDB      *gorm.DB
	riverClient *river.Client[*sql.Tx]
}

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

func setupAgreements() {
	agreements.Agreements = &agreements.AgreementStore{
		ESign: agreements.AgreementStoreEntry{
			Name: "e-sign",
			JSON: "{}",
			Hash: "e-sign-hash",
		},
		DreamfiAchAuthorization: agreements.AgreementStoreEntry{
			Name: "dreamfi-ach-authorization",
			JSON: "{}",
			Hash: "dreamfi-ach-authorization-hash",
		},
		CardAndDeposit: agreements.AgreementStoreEntry{
			Name: "card-and-deposit",
			JSON: "{}",
			Hash: "card-and-deposit-hash",
		},
		PrivacyNotice: agreements.AgreementStoreEntry{
			Name: "privacy-notice",
			JSON: "{}",
			Hash: "privacy-notice-hash",
		},
		TermsOfService: agreements.AgreementStoreEntry{
			Name: "terms-of-service",
			JSON: "{}",
			Hash: "terms-of-service-hash",
		},
	}
}

func migrateTables(database *gorm.DB) {
	result := database.Exec(`
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
	`)
	if result.Error != nil {
		panic(result.Error)
	}

	err := db.Automigrate(database.DB(), "../")
	if err != nil {
		panic(err)
	}
}

func (suite *IntegrationTestSuite) SetupSuite() {
	err := validators.InitValidationRules()
	if err != nil {
		log.Fatalf("Error initiating validation rules: %v", err)
	}

	err = config.ReadConfig(nil)
	suite.NoError(err)

	logging.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	setupAgreements()

	initialDBConfig := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		"middleware", // username
		"asdf123",    // password
		"localhost",  // host
		"5433",       // port
		"middleware", // database name
		"disable",    // certificates
	)

	initialDB, err := gorm.Open("postgres", initialDBConfig)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	migrateTables(initialDB)

	suite.initialDB = initialDB

	err = utils.InitializePosthogClient(config.Config.Posthog)
	if err != nil {
		log.Fatalf("Error in initializing posthog client: %s", err.Error())
	}
}

func (suite *IntegrationTestSuite) BeforeTest(suiteName, testName string) {
	logging.Logger.Info("Running", "testName", testName)

	utils.SetKmsClient(mockKMSBinaryClient{})
	suite.TestDB = suite.initialDB.Begin()
	db.DB = suite.TestDB

	cfg := config.NewConfig()
	cfg.Plaid.Environment = "http://localhost:5007"
	suite.riverClient = suite.newRiver(cfg, testName)
}

// WARNING: calling suite.Error, suite.Fail, suite.FailNow in AfterTest will result in an infinite loop
func (suite *IntegrationTestSuite) AfterTest(suiteName, testName string) {
	suite.Require().NotNil(suite.riverClient, "river client should not be nil")
	ctx := context.Background()

	params := river.NewJobListParams().States(rivertype.JobStateAvailable, rivertype.JobStatePending, rivertype.JobStateRunning, rivertype.JobStateScheduled, rivertype.JobStateRetryable)
	jobRows, err := suite.riverClient.JobList(ctx, params)
	if err != nil {
		suite.T().Errorf("Failed to query remaining jobs: %s", err.Error())
	} else {
		if len(jobRows.Jobs) > 0 {
			suite.T().Errorf("test finished with running jobs. did you forget to run WaitForJobsDone in: %s", testName)
		} else {
			err := suite.riverClient.StopAndCancel(ctx)
			if err != nil {
				suite.T().Errorf("failed to river client: %s", err.Error())
			}

			logging.Logger.Info("River client stopped", "testName", testName)
		}
	}

	suite.TestDB.Rollback()
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	if err := suite.initialDB.Close(); err != nil {
		log.Printf("Error closing test database: %v", err)
	}
}

type PartialMasterUserRecordDao struct {
	Id                       *string
	MobileNo                 *string
	Email                    *string
	LastName                 *string
	UserStatus               *string
	Password                 *string
	ResetToken               *string
	LedgerCustomerNumber     *string
	DebtwiseCustomerNumber   *int
	DebtwiseOnboardingStatus *string
}

func (suite *IntegrationTestSuite) createTestUser(partialUser PartialMasterUserRecordDao) dao.MasterUserRecordDao {
	sessionId := uuid.New().String()
	dob, _ := time.Parse("01/02/2006", "11/09/2000")

	ledgerPassword, err := utils.EncryptKmsBinary("@8Kf0exhwDN6$sx@$3nazrABuaVBQxsI")
	suite.Require().NoError(err, "Failed to encrypt ledgerPassword")

	debtwiseCustomerNumber := 1

	testUser := dao.MasterUserRecordDao{
		Id:                         sessionId,
		FirstName:                  "Test",
		LastName:                   "Bar",
		Email:                      "testuser@gmail.com",
		MobileNo:                   "+14159871234",
		DOB:                        dob,
		KmsEncryptedLedgerPassword: ledgerPassword,
		UserStatus:                 constant.ACTIVE,
		StreetAddress:              "123 Main St",
		ApartmentNo:                "",
		ZipCode:                    "11111",
		City:                       "Adak",
		State:                      "AK",
		LedgerCustomerNumber:       "101110007687",
		DebtwiseCustomerNumber:     &debtwiseCustomerNumber,
		DebtwiseOnboardingStatus:   "uninitialized",
	}

	if partialUser.Id != nil {
		testUser.Id = *partialUser.Id
	}
	if partialUser.Email != nil {
		testUser.Email = *partialUser.Email
	}
	if partialUser.MobileNo != nil {
		testUser.MobileNo = *partialUser.MobileNo
	}
	if partialUser.LastName != nil {
		testUser.LastName = *partialUser.LastName
	}
	if partialUser.UserStatus != nil {
		testUser.UserStatus = *partialUser.UserStatus
	}
	if partialUser.Password != nil {
		encryptedPassword, err := utils.EncryptKmsBinary(*partialUser.Password)
		suite.Require().NoError(err, "Failed to encrypt example ledgerPassword")
		testUser.Password = []byte(encryptedPassword)
	}
	if partialUser.ResetToken != nil {
		testUser.ResetToken = *partialUser.ResetToken
	}
	if partialUser.LedgerCustomerNumber != nil {
		testUser.LedgerCustomerNumber = *partialUser.LedgerCustomerNumber
	}
	if partialUser.DebtwiseCustomerNumber != nil {
		testUser.DebtwiseCustomerNumber = partialUser.DebtwiseCustomerNumber
	}
	if partialUser.DebtwiseOnboardingStatus != nil {
		testUser.DebtwiseOnboardingStatus = *partialUser.DebtwiseOnboardingStatus
	}

	err = suite.TestDB.Create(&testUser).Error
	suite.Require().NoError(err, "Failed to create test user")
	return testUser
}

func (suite *IntegrationTestSuite) createMembershipRecord(userId string) dao.UserMembershipDao {
	membershipRecord := dao.UserMembershipDao{
		Id:               uuid.New().String(),
		UserID:           userId,
		MembershipStatus: constant.SUBSCIBED,
	}
	err := suite.TestDB.Create(&membershipRecord).Error
	suite.Require().NoError(err, "Failed to create mebership record")
	return membershipRecord
}

func (suite *IntegrationTestSuite) configEmail() {
	config.Config.Email.ApiBase = "http://localhost:5001"
	config.Config.Email.TemplateDirectory = "../email-templates/"
	config.Config.Email.Domain = "mg.netxd.com"
	config.Config.Email.ApiKey = "sendgrid-api-key"
	config.Config.Email.FromAddr = "Support@DreamFi.com"
}

func (suite *IntegrationTestSuite) createOtpRecord(user dao.MasterUserRecordDao, apiPath string, createdAt time.Time) dao.MasterUserOtpDao {
	otpSessionId := uuid.New().String()

	userOtpRecord := dao.MasterUserOtpDao{
		OtpId:     otpSessionId,
		Otp:       "123456",
		OtpStatus: constant.OTP_SENT,
		Email:     user.Email,
		ApiPath:   apiPath,
		UserId:    user.Id,
		IP:        "1.1.1.1",
		CreatedAt: createdAt,
	}

	err := suite.TestDB.Select("otp_id", "otp", "otp_status", "email", "api_path", "user_id", "ip", "created_at").Create(&userOtpRecord).Error
	suite.Require().NoError(err, "Failed to insert otp record")
	return userOtpRecord
}

func (suite *IntegrationTestSuite) newRiver(cfg *config.Configs, testName string) *river.Client[*sql.Tx] {
	workers := river.NewWorkers()

	handler.RegisterTransactionMonitoringWorker(workers)

	handler.RegisterRefreshBalancesWorker(workers, plaid.NewPlaid(cfg))
	handler.RegisterStatementNotificationWorker(workers)
	statementNotificationBatchWorker := handler.RegisterStatementNotificationEmailEnqueueBatchWorker(workers, nil)

	riverClient, err := river.NewClient(riverdatabasesql.New(suite.initialDB.DB()), &river.Config{
		FetchPollInterval: 50 * time.Millisecond,
		FetchCooldown:     25 * time.Millisecond,
		TestOnly:          true,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 1},
			"debtwise":         {MaxWorkers: 1},
			"plaid":            {MaxWorkers: 1},
			"sendgrid":         {MaxWorkers: 1},
		},
		Workers: workers,
	})
	suite.Require().NoError(err)
	logging.Logger.Info("River client initialized", "testName", testName)

	statementNotificationBatchWorker.SetRiverClientForBatchWorker(riverClient)

	ctx := context.Background()

	err = riverClient.Start(ctx)
	suite.Require().NoError(err)

	return riverClient
}

func (suite *IntegrationTestSuite) WaitForJobsDone(jobCount int) {
	subscribeChan, subscribeCancel := suite.riverClient.Subscribe(river.EventKindJobCompleted)
	defer subscribeCancel()

	jobsSeen := 0
	for {
		select {
		case <-subscribeChan:
			jobsSeen += 1
			if jobsSeen >= jobCount {
				return
			}
		case <-time.After(3 * time.Second):
			suite.FailNow("WaitForJobsDone timed out", "Expected %d jobs. Received %d jobs.", jobCount, jobsSeen)
		}
	}
}

func (suite *IntegrationTestSuite) newHandler() *handler.Handler {
	Config := config.NewConfig()
	Config.Plaid.Environment = "http://localhost:5007"

	return &handler.Handler{
		Config:      Config,
		Plaid:       plaid.NewPlaid(Config),
		RiverClient: suite.riverClient,
		Env:         "test",
	}
}

func (suite *IntegrationTestSuite) createUserPublicKeyRecord(userId string) dao.UserPublicKey {
	encryptedApiKey, _ := utils.EncryptKmsBinary("c077ad8b3d6f40c9896f5fb475f738d6")

	userPublicKey := dao.UserPublicKey{
		UserId:             userId,
		KeyId:              "12345",
		KmsEncryptedApiKey: encryptedApiKey,
		PublicKey:          "test_public_key",
	}
	err := suite.TestDB.Create(&userPublicKey).Error
	suite.Require().NoError(err, "Failed to insert test userPublicKey record")
	return userPublicKey
}

func (suite *IntegrationTestSuite) createUserAccountCard(userId string) dao.UserAccountCardDao {
	userAccountCard := dao.UserAccountCardDao{
		CardHolderId:  "CH0000060090",
		CardId:        "fcf2f39199174939fe437",
		AccountNumber: "987546218371925",
		AccountStatus: "ACTIVE",
		UserId:        userId,
	}

	err := suite.TestDB.Create(&userAccountCard).Error
	suite.Require().NoError(err, "Failed to insert card and cardHolder data")
	return userAccountCard
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
