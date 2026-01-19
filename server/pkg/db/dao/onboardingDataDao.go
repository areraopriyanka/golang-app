package dao

import (
	"encoding/json"
	"time"
)

type OnboardingDataDao struct {
	TempCustomerNo      string          `json:"tempCustomerNo" gorm:"column:temp_customer_no" mask:"true"`
	CustomerNo          string          `json:"customerNo" gorm:"column:customer_no" mask:"true"`
	KycReferenceNo      string          `json:"kycReferenceNo" gorm:"column:kyc_reference_no"`
	KycApplicationId    string          `json:"kycApplicationId" gorm:"column:kyc_application_id"`
	KycDetails          json.RawMessage `json:"kycDetails" gorm:"column:kyc_details" mask:"true"`
	LastScreenSubmitted string          `json:"lastScreenSubmitted" gorm:"column:last_screen_submitted"`
	UserDetails         json.RawMessage `json:"userDetails" gorm:"column:user_details"`
	Extras              json.RawMessage `json:"extras" gorm:"column:extras"`
	AddCustomerResp     json.RawMessage `json:"addCustomerResp" gorm:"column:add_customer_resp" mask:"true"`
	AddCheckingAccResp  json.RawMessage `json:"addCheckingAccResp" gorm:"column:add_checking_acc_resp" mask:"true"`
	AddSavingAccResp    json.RawMessage `json:"addSavingAccResp" gorm:"column:add_saving_acc_resp" mask:"true"`
	AddCardHolderResp   json.RawMessage `json:"addCardHolderResp" gorm:"column:add_card_holder_resp" mask:"true"`
	AddCardResp         json.RawMessage `json:"addCardResp" gorm:"column:add_card_resp" mask:"true"`
	InitiateKycResp     json.RawMessage `json:"initiateKycResp" gorm:"column:initiate_kyc_resp" mask:"true"`
	GetKycDocuments     json.RawMessage `gorm:"column:get_kyc_documents"`
	EndOnboardingResp   json.RawMessage `json:"endOnboardingResp" gorm:"column:end_onboarding_resp"`
	CreatedAt           time.Time       `json:"created_at" gorm:"column:created_at"`
	CreatedBy           string          `json:"createdBy" gorm:"column:created_by"`
	UpdatedAt           time.Time       `json:"updatedAt" gorm:"column:updated_at"`
	UpdatedBy           string          `json:"updatedBy" gorm:"column:updated_by"`
}

// TableName overrides the table name used by CustomerDao to `customers`
func (OnboardingDataDao) TableName() string {
	return "onboarding_data"
}
