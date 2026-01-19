package request

type SardineKycRequest struct {
	SSN               string `json:"ssn" validate:"required,len=9" mask:"true"` // SSN must be exactly 9 characters
	SardineSessionKey string `json:"sardineSessionKey" validate:"required" mask:"true"`
}
