package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/salesforce"
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"process-api/pkg/validators"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "process-api/pkg/docs-salesforce"

	echoSwagger "github.com/swaggo/echo-swagger"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (c CustomValidator) Validate(i any) error {
	if err := c.validator.Struct(i); err != nil {
		switch validationErr := err.(type) {
		case validator.ValidationErrors:
			errors := make([]response.BadRequestError, len(validationErr))
			for index, error := range validationErr {
				var errorText string
				if error.Param() != "" {
					errorText = fmt.Sprintf("%s=%s", error.ActualTag(), error.Param())
				} else {
					errorText = error.ActualTag()
				}

				errors[index] = response.BadRequestError{
					FieldName: error.Field(),
					Error:     errorText,
				}
			}
			return response.BadRequestErrors{
				Errors: errors,
			}
		case *validator.InvalidValidationError:
			message := fmt.Sprintf("validate struct failed with invalid tags: %s", validationErr.Error())
			return response.ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				LogMessage: message,
				ErrorCode:  "INTERNAL_SERVER_ERROR",
				Message:    message,
			}
		default:
			message := fmt.Sprintf("validate struct failed with unexpected error: %s", validationErr.Error())
			return response.ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				LogMessage: message,
				ErrorCode:  "INTERNAL_SERVER_ERROR",
				Message:    message,
			}
		}
	}
	return nil
}

func NewEcho() *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = utils.CustomHTTPErrorHandler
	customValidator, err := validators.NewValidator()
	if err != nil {
		logging.Logger.Error("validators.NewValidator() failed", "err", err.Error())
		panic(err)
	}
	e.Validator = CustomValidator{validator: customValidator}
	e.Use(logging.RequestContextLogger)

	bodyDumpConfig := middleware.BodyDumpConfig{
		Handler: logging.RequestResponseBodyLogger,
		Skipper: func(c echo.Context) bool {
			return logging.Logger.Handler().Enabled(context.Background(), slog.LevelDebug)
		},
	}
	e.Use(middleware.BodyDumpWithConfig(bodyDumpConfig))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     config.Config.Cors.AllowOrigins,
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowCredentials: true,
		ExposeHeaders: []string{
			echo.HeaderAuthorization,
			"X-Auth-Token",
		},
	}))

	return e
}

