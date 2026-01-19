package ledger

import (
	"process-api/pkg/config"
	"process-api/pkg/crypto"
)

// SignWithLedgerKey signs payload using the Ledger service private key
func SignWithLedgerKey(payload []byte) (string, error) {
	return crypto.Sign(payload, config.Config.Ledger.PrivateKey)
}

// SignWithKycKey signs payload using the KYC service private key
func SignWithKycKey(payload []byte) (string, error) {
	return crypto.Sign(payload, config.Config.Kyc.PrivateKey)
}
