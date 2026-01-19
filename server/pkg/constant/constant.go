package constant

// store all common constants in this file
const (
	EMAIL_REGX                         = "^[_A-Za-z0-9+-]+(\\.[_A-Za-z0-9+-]+)*@[A-Za-z0-9-]+(\\.[A-Za-z0-9]+)*(\\.[A-Za-z]{2,})$"
	MOBILE_REGX                        = "^[0-9]{10}$"
	AWS_PROTOCOL                       = "https://"
	AWS_S3_DOMAIN                      = ".s3.amazonaws.com/"
	DB_DRIVER                          = "postgres"
	EMAIL_VERIFICATION_TEMPLATE_NAME   = "emailVerificationOtpTemplate.html"
	RESET_PASSWORD_EMAIL_TEMPLATE_NAME = "resetPasswordOtpTemplate.html"
)
