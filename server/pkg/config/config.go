package config

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var Config *Configs = &Configs{}

// Config the configuration for the application
type Configs struct {
	Server           ServerConfigs
	Cors             CorsConfigs
	Database         DatabaseConfigs
	Ledger           LedgerConfigs
	Jwt              JwtConfigs
	Logger           LoggerConfigs
	Aws              AwsConfigs
	AwsSecretManager AwsSecretManagerConfigs
	Encrypt          EncryptConfigs
	Email            EmailConfigs
	Twilio           TwilioConfigs
	Sardine          SardineConfigs
	Debtwise         DebtwiseConfigs
	Plaid            PlaidConfigs
	Otp              OtpConfigs
	Kyc              KycConfigs
	SmartyStreets    SmartyStreetsConfigs
	OnboardingData   OnboardingDataConfig
	Schedulers       SchedulersConfig
	Environment      EnvironmentConfig
	Webhook          WebhookConfig
	Auth0            Auth0Configs
	Admin            AdminConfigs
	Posthog          PosthogConfigs
	Salesforce       SalesforceConfigs
	VisaSimulator    VisaSimulatorConfigs
}

// ServerConfigurations exported
type ServerConfigs struct {
	Port    int    `json:"port"`
	BaseUrl string `json:"baseUrl"`
}

type CorsConfigs struct {
	AllowOrigins []string `json:"allowOrigins"`
}

// DatabaseConfigurations exported
type DatabaseConfigs struct {
	Host            string        `json:"db-host"`
	Port            int           `json:"db-port"`
	DBName          string        `json:"db-name"`
	DBUser          string        `json:"db-user"`
	DBPassword      string        `json:"db-password"`
	SSLMode         string        `json:"sslMode"`
	MaxOpenConns    int           `json:"maxOpenConns"`
	MaxIdleConns    int           `json:"maxIdleConns"`
	ConnMaxLifetime time.Duration `json:"connMaxLifetime"`
	ConnMaxIdleTime time.Duration `json:"connMaxIdleTime"`
}

// WebhookConfig exported
type WebhookConfig struct {
	PublicKey string `json:"webhook-publicKey"`
}

// LedgerConfigurations exported
type LedgerConfigs struct {
	Endpoint          string `json:"endpoint"`
	LedgerCategory    string `json:"ledgerCategory"`
	CardsEndpoint     string `json:"cardsEndpoint"`
	CardsPublicKey    string `json:"cardsPublicKey"`
	CardsProduct      string `json:"cardsProduct"`
	CardsChannel      string `json:"cardsChannel"`
	CardsProgram      string `json:"cardsProgram"`
	PaymentsEndpoint  string `json:"paymentsEndpoint"`
	MobileRpcEndpoint string `json:"mobileRpcEndpoint"`
	KeyId             string `json:"ledger-keyId"`
	ApiKey            string `json:"ledger-apiKey"`
	Credential        string `json:"ledger-credential"`
	PrivateKey        string `json:"ledger-privateKey"`
	ClientUrl         string `json:"clientUrl"`
}

// JwtConfigurations exported
type JwtConfigs struct {
	SecreteKey                   string `json:"jwt-secreteKey"`
	TimeoutInMinutes             int    `json:"timeout"`
	MinPasswordLength            int    `json:"minPasswordLength"`
	OnboardingUserAutoLogoffTime int    `json:"onboardingUserAutoLogoffTime"`
	LedgerUserAutoLogoffTime     int    `json:"ledgerUserAutoLogoffTime"`
	BufferTimeRefreshToken       int    `json:"bufferTimeRefreshToken"`
	LedgerTokenExpTime           int    `json:"ledgerTokenExpTime"`
}

// LoggerConfigurations exported
type LoggerConfigs struct {
	Filename                   string `json:"filename"`
	MaxSize                    int    `json:"maxSize"`
	MaxBackups                 int    `json:"maxBackups"`
	MaxAge                     int    `json:"maxAge"`
	Compress                   bool   `json:"compress"`
	Directory                  string `json:"directory"`
	DeleteLogFileOlderThanDays int    `json:"deleteLogFileOlderThanDays"`
}

