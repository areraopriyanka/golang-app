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

	// Endpoints for users who are logged in but need to register device id
	unregisteredGroup := e.Group(clientUrl+"register-device/otp", security.LoggedInUnregisteredUserMiddleware, security.ResetTokenMiddleware)

	// Endpoints for users who need to authenticate to recover onboarding
	recoverOnboardingGroup := e.Group(clientUrl+"recover-onboarding", security.RecoverOnboardingUserMiddleware, security.ResetTokenMiddleware)

	refreshSessionGroup := e.Group(clientUrl+"refresh-session", security.RefreshSessionMiddleware, security.ResetTokenMiddleware)

	refreshSessionGroup.GET("", RefreshSession)

	e.GET(clientUrl+"debtwise", DebtwiseHandler)
	e.POST(clientUrl+"sort-job/start", h.StartSortJob)
	e.GET(clientUrl+"sort-job/status", h.GetSortJobStatus)

	e.PUT(clientUrl+"onboarding/customer", CreateUser)
	e.POST(clientUrl+"onboarding/customer/account", CreateUserAccount)
	e.POST(clientUrl+"onboarding/emailDuplicate", IsEmailDuplicate)
	e.PUT(clientUrl+"onboarding/customer/:userId/mobile", SendMobileVerificationOtp)
	e.POST(clientUrl+"onboarding/customer/:userId/mobile", VerifyMobileVerificationOtp)
	// Todo: To remove this API once new onboarding is in place, as we will not use for the new flow
	e.POST(clientUrl+"onboarding/customer/:userId/dob", UpdateCustomerDOB)
	e.POST(clientUrl+"onboarding/customer/:userId/address", UpdateCustomerAddress)
	e.POST(clientUrl+"onboarding/customer/:userId/password", UpdateCustomerPassword)
	e.POST(clientUrl+"onboarding/customer/:userId", UpdateCustomer)
	e.PUT(clientUrl+"onboarding/customer/:userId/kyc", h.SubmitUserDataToSardine)
	e.PUT(clientUrl+"onboarding/customer/:userId/complete-ledger-customer", CompleteLedgerCustomer)

	e.POST(clientUrl+"login", Login)
	e.GET(clientUrl+"version", GetApplicationVersion)

	// Reset password API's
	e.PUT(clientUrl+"reset-password/send-otp", SendResetPasswordOTP)
	e.POST(clientUrl+"reset-password/verify-otp", VerifyResetPasswordOTP)
	e.POST(clientUrl+"reset-password", ResetPassword)

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
	onboardingGroup.POST("/customer/agreements/terms-of-service", ReviewTOSAgreements)
	onboardingGroup.POST("/customer/agreements/card-agreements", ReviewCardAgreements)
	// Smarty address autoComplete
	onboardingGroup.GET("/customer/address-autocomplete", AddressAutoComplete)
	onboardingGroup.POST("/customer/secondary-address-autocomplete", SecondaryAddressAutoComplete)
	onboardingGroup.PUT("/customer/kyc", h.SubmitUserDataToSardine)
	onboardingGroup.GET("/customer/kyc/:jobId", h.GetSardineJobStatus)

	onboardingGroup.PUT("/customer/complete-ledger-customer", CompleteLedgerCustomer)

	// card API's
	accountGroup.GET("/cards", GetCardDetails)
	if env != constant.PROD {
		accountGroup.POST("/cards/cvv", GetCardCvvFromVisa)
	}
	accountGroup.POST("/cards/validate-cvv/build", BuildValidateCvvPayload)
	accountGroup.POST("/cards/validate-cvv", ValidateCvv)
	accountGroup.POST("/cards/pin/build", BuildCardPinPayload)
	accountGroup.POST("/cards/pin", SetCardPin)
	accountGroup.POST("/cards/pin/otp", CreateSetCardPinOTP)
	accountGroup.POST("/cards/pin/otp/verify", ChallengeSetCardPinOtp)
	accountGroup.POST("/cards/replace/build", BuildReplaceCardPayload)
	// To replace a card, its status must first be set to LOST_STOLEN.
	accountGroup.POST("/cards/replace", ReplaceCard)
	accountGroup.POST("/cards/activate/build", BuildActivateCardPayload)
	accountGroup.POST("/cards/freeze", FreezeCard)
	accountGroup.POST("/cards/unfreeze", UnfreezeCard)

	// dashboard API's
	accountGroup.GET("/dashboard/accounts", ListAccounts)
	accountGroup.POST("/dashboard/card-limit/build", BuildGetCardLimitPayload)
	accountGroup.POST("/dashboard/card-limit", GetCardLimit)

	accountGroup.GET("/accounts/transactions", ListTransactions)
	accountGroup.POST("/accounts/ach/push/build", h.BuildTransactionAchPushPayload)
	accountGroup.POST("/accounts/ach/push", h.TransactionAchPush)
	accountGroup.POST("/accounts/ach/pull/build", h.BuildTransactionAchPullPayload)
	accountGroup.POST("/accounts/ach/pull", h.TransactionAchPull)

	// Handler to suspend an account for 60 days
	accountGroup.PUT("/close", SuspendAccount)
	accountGroup.GET("/closure-status", GetAccountClosureStatus)

	// Monthly statement API's
	accountGroup.POST("/list-statements/build", BuildListStatementsPayload)
	accountGroup.POST("/list-statements", ListStatements)
	accountGroup.POST("/get-statement/build", BuildGetStatementPayload)
	accountGroup.POST("/get-statement", GetStatement)

	accountGroup.POST("/change-password", ChangePassword)
	accountGroup.POST("/change-password/send-otp", SendChangePasswordOTP)
	accountGroup.POST("/change-password/verify-otp", ChallengeChangePasswordOTP)

	accountGroup.POST("/debtwise/user", h.CreateDebtwiseUser)
	accountGroup.POST("/debtwise/equifax/invoke", h.EquifaxInvokeHandler)
	accountGroup.POST("/debtwise/equifax/otp", EquifaxValidateOtp)
	accountGroup.POST("/debtwise/equifax/otp/resend", EquifaxOtpResend)
	accountGroup.POST("/debtwise/equifax/consent", EquifaxConsentHandler)
	accountGroup.POST("/debtwise/equifax/identity", h.EquifaxIdentityHandler)
	accountGroup.GET("/debtwise/credit_score_overview", CreditScoreOverviewHandler)
	accountGroup.GET("/debtwise/latest_credit_score", LatestCreditScoreHandler)
	accountGroup.POST("/debtwise/equifax/complete", h.CompleteDebtwiseOnboarding)
	accountGroup.GET("/debtwise/equifax/complete/:jobId", h.GetCreditScoreJobStatus)

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

	unregisteredGroup.POST("", GetDeviceRegistrationOtp)
	unregisteredGroup.POST("/verify", ChallengeDeviceRegistrationOtp)

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
