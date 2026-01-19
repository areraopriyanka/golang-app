package dao

import (
	"errors"
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/model/response"
	"regexp"
	"strings"
	"time"

	"braces.dev/errtrace"
	"github.com/jinzhu/gorm"
)

// define model for user
type MasterUserRecordDao struct {
	Id                         string    `json:"id" gorm:"column:id;primaryKey"`
	FirstName                  string    `json:"firstName" gorm:"column:first_name"`
	LastName                   string    `json:"lastName" gorm:"column:last_name"`
	Suffix                     string    `json:"suffix"`
	Email                      string    `json:"email" mask:"true"`
	UserStatus                 string    `json:"userStatus" gorm:"column:user_status"`
	MobileNo                   string    `json:"mobileNo" gorm:"column:mobile_no" mask:"true"`
	DOB                        time.Time `json:"dob" mask:"true"`
	Password                   []byte    `json:"password" mask:"true"`
	StreetAddress              string    `json:"streetAddress" gorm:"column:street_address" mask:"true"`
	ApartmentNo                string    `json:"apartmentNo" gorm:"column:apartment_no"`
	ZipCode                    string    `json:"zipCode" gorm:"column:zip_code"`
	City                       string    `json:"city"`
	State                      string    `json:"state"`
	LedgerPassword             []byte    `json:"ledgerPassword" gorm:"column:ledger_password"`
	KmsEncryptedLedgerPassword []byte    `json:"kmsEncryptedLedgerPassword" gorm:"column:kms_encrypted_ledger_password"`
	LedgerCustomerNumber       string    `json:"ledgerCustomerNumber" gorm:"column:ledger_customer_number" mask:"true"`
	ResetToken                 string    `json:"resetToken" gorm:"column:reset_token" mask:"true"`
	DebtwiseOnboardingStatus   string    `json:"debtwiseOnboardingStatus" gorm:"column:debtwise_onboarding_status;default:'uninitialized'"`
	DebtwiseCustomerNumber     *int      `json:"debtwiseCustomerNumber" gorm:"column:debtwise_customer_number"`
	CreatedAt                  time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt                  time.Time `gorm:"column:updated_at;autoUpdateTime"`
	// Agreement hash values
	AgreementPrivacyNoticeHash               *string
	AgreementCardAndDepositHash              *string
	AgreementDreamfiAchAuthorizationHash     *string
	AgreementESignHash                       *string
	AgreementTermsOfServiceHash              *string
	AgreementPrivacyNoticeSignedAt           *time.Time
	AgreementCardAndDepositSignedAt          *time.Time
	AgreementDreamfiAchAuthorizationSignedAt *time.Time
	AgreementESignSignedAt                   *time.Time
	AgreementTermsOfServiceSignedAt          *time.Time
}

func (m MasterUserRecordDao) FullName() string {
	return fmt.Sprintf("%s %s", m.FirstName, m.LastName)
}

// TableName overrides the table name
func (MasterUserRecordDao) TableName() string {
	return "master_user_records"
}

func (MasterUserRecordDao) FindUserByEmail(email string) (*MasterUserRecordDao, error) {
	var user MasterUserRecordDao
	result := db.DB.Where("LOWER(email) = LOWER(?)", email).Take(&user)
	// user not found
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, errtrace.Wrap(result.Error)
	}
	return &user, nil
}

func (MasterUserRecordDao) FindOneByUserId(userId string) (*MasterUserRecordDao, error) {
	var user MasterUserRecordDao
	result := db.DB.Where("id= ?", userId).Take(&user)
	// user not found
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, errtrace.Wrap(result.Error)
	}
	return &user, nil
}

func (MasterUserRecordDao) FindUserByMobileNumber(mobileNo string) (*MasterUserRecordDao, error) {
	var user MasterUserRecordDao
	mobileNumber := "+1" + mobileNo
	result := db.DB.Where("mobile_no = ?", mobileNumber).Take(&user)
	// user not found
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, errtrace.Wrap(result.Error)
	}
	return &user, nil
}

