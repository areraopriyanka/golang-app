package main

//nolint:gofumpt
import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"process-api/pkg/clock"
	"process-api/pkg/config"
	"process-api/pkg/db"
	"process-api/pkg/handler"
	"process-api/pkg/logging"
	"process-api/pkg/maintenance"
	"process-api/pkg/plaid"
	"process-api/pkg/resource/agreements"
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"process-api/pkg/validators"
	"process-api/pkg/version"

	_ "process-api/pkg/docs"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/robfig/cron/v3"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func init() {
	logging.InitLogger()
}

// @title DreamFi Middleware API
// @version 1.0
// @BasePath /api/v1
func main() {
	ctx := context.Background()

	env := utils.GetEnv()

	region, hasEnvRegion := os.LookupEnv("AWSSECRETMANAGER_REGION")
	if !hasEnvRegion {
		region = "us-east-1"
	}

	secretName := os.Getenv("AWSSECRETMANAGER_SECRETNAME")

	var secretJsonReader *io.Reader
	if secretName != "" {
		logging.Logger.Info("Retrieving secret", "secretName", secretName)
		secretJson, err := utils.GetSecret(region, secretName)
		if err != nil {
			logging.Logger.Error("failed to read secretfor config", "secretName", secretName, "error", err.Error())
		} else {
			logging.Logger.Info("Retrieved secret", "secretName", secretName)
			var temp io.Reader = strings.NewReader(secretJson)
			secretJsonReader = &temp
		}
	}

	err := config.ReadConfig(secretJsonReader)
	if err != nil {
		logging.Logger.Error("Failed to read config: " + err.Error())
	}

	logging.Logger.Info("Application Version:: " + version.ApplicationVersion)
	logging.Logger.Info("Run environment set received", "env", env)

	// Print the updated Configs struct
	logging.Logger.Info("Configurations", "config", config.Config)

	err = config.AssertConfig()
	if err != nil {
		panic(fmt.Sprintf("Can not start: %s", err))
	}

	err = db.DBconnection()
	if err != nil {
		// This is an implementation limitation of gorm. We can not recover from the initial Open failing
		// https://github.com/go-gorm/gorm/issues/7241
		// TODO: refactor the db pkg to remove the DB singleton so we can retry open connections after startup
		panic(fmt.Sprintf("Can not start with failed database connection: %s", err))
	}
	err = db.Automigrate(db.DB.DB(), "./")
	if err != nil {
		panic(fmt.Errorf("failed to Automigrate: %w", err))
	}
	err = agreements.PrepareAgreements()
	if err != nil {
		panic(fmt.Sprintf("Can not start with failed agreement preparation: %s", err))
	}
	utils.AwsConnection()
	utils.AwsKmsConnection()
	utils.InitializeTwilioClient(config.Config.Twilio)
	err = validators.InitValidationRules()
	if err != nil {
		log.Fatal("Error in initializing validation rules: " + err.Error())
	}

	err = utils.InitializePosthogClient(config.Config.Posthog)
	if err != nil {
		panic(fmt.Sprintf("Error in initializing posthog client: %s", err.Error()))
	}

	plaidClient := plaid.NewPlaid(config.Config)
	workers := river.NewWorkers()

	handler.RegisterStatementNotificationWorker(workers)
	handler.RegisterRefreshBalancesWorker(workers, plaidClient)

	riverClient, err := river.NewClient(riverdatabasesql.New(db.DB.DB()), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
			"debtwise":         {MaxWorkers: 100},
			"plaid":            {MaxWorkers: 100},
			"sendgrid":         {MaxWorkers: 100},
		},
		Workers: workers,
	})
	if err != nil {
		panic(err)
	}

	logging.Logger.Info("River client initialized")

	go func() {
		if err := riverClient.Start(ctx); err != nil {
			panic(err)
		}
	}()

	err = security.InitAuth0(config.Config.Auth0)
	if err != nil {
		log.Fatal("Error in initializing Auth0: " + err.Error())
	}

	e := handler.NewEcho()
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	clientUrl := config.Config.Ledger.ClientUrl

	h := handler.Handler{
		Config:      config.Config,
		Plaid:       plaid.NewPlaid(config.Config),
		Env:         env,
		RiverClient: riverClient,
	}

	h.BuildRoutes(e, clientUrl, env)

	// TODO: Remove `legacyClientUrl` support after 2025-11-01
	// Currently both `/process-api/evolvingsb` and `/api/v1` are supported to maintain backward compatibility.
	legacyClientUrl := "/process-api/evolvingsb/"
	h.BuildRoutes(e, legacyClientUrl, env)

	c := cron.New()

	// Schedule DeleteExpiredOTPs() to run at 2:35 AM daily
	_, err = c.AddFunc(config.Config.Schedulers.DeleteExpiredOTPsCronExp, func() {
		maintenance.DeleteExpiredOTPs()
		logging.Logger.Info("Ended DeleteExpiredOTPs() scheduler at", "time", clock.Now())
	})
	if err != nil {
		logging.Logger.Error("Failed to add function", "error", err)
	}

	// Schedule DeleteOldLedgerTokenRecords() to run after every 30 minutes
	_, err = c.AddFunc(config.Config.Schedulers.DeleteOldLedgerTokensCronExp, func() {
		maintenance.DeleteLedgerTokenRecords()
		logging.Logger.Info("Ended DeleteLedgerTokenRecords() scheduler at", "time", clock.Now())
	})
	if err != nil {
		logging.Logger.Error("Failed to add function", "error", err)
	}

	c.Start()

	sigintOrTerm := make(chan os.Signal, 1)
	signal.Notify(sigintOrTerm, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		PORT := strconv.Itoa(config.Config.Server.Port)
		if err := e.Start(":" + PORT); err != nil && err != http.ErrServerClosed {
			logging.Logger.Error("Error while starting the server", "error", err)
		}
	}()

	<-sigintOrTerm
	logging.Logger.Info("received SIGINT/SIGTERM and stopping")

	if err := riverClient.Stop(ctx); err != nil {
		logging.Logger.Error("Error stopping river client", "error", err)
	}

	if err := utils.ClosePosthogClient(); err != nil {
		logging.Logger.Error("Error in closing posthog client", "error", err)
	}

	err = e.Shutdown(ctx)
	if err != nil {
		logging.Logger.Info("echo shutdown", "error", err.Error())
		panic(err)
	}
	logging.Logger.Info("echo shutdown succeeded")
}
