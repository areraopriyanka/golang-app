package request

type DemographicUpdateSendOtpRequest struct {
	Type string `json:"type" validate:"required,otpType"`
}