// AwsConfigs exported
type AwsConfigs struct {
	Region                 string        `json:"aws-region"`
	AccessKeyId            string        `json:"aws-accessKeyId"`
	SecreteKeyId           string        `json:"aws-secreteKeyId"`
	BucketName             string        `json:"aws-bucketName"`
	PreSignedUrlExpiration time.Duration `json:"preSignedUrlExpiration"`
	KmsEncryptionKeyId     string        `json:"aws-kmsEncryptionKeyId"`
}

// AwsSecretManagerConfigs exported
type AwsSecretManagerConfigs struct {
	Region       string `json:"region"`
	AccessKeyId  string `json:"accessKeyId"`
	SecreteKeyId string `json:"secreteKeyId"`
	SecretName   string `json:"secretName"`
}

// EncryptConfigs exported
type EncryptConfigs struct {
	EncryptionKey string `json:"encrypt-encryptionKey"`
}

// EmailConfigurations exported
type EmailConfigs struct {
	Domain            string `json:"domain"`
	ApiKey            string `json:"email-apiKey"`
	ApiBase           string `json:"email-apiBase"`
	FromName          string `json:"fromName"`
	FromAddr          string `json:"fromAddr"`
	ClientName        string `json:"clientName"`
	TemplateDirectory string `json:"templateDirectory"`
}

// TwilioConfigurations exported
type TwilioConfigs struct {
	AccountSid  string `json:"twilio-accountSid"`
	AuthToken   string `json:"twilio-authToken"`
	From        string `json:"twilio-from"`
	ApiBase     string `json:"twilio-apiBase"`
	CallbackUrl string `json:"twilio-callbackUrl"`
}

// SardineConfigs exported
type SardineConfigs struct {
	Credential       string `json:"sardine-credential"`
	ApiBase          string `json:"sardine-apiBase"`
	SendTransactions bool   `json:"sardine-sendTransactions"`
}

// DebtwiseConfigs exported
type DebtwiseConfigs struct {
	Credential string `json:"debtwise-credential"`
	ApiBase    string `json:"debtwise-apiBase"`
}

type PlaidConfigs struct {
	ClientId        string `json:"plaid-clientId"`
	Secret          string `json:"plaid-secret"`
	Environment     string `json:"plaid-environment"`
	LinkRedirectURI string `json:"plaid-linkRedirectURI"`
	WebhookURL      string `json:"plaid-webhookURL"`
}

// OtpConfigurations exported
type OtpConfigs struct {
	HardcodedOtp      string `json:"hardcodedOtp"`
	UseHardcodedOtp   bool   `json:"useHardcodedOtp"`
	MaxOtpRetryCount  int    `json:"maxOtpRetryCount"`
	OtpExpiryDuration int    `json:"otpExpiryDuration"`
	OtpDigits         int    `json:"otpDigits"`
}

// KycConfigs exported
type KycConfigs struct {
	Credential         string `json:"kyc-credential"`
	KeyId              string `json:"kyc-keyId"`
	PrivateKey         string `json:"kyc-privateKey"`
	Algorithm          string `json:"algorithm"`
	GetKycUrl          string `json:"getkycurl"`
	PostKYCUrl         string `json:"postkycurl"`
	AddKycDocumentsUrl string `json:"addkycdocumentsurl"`
	GetKycDocumentsUrl string `json:"getkycdocumentsurl"`
	ApiKey             string `json:"kyc-apiKey"`
}

// AddressAutoCompleteConfigs exported
type SmartyStreetsConfigs struct {
	AuthId    string `json:"smartyStreets-authId"`
	AuthToken string `json:"smartyStreets-authToken"`
	LookupUrl string `json:"lookupUrl"`
}

// OnboardingDataConfig exported
type OnboardingDataConfig struct {
	FileFormatsAllowed string `json:"fileFormatsAllowed"`
}

// SchedulersConfig exported
type SchedulersConfig struct {
	DeleteLogCronExp              string `json:"deleteLogCronExp"`
	DeleteExpiredOTPsCronExp      string `json:"deleteExpiredOTPsCronExp"`
	DeleteOldNotificationsCronExp string `json:"deleteOldNotificationsCronExp"`
	DeleteOldLedgerTokensCronExp  string `json:"deleteOldLedgerTokensCronExp"`
	CloseSuspendedAccountsCronExp string `json:"closeSuspendedAccountCronExp"`
}

