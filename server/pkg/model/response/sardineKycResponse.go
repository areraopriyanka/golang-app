package response

type SardinePayload struct {
	Result *KycResult `json:"result,omitempty"`
	Error  *KycError  `json:"error,omitempty"`
}

type KycResult struct {
	UserId    string `json:"userId"`
	KycStatus string `json:"kycStatus"`
}

type KycError struct {
	Message string `json:"message"`
}