func (MasterUserRecordDao) FindAll() ([]MasterUserRecordDao, error) {
	var users []MasterUserRecordDao
	err := db.DB.Find(&users).Error
	return users, errtrace.Wrap(err)
}

func RequireUserWithState(userId string, userStates ...string) (*MasterUserRecordDao, *response.ErrorResponse) {
	user, err := MasterUserRecordDao{}.FindOneByUserId(userId)
	if err != nil {
		return nil, &response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("DB Error: %s", err), MaybeInnerError: errtrace.Wrap(err)}
	}
	if user == nil {
		return nil, &response.ErrorResponse{ErrorCode: constant.USER_NOT_FOUND, StatusCode: http.StatusNotFound, LogMessage: "User record not found in DB", MaybeInnerError: errtrace.New("")}
	}
	activeUserRequired := false

	for _, state := range userStates {
		if state == "ACTIVE" {
			activeUserRequired = true
		}

		if user.UserStatus == state {
			return user, nil
		}
	}

	// TODO: This is kept for backward compatibility with the LoginScreen.tsx file in the frontend.
	// It should be removed once we switch to using INVALID_USER_STATE or handle HTTP status PreconditionFailed on the frontend.
	if activeUserRequired {
		return user, &response.ErrorResponse{
			ErrorCode:       constant.USER_NOT_ACTIVE,
			Message:         constant.USER_NOT_ACTIVE_MSG,
			StatusCode:      http.StatusPreconditionFailed,
			LogMessage:      fmt.Sprintf("Expected user to be 'ACTIVE', but found '%s'", user.UserStatus),
			MaybeInnerError: errtrace.New(""),
		}
	}
	return user, &response.ErrorResponse{
		ErrorCode:       constant.INVALID_USER_STATE,
		StatusCode:      http.StatusPreconditionFailed,
		Message:         constant.INVALID_USER_STATE_MSG,
		LogMessage:      fmt.Sprintf("User's state is invalid. Current status: '%s'", user.UserStatus),
		MaybeInnerError: errtrace.New(""),
	}
}

type SearchTerm struct {
	Value     string
	FieldHint SearchFieldHint
}

type SearchFieldHint int

const (
	FieldHintAlpha SearchFieldHint = iota
	FieldHintEmail
	FieldHintPhone
	FieldHintNumber
	FieldHintGUID
	FieldHintReferenceId
	FieldHintUnknown
)

