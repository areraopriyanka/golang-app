package validators

import (
	"fmt"
	"process-api/pkg/clock"
	"process-api/pkg/constant"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"braces.dev/errtrace"
	"github.com/dongri/phonenumber"
	"github.com/go-playground/validator/v10"
)

const (
	NAME_REGX      = "^[a-zA-Z.,\\- ]{1,50}$"
	SSN_REGX       = "^[0-9]{9}$"
	ADDRESS_REGX   = "^[a-zA-Z0-9,\\-./# ]{0,100}$"
	CITY_NAME_REGX = "^[a-zA-Z-. ]{1,40}$"
	STATE_REGX     = "^(AL|AK|AZ|AR|CA|CO|CT|DE|DC|FL|GA|HI|ID|IL|IN|IA|KS|KY|LA|ME|MD|MA|MI|MN|MS|MO|MT|NE|NV|NH|NJ|NM|NY|NC|ND|OH|OK|OR|PA|RI|SC|SD|TN|TX|UT|VT|VA|WA|WV|WI|WY|AS|GU|MP|PR|VI)$"
	// According to NetXD, the regex is "^[0-9]{5}$|^[0-9]{9}$", but the Ledger does not accept 9-digit zip code
	// Currently, restricting it to 5-digit zip code only
	ZIP_CODE_REGX          = "^[0-9]{5}$"
	TRANSACTION_NOTES_REGX = "^[a-zA-Z0-9 ]{1,10}$"
)

var validate = validator.New()

func InitValidationRules() error {
	newValidator, err := NewValidator()
	if err != nil {
		return err
	}
	validate = newValidator
	return nil
}

