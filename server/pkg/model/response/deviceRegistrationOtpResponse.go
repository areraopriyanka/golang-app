package response

type ChallengeDeviceRegistrationOtpResponse struct {
	KeyId string `json:"keyId" validate:"required" mask:"true"`
}