func parseSearchTerms(searchTerm string) []SearchTerm {
	var terms []SearchTerm
	remaining := strings.TrimSpace(searchTerm)

	emailRegex := regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`)
	for _, email := range emailRegex.FindAllString(remaining, -1) {
		terms = append(terms, SearchTerm{
			Value:     strings.ToLower(email),
			FieldHint: FieldHintEmail,
		})
		remaining = strings.ReplaceAll(remaining, email, " ")
	}

	phoneRegex := regexp.MustCompile(`\b(?:\+?1[-.\s]?)?\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})\b`)
	for _, phone := range phoneRegex.FindAllString(remaining, -1) {
		// Extract just digits (and +)
		normalized := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")
		if len(normalized) >= 10 {
			terms = append(terms, SearchTerm{
				Value:     normalized,
				FieldHint: FieldHintPhone,
			})
			remaining = strings.ReplaceAll(remaining, phone, " ")
		}
	}

	guidRegex := regexp.MustCompile(`\b[a-fA-F\d]{8}-?[a-fA-F\d]{4}-?[a-fA-F\d]{4}-?[a-fA-F\d]{4}-?[a-fA-F\d]{12}\b`)
	guids := guidRegex.FindAllString(remaining, -1)
	for _, word := range guids {
		terms = append(terms, SearchTerm{
			Value:     strings.ToLower(word),
			FieldHint: FieldHintGUID,
		})
		remaining = strings.ReplaceAll(remaining, word, " ")
	}

	referenceIdRegex := regexp.MustCompile(`\bledger\.[a-z0-9]+\.[a-z_]+_[0-9]{19,}\b`)
	referenceIds := referenceIdRegex.FindAllString(remaining, -1)
	for _, word := range referenceIds {
		terms = append(terms, SearchTerm{
			Value:     strings.ToLower(word),
			FieldHint: FieldHintReferenceId,
		})
		remaining = strings.ReplaceAll(remaining, word, " ")
	}

	alphaRegex := regexp.MustCompile(`\b[a-zA-Z]+\b`)
	alphas := alphaRegex.FindAllString(remaining, -1)
	for _, word := range alphas {
		if len(word) > 1 {
			terms = append(terms, SearchTerm{
				Value:     strings.ToLower(word),
				FieldHint: FieldHintAlpha,
			})
			remaining = strings.ReplaceAll(remaining, word, " ")
		}
	}

	digitRegex := regexp.MustCompile(`\b[0-9]+\b`)
	digits := digitRegex.FindAllString(remaining, -1)
	for _, word := range digits {
		if len(word) > 1 { // Skip single characters
			terms = append(terms, SearchTerm{
				Value:     strings.ToLower(word),
				FieldHint: FieldHintNumber,
			})
			remaining = strings.ReplaceAll(remaining, word, " ")
		}
	}

	unknownRegex := regexp.MustCompile(`[a-zA-Z0-9_\.\-]+`)
	unknowns := unknownRegex.FindAllString(remaining, -1)
	for _, word := range unknowns {
		if len(word) > 1 { // Skip single characters
			terms = append(terms, SearchTerm{
				Value:     strings.ToLower(word),
				FieldHint: FieldHintUnknown,
			})
			remaining = strings.ReplaceAll(remaining, word, " ")
		}
	}

	return terms
}

func (MasterUserRecordDao) SearchUsers(searchTerm string, userStatus string, limit int, offset int) ([]MasterUserRecordDao, int64, error) {
	var users []MasterUserRecordDao
	var totalCount int64

	query := db.DB.Model(&MasterUserRecordDao{})

	if userStatus != "" && userStatus != "ALL" {
		query = query.Where("user_status = ?", userStatus)
	}

	if searchTerm != "" {
		searchTerms := parseSearchTerms(searchTerm)

		if len(searchTerms) > 0 {
			var conditions []string
			var args []any

			for _, term := range searchTerms {
				searchPattern := "%" + term.Value + "%"

				switch term.FieldHint {
				case FieldHintEmail:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"email",
					)
				case FieldHintPhone:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"mobile_no",
					)
				case FieldHintAlpha:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"email",
						"first_name",
						"last_name",
					)
				case FieldHintNumber:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"ledger_customer_number",
						"mobile_no",
					)
				case FieldHintGUID:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"id::text",
					)
				case FieldHintUnknown:
					conditions, args = addSearchTerm(conditions, args, searchPattern,
						"email",
						"first_name",
						"id::text",
						"last_name",
						"ledger_customer_number",
						"mobile_no",
					)
				}
			}

			// Join all conditions with AND
			whereClause := strings.Join(conditions, " AND ")
			query = query.Where(whereClause, args...)
		}
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, errtrace.Wrap(err)
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, errtrace.Wrap(err)
	}

	return users, totalCount, nil
}

func addSearchTerm(conditions []string, args []any, searchPattern string, searchColumns ...string) ([]string, []any) {
	var searchConditions []string

	for _, searchColumn := range searchColumns {
		searchConditions = append(searchConditions, fmt.Sprintf("%s ILIKE ?", searchColumn))
		args = append(args, searchPattern)
	}

	conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(searchConditions, " OR ")))
	return conditions, args
}