// EnvironmentConfig exported
type EnvironmentConfig struct {
	EnvName string `json:"envName"`
}

// Auth0Configs exported
type Auth0Configs struct {
	Domain            string `json:"domain"`
	ClientId          string `json:"clientId"`
	ClientSecret      string `json:"clientSecret"`
	RedirectUrl       string `json:"redirectUrl"`
	LogoutReturnToUrl string `json:"logoutReturnToUrl"`
}

type AdminConfigs struct {
	SessionSecretKey string `json:"SessionSecretKey"`
}

type PosthogConfigs struct {
	BaseUrl    string `json:"baseUrl"`
	ProjectKey string `json:"projectKey"`
}

type SalesforceConfigs struct {
	IntegrationEnabled   bool   `json:"integrationEnabled"`
	InboundAuth0Domain   string `json:"inboundAuth0Domain"`
	InboundAuth0Audience string `json:"inboundAuth0Audience"`
}

type VisaSimulatorConfigs struct {
	BaseUrl string `json:"baseUrl"`
}

func ReadConfig(configJson *io.Reader) error {
	return initViper(Config, configJson)
}

func initViper(config *Configs, configJson *io.Reader) error {
	SetDefaults()

	// Enable VIPER to read Environment Variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	viper.SetConfigType("json")
	if configJson != nil {
		if err := viper.ReadConfig(*configJson); err != nil {
			fmt.Printf("\nError reading config json, %s", err)
		}
	}
	err := viper.Unmarshal(&config)
	if err != nil {
		fmt.Printf("\nUnable to decode into struct, %v", err)
		return err
	}
	return nil
}

func NewConfig() *Configs {
	config := &Configs{}
	err := initViper(config, nil)
	if err != nil {
		log.Fatal("Failed to read config: " + err.Error())
	}
	return config
}