// Maintains backwards compatibility with the singleton until it can be removed
func NewValidator() (*validator.Validate, error) {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		jsonTag := fld.Tag.Get("json")
		name := strings.SplitN(jsonTag, ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	// Registering custom date validation
	err := validate.RegisterValidation("validateDOB", func(fl validator.FieldLevel) bool {
		return validateDOB(fl.Field().String())
	})
	if err != nil {
		return nil, err
	}

	// Registering custom maxbyte validation
	err = validate.RegisterValidation("maxByte", func(fl validator.FieldLevel) bool {
		maxBytes, err := strconv.Atoi(fl.Param())
		if err != nil {
			return false
		}
		return len(fl.Field().String()) <= maxBytes
	})
	if err != nil {
		return nil, err
	}

	// Registering custom email validation
	err = validate.RegisterValidation("validateEmail", func(fl validator.FieldLevel) bool {
		return validateEmail(fl.Field().String())
	})
	if err != nil {
		return nil, err
	}

	err = validate.RegisterValidation("otpType", func(fl validator.FieldLevel) bool {
		otpType := fl.Field().String()
		return (otpType == constant.CALL || otpType == constant.SMS || otpType == constant.EMAIL)
	})
	if err != nil {
		return nil, err
	}

	// Registering custom ssn validation
	err = validate.RegisterValidation("validateSSN", func(fl validator.FieldLevel) bool {
		return validateSSN(fl.Field().String())
	})
	if err != nil {
		return nil, err
	}

	// Registering custom phone validation
	err = validate.RegisterValidation("validatePhone", func(fl validator.FieldLevel) bool {
		return validatePhone(fl.Field().String())
	})
	if err != nil {
		return nil, err
	}

	// Registering custom address validation
	err = validate.RegisterValidation("validateAddress", func(fl validator.FieldLevel) bool {
		return validateAddress(fl.Field().String())
	})
	if err != nil {
		return nil, err
	}

	// Registering custom state validation
	err = validate.RegisterValidation("validateState", func(fl validator.FieldLevel) bool {
		return validateState(fl.Field().String())
	})
	if err != nil {
		return nil, err
	}

	// Registering custom city name validation
	err = validate.RegisterValidation("validateCity", func(fl validator.FieldLevel) bool {
		return validateCity(fl.Field().String())
	})
	if err != nil {
		return nil, err
	}

	// Registering custom zipcode validation
	err = validate.RegisterValidation("validateZipCode", func(fl validator.FieldLevel) bool {
		return validateZipCode(fl.Field().String())
	})
	if err != nil {
		return nil, err
	}

	// Registering custom ledger DOB validation
	err = validate.RegisterValidation("validateLedgerDOB", func(fl validator.FieldLevel) bool {
		return validateLedgerDOB(fl.Field().String())
	})
	if err != nil {
		return nil, err
	}

	// Registering custom ledger Name validation
	err = validate.RegisterValidation("validateName", func(fl validator.FieldLevel) bool {
		return validateName(fl.Field().String())
	})
	if err != nil {
		return nil, err
	}
	// Registering custom transaction note
	err = validate.RegisterValidation("validateTransactionNote", func(fl validator.FieldLevel) bool {
		return validateTransactionNote(fl.Field().String())
	})
	if err != nil {
		return nil, err
	}
	return validate, nil
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

func ConvertValidationErrorsToString(errors validator.ValidationErrors) string {
	var message string
	for _, err := range errors {
		message += fmt.Sprintf("Field '%s' failed validation with tag '%s'.", err.Field(), err.Tag())
	}
	return message
}

func VerifyAndFormatMobileNumber(phoneNumber string) (string, error) {
	parsedPhoneNumber := phonenumber.Parse(phoneNumber, "US")
	if parsedPhoneNumber == "" {
		return "", errtrace.Wrap(fmt.Errorf("only US phone numbers are valid. received: %s", phoneNumber))
	}
	return "+" + parsedPhoneNumber, nil
}

func validateName(name string) bool {
	if strings.HasPrefix(name, " ") || strings.HasSuffix(name, " ") {
		return false
	}
	re := regexp.MustCompile(NAME_REGX)
	return re.MatchString(name)
}

func validateEmail(email string) bool {
	re := regexp.MustCompile(constant.EMAIL_REGX)
	return re.MatchString(email)
}

func validateSSN(ssn string) bool {
	re := regexp.MustCompile(SSN_REGX)
	return re.MatchString(ssn)
}

func validatePhone(phoneNumber string) bool {
	re := regexp.MustCompile(constant.MOBILE_REGX)
	if re.MatchString(phoneNumber) {
		_, err := VerifyAndFormatMobileNumber(phoneNumber)
		return err == nil
	}
	return false
}

func validateAddress(address string) bool {
	if strings.HasPrefix(address, " ") || strings.HasSuffix(address, " ") {
		return false
	}
	re := regexp.MustCompile(ADDRESS_REGX)
	return re.MatchString(address)
}

func validateState(state string) bool {
	if strings.HasPrefix(state, " ") || strings.HasSuffix(state, " ") {
		return false
	}
	re := regexp.MustCompile(STATE_REGX)
	return re.MatchString(state)
}

func validateCity(city string) bool {
	if strings.HasPrefix(city, " ") || strings.HasSuffix(city, " ") {
		return false
	}
	re := regexp.MustCompile(CITY_NAME_REGX)
	return re.MatchString(city)
}

func validateZipCode(zipCode string) bool {
	re := regexp.MustCompile(ZIP_CODE_REGX)
	return re.MatchString(zipCode)
}

func validateLedgerDOB(dob string) bool {
	_, err := time.Parse("20060102", dob)
	return err == nil
}

func validateDOB(dobStr string) bool {
	dob, err := time.Parse("01/02/2006", dobStr)
	if err != nil {
		return false
	}
	now := clock.Now().UTC()
	age := now.Year() - dob.Year()
	if now.YearDay() < dob.YearDay() {
		age--
	}
	if age > 200 {
		return false
	}

	return age >= 18
}

func validateTransactionNote(note string) bool {
	if strings.HasPrefix(note, " ") || strings.HasSuffix(note, " ") {
		return false
	}
	if note == "" {
		return true
	}
	re := regexp.MustCompile(TRANSACTION_NOTES_REGX)
	return re.MatchString(note)
}
