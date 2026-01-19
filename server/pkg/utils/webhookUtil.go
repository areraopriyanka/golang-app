package utils

import (
	"encoding/json"
	"process-api/pkg/config"
	"process-api/pkg/crypto"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
)

func VerifyLedgerWebhookSignature(event request.LedgerEventPayload) bool {
	logging.Logger.Debug("Inside verifySignature")
	// Extract payload and signature
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		logging.Logger.Error("Error marshalling payload", "error", err.Error())
	}

	verifyResult, err := crypto.VerifyECDSA(payload, config.Config.Webhook.PublicKey, event.Signature)
	if err != nil {
		logging.Logger.Error("Failed to verify ecdsa signature for webhook event payload", "error", err)
		return false
	}

	return verifyResult
}