func SetDefaults() {
	// These defaults are intended for running the server directly on the developer's machine
	viper.SetDefault("admin.sessionsecretkey", nil)
	viper.SetDefault("auth0.domain", nil)
	viper.SetDefault("auth0.clientid", nil)
	viper.SetDefault("auth0.clientsecret", nil)
	viper.SetDefault("auth0.redirecturl", "http://localhost:5000/admin/callback")
	viper.SetDefault("auth0.logoutreturntourl", "http://localhost:5000/admin/login")
	viper.SetDefault("aws.accesskeyid", nil)
	viper.SetDefault("aws.secretekeyid", nil)
	viper.SetDefault("aws.bucketname", "springct-crump-sandbox")
	viper.SetDefault("aws.kmsencryptionkeyid", "fafd880a-3997-4a56-8418-8f7cdfddcbe6")
	viper.SetDefault("aws.presignedurlexpiration", 900000)
	viper.SetDefault("aws.region", "us-east-1")
	viper.SetDefault("awssecretmanager.accesskeyid", nil)
	viper.SetDefault("awssecretmanager.region", "us-east-1")
	viper.SetDefault("awssecretmanager.secretekeyid", nil)
	viper.SetDefault("awssecretmanager.secretname", nil)
	viper.SetDefault("database.connmaxidletime", 3600000)
	viper.SetDefault("database.connmaxlifetime", 7200000)
	viper.SetDefault("database.dbname", "middleware")
	viper.SetDefault("database.dbpassword", "Dreamfi#1234")
	viper.SetDefault("database.dbuser", "middleware")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.maxidleconns", "10")
	viper.SetDefault("database.maxopenconns", "100")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("email.apibase", "http://localhost:5001")
	viper.SetDefault("email.apikey", nil)
	viper.SetDefault("email.clientname", "DreamFi")
	viper.SetDefault("email.domain", "mg.netxd.com")
	viper.SetDefault("email.fromaddr", "Support@DreamFi.com")
	viper.SetDefault("email.fromname", "NetXD")
	viper.SetDefault("email.templatedirectory", "./email-templates/")
	viper.SetDefault("twilio.apibase", "http://localhost:5003")
	viper.SetDefault("twilio.accountsid", nil)
	viper.SetDefault("twilio.authtoken", nil)
	viper.SetDefault("twilio.from", "+16206348340")
	viper.SetDefault("twilio.callbackurl", "https://middleware.sandbox.dreamfi.com/api/v1/twilio/events")
	viper.SetDefault("sardine.credential", nil)
	viper.SetDefault("sardine.apibase", "http://localhost:5004")
	viper.SetDefault("sardine.sendtransactions", true)
	viper.SetDefault("debtwise.apibase", "http://localhost:5006")
	viper.SetDefault("debtwise.credential", "")
	viper.SetDefault("plaid.secret", nil)
	viper.SetDefault("plaid.clientid", nil)
	// In order to test Plaid Link on the FE, Plaid's environment must be set to
	// either production or sandbox (e.g., "https://sandbox.plaid.com")
	viper.SetDefault("plaid.environment", "http://localhost:5007")
	// linkredirecturi comes from https://github.com/plaid/tiny-quickstart/tree/main/react_native and is used for Plaid sandbox
	viper.SetDefault("plaid.linkredirecturi", "https://cdn-testing.plaid.com/link/v2/stable/sandbox-oauth-a2a-react-native-redirect.html")
	viper.SetDefault("plaid.webhookurl", "")
	viper.SetDefault("encrypt.encryptionkey", nil)
	viper.SetDefault("environment.envname", "dreamfiSandbox")
	viper.SetDefault("jwt.buffertimerefreshtoken", 300000)
	viper.SetDefault("jwt.ledgeruserautologofftime", 180000)
	viper.SetDefault("jwt.minpasswordlength", 8)
	viper.SetDefault("jwt.onboardinguserautologofftime", 180000)
	viper.SetDefault("jwt.secretekey", nil)
	viper.SetDefault("jwt.timeoutinminutes", 30)
	viper.SetDefault("kyc.addkycdocumentsurl", "https://dreamfisb.netxd.com/ekyc/rpc/KycService/AddDocuments")
	viper.SetDefault("kyc.algorithm", "ecdsa-sha256")
	viper.SetDefault("kyc.apikey", nil)
	viper.SetDefault("kyc.credential", nil)
	viper.SetDefault("kyc.keyid", nil)
	viper.SetDefault("kyc.getkycdocumentsurl", "https://dreamfisb.netxd.com/ekyc/rpc/KycService/GetDocuments")
	viper.SetDefault("kyc.getkycurl", "https://dreamfisb.netxd.com/ekyc/rpc/KycService/GetTransactionByReferenceId")
	viper.SetDefault("kyc.postkycurl", "https://dreamfisb.netxd.com/ekyc/rpc/KycService/Check")
	viper.SetDefault("kyc.privatekey", nil)
	viper.SetDefault("ledger.apikey", nil)
	viper.SetDefault("ledger.cardsendpoint", "https://dreamfisb.netxd.com/pl/cardv2")
	// Sourced from RSA configuration in the ledger. Documentation captured:
	// https://dreamfi.atlassian.net/wiki/spaces/DevOps/pages/176816133/NetXD+Ledger+Architecture
	// https://dreamfi.atlassian.net/browse/DT-389
	viper.SetDefault("ledger.cardspublickey", `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApSQgtJ7XXaiYhcTgjuhJ
VWJDdXykujwP4yRCpcEKQBTtdVPkHquPaeDhozCtATO/XoZFJ7d1WtR8Ia//8jZg
FxsFDL7hR8ISR+vVB3sb8cOJvRTCXlMo7OPygi/RpGW8stQaBOSpf6fu4dp8A5Ty
6agiQ1zptDMB2ZHcSbBM36L1ri2UssvPxqn97UTb/3dqwleh0kCITybx8ZspXvji
65YT8JdTPLNEOby+5K8JRlyGJyDmfpjB6SAenBOpOp1JpBlxN8L4+SGINEsYq5//
MAocOQ27Ae2/cV4007V59D1y+FHRrkT++K3freRxIUDmGrzmjfUmMDqB75oD+fYV
4QIDAQAB
-----END PUBLIC KEY-----`)
	// find in mongosh:
	// db.ProgramChannelSettings.find({}, { channel: 1, programName: 1, product: 1, transactionType: 1 })
	viper.SetDefault("ledger.cardsproduct", "PL")
	viper.SetDefault("ledger.cardschannel", "VISA_DPS")
	viper.SetDefault("ledger.cardsprogram", "DREAMFI_MVP")
	viper.SetDefault("ledger.clienturl", "/api/v1/")
	viper.SetDefault("ledger.credential", nil)
	viper.SetDefault("ledger.keyid", nil)
	// Note: different from Product name "DREAMFI_MVP"
	// ┻━┻︵ \(°□°)/ ︵ ┻━┻
	viper.SetDefault("ledger.ledgercategory", "DreamFi_MVP")
	viper.SetDefault("ledger.endpoint", "https://dreamfisb.netxd.com/pl/jsonrpc")
	viper.SetDefault("ledger.mobilerpcendpoint", "https://dreamfisb.netxd.com/gw/mobilerpc")
	viper.SetDefault("ledger.paymentsendpoint", "https://dreamfisb.netxd.com/pl/rpc/paymentv2")
	viper.SetDefault("ledger.privatekey", nil)
	viper.SetDefault("logger.compress", "false")
	viper.SetDefault("logger.deletelogfileolderthandays", 30)
	viper.SetDefault("logger.directory", "logs")
	viper.SetDefault("logger.filename", "processApi")
	viper.SetDefault("logger.maxage", "1")
	viper.SetDefault("logger.maxbackups", 30)
	viper.SetDefault("logger.maxsize", 20)
	viper.SetDefault("onboardingdata.fileformatsallowed", ".pdf,.jpeg,.jpg,.png")
	viper.SetDefault("otp.hardcodedotp", "123456")
	viper.SetDefault("otp.usehardcodedotp", false)
	viper.SetDefault("otp.maxotpretrycount", 3)
	viper.SetDefault("otp.otpdigits", 6)
	viper.SetDefault("otp.otpexpiryduration", 300000)
	viper.SetDefault("schedulers.deleteexpiredotpscronexp", "35 02 * * *")
	viper.SetDefault("schedulers.deletelogcronexp", "30 02 * * *")
	viper.SetDefault("schedulers.deleteoldnotificationscronexp", "40 02 * * *")
	viper.SetDefault("schedulers.deleteoldledgertokenscronexp", "*/30 * * * *")
	viper.SetDefault("schedulers.closesuspendedaccountscronexp", "0 8 * * *")
	viper.SetDefault("server.port", 5000)
	viper.SetDefault("cors.alloworigins", []string{"http://localhost:5000", "http://localhost:5002", "http://localhost:5173", "middleware.sandbox.dreamfi.com"})
	viper.SetDefault("server.baseurl", "https://middleware.sandbox.dreamfi.com/api/v1/")
	viper.SetDefault("smartystreets.authid", nil)
	viper.SetDefault("smartystreets.authtoken", nil)
	viper.SetDefault("smartystreets.lookupurl", "https://us-autocomplete-pro.api.smarty.com/lookup")
	viper.SetDefault("posthog.baseurl", "https://app.posthog.com")
	viper.SetDefault("posthog.projectkey", "phc_vV7tEhmrOmINglr0PAJZIhFda4SV0LlR2kfOj0tt2Y7")
	viper.SetDefault("webhook.publickey", nil)

	viper.SetDefault("salesforce.integrationenabled", false)
	viper.SetDefault("salesforce.inboundauth0domain", nil)
	viper.SetDefault("salesforce.inboundauth0audience", "https://middleware.dreamfi.com/salesforce")

	viper.SetDefault("visasimulator.baseurl", "https://dreamfisb.netxd.com/visafwd")
}

func AssertConfig() error {
	requiredKeys := []string{
		"admin.sessionsecretkey",
		"auth0.domain",
		"auth0.clientid",
		"auth0.clientsecret",
		"email.apikey",
		"sardine.credential",
		"encrypt.encryptionkey",
		"jwt.secretekey",
		"ledger.apikey",
		"ledger.credential",
		"ledger.keyid",
		"ledger.privatekey",
		"twilio.accountsid",
		"twilio.authtoken",
		"smartystreets.authtoken",
		"smartystreets.authid",
		"plaid.secret",
		"plaid.clientid",
		"webhook.publickey",
		"posthog.projectkey",
	}

	missingKeys := []string{}
	for _, key := range requiredKeys {
		if viper.Get(key) == nil {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		return fmt.Errorf("missing required keys: %v", missingKeys)
	}
	return nil
}