func (h *Handler) BuildRoutes(e *echo.Echo, clientUrl string, env string) {
	accountGroup := e.Group(clientUrl+"account", security.LoggedInRegisteredUserMiddleware, security.ResetTokenMiddleware)
	onboardingGroup := e.Group(clientUrl+"onboarding", security.OnboardingUserMiddleware, security.ResetTokenMiddleware)

	// Endpoints for users who need to authenticate to recover onboarding
	recoverOnboardingGroup := e.Group(clientUrl+"recover-onboarding", security.RecoverOnboardingUserMiddleware, security.ResetTokenMiddleware)

	refreshSessionGroup := e.Group(clientUrl+"refresh-session", security.RefreshSessionMiddleware, security.ResetTokenMiddleware)

	refreshSessionGroup.GET("", RefreshSession)

	e.PUT(clientUrl+"onboarding/customer", CreateUser)

	e.POST(clientUrl+"onboarding/emailDuplicate", IsEmailDuplicate)
	e.PUT(clientUrl+"onboarding/customer/:userId/mobile", SendMobileVerificationOtp)
	e.POST(clientUrl+"onboarding/customer/:userId/mobile", VerifyMobileVerificationOtp)
	// Todo: To remove this API once new onboarding is in place, as we will not use for the new flow
	e.POST(clientUrl+"onboarding/customer/:userId/dob", UpdateCustomerDOB)
	e.POST(clientUrl+"onboarding/customer/:userId/address", UpdateCustomerAddress)
	e.POST(clientUrl+"onboarding/customer/:userId/password", UpdateCustomerPassword)
	e.POST(clientUrl+"onboarding/customer/:userId", UpdateCustomer)

	e.POST(clientUrl+"login", Login)
	e.GET(clientUrl+"version", GetApplicationVersion)

	// Endpoint that generates and returns a TwiML XML response for OTP voice calls
	e.POST(clientUrl+"voice-xml", GenerateVoiceXML)

	e.POST(clientUrl+"twilio/events", security.TwilioWebhookMiddleware(TwilioWebhookHandler))
	e.POST(clientUrl+"ledger/events", h.LedgerWebhookHandler)
	e.POST(clientUrl+"plaid/webhooks", h.PlaidWebhookHandler)

	e.GET(clientUrl+"agreements/:agreementName", GetAgreement)
	e.GET(clientUrl+"agreements/pdf/:agreementName", GetAgreementPDF)

	// New onboarding API's
	onboardingGroup.POST("/customer/details", AddPersonalDetails)
	onboardingGroup.PUT("/customer/mobile", SendMobileVerificationOtp)
	onboardingGroup.POST("/customer/mobile", VerifyMobileVerificationOtp)
	onboardingGroup.POST("/customer/address", UpdateCustomerAddress)

	// card API's
	accountGroup.GET("/cards", GetCardDetails)
	if env != constant.PROD {
		accountGroup.POST("/cards/cvv", GetCardCvvFromVisa)
	}

	// To replace a card, its status must first be set to LOST_STOLEN.

	// dashboard API's
	accountGroup.GET("/dashboard/accounts", ListAccounts)

	accountGroup.POST("/accounts/ach/pull", h.TransactionAchPull)

	// Handler to suspend an account for 60 days

	accountGroup.GET("/closure-status", GetAccountClosureStatus)

	accountGroup.GET("/status/jobs/:jobId", h.JobStatusHandler)

	accountGroup.GET("/personal-details", GetPersonalDetails)
	accountGroup.POST("/plaid/link/token", h.PlaidCreateLinkToken)
	accountGroup.POST("/plaid/link/token/update", h.PlaidUpdateLinkToken)

	accountGroup.POST("/plaid/public_token/exchange", h.PlaidExchangePublicToken)
	accountGroup.DELETE("/plaid/account", h.PlaidAccountUnlink)
	accountGroup.POST("/plaid/accounts/reconnected", h.PlaidAccountsReconnected)
	accountGroup.POST("/balance/refresh", h.BalanceRefresh)
	accountGroup.GET("/balance/refresh/status/:jobId", h.BalanceRefreshStatus)

	// Demographic update APIs
	accountGroup.POST("/customer/demographic-update/mobile", DemographicUpdateSendOtp)
	accountGroup.POST("/customer/demographic-update/mobile/verify", DemographicUpdateVerifyOtp)
	accountGroup.POST("/customer/demographic-update/full-name", SubmitFullNameDemographicUpdates)
	accountGroup.POST("/customer/demographic-update/address", SubmitAddressDemographicUpdates)
	accountGroup.GET("/customer/demographic-update/address-autocomplete", DemographicUpdateAddressAutoComplete)
	accountGroup.POST("/customer/demographic-update/secondary-address-autocomplete", DemographicUpdateSecondaryAddressAutoComplete)
	accountGroup.GET("/customer/demographic-update", GetUserDetailsAndDemographicUpdateStatus)

	// Transaction dispute APIs
	accountGroup.POST("/customer/transaction/:referenceId/dispute", SubmitTransactionDispute)

	// Membership APIs
	accountGroup.GET("/membership/status", GetMemberShipStatus)
	accountGroup.PUT("/membership/status", UpdateMembershipStatus)

	accountGroup.GET("/customer/disclosure-accepted-date", GetDisclosuresAcceptedDate)

	recoverOnboardingGroup.POST("/send-otp", RecoverOnboardingOTP)
	recoverOnboardingGroup.POST("/verify-otp", ChallengeRecoverOnboardingOTP)
}

func (h *Handler) BuildSalesForceRoutes(e *echo.Echo) {
	url := echoSwagger.URL("/api/salesforce/swagger-json/salesforce-swagger.json")
	instanceName := echoSwagger.InstanceName("salesforce")
	e.GET("/api/salesforce/swagger/*", echoSwagger.EchoWrapHandler(url, instanceName))

	salesforceGroup := e.Group("/api/salesforce", salesforce.SalesforceAuth0Middleware())
	salesforceGroup.GET("/accounts/:ledgerAccountID/transactions", salesforce.SalesforceGetTransactions)
	salesforceGroup.GET("/accounts/:ledgerAccountID/balance", salesforce.SalesforceGetBalance)
}
