package request

type InitiateResetPasswordAndSendEmailRequest struct {
	Username string `json:"username"`
	IsRetry  bool   `json:"isRetry"`
	Mfp      string `json:"mfp" mask:"true"`
}
