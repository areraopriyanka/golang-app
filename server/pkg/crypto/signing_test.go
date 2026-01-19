package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorrectPEMPrivateKeyFormatNoop(t *testing.T) {
	_, privateKey, err := CreateKeys()
	assert.NoError(t, err)

	result := CorrectPEMPrivateKeyFormat(privateKey)
	assert.Equal(t, privateKey, result)
}

func TestCorrectPEMPublicKeyFormatNoop(t *testing.T) {
	publicKey, _, err := CreateKeys()
	assert.NoError(t, err)

	result := CorrectPEMPublicKeyFormat(publicKey)
	assert.Equal(t, publicKey, result)
}

func TestLedgerSigning(t *testing.T) {
	publicKey, privateKey, err := CreateKeys()
	assert.NoError(t, err)

	rawPayload := []byte("asdf1234")

	signature, err := Sign(rawPayload, privateKey)
	assert.NoError(t, err)

	valid, err := Verify(rawPayload, publicKey, signature)
	assert.NoError(t, err)

	assert.True(t, valid, "signature was not valid")
}

func TestEcdsaSigning(t *testing.T) {
	publicKey, privateKey, err := CreateKeys()
	assert.NoError(t, err)

	rawPayload := []byte("asdf1234")

	signature, err := SignECDSA(rawPayload, privateKey)
	assert.NoError(t, err)

	valid, err := VerifyECDSA(rawPayload, publicKey, signature)
	assert.NoError(t, err)

	assert.True(t, valid, "signature was not valid")
}

// The mobile native java and swift implementation both use standard ASN1 function calls
// This asserts that the irregular implementation of Verify/Sign matches the ANS1 standard
// implementations of verifyECDSA/signECDSA
func TestSigningImplementationEquivalence(t *testing.T) {
	publicKey, privateKey, err := CreateKeys()
	assert.NoError(t, err)

	rawPayload := []byte("asdf1234")

	signature, err := SignECDSA(rawPayload, privateKey)
	assert.NoError(t, err)

	valid, err := Verify(rawPayload, publicKey, signature)
	assert.NoError(t, err)

	assert.True(t, valid, "signature was not valid")

	signature, err = Sign(rawPayload, privateKey)
	assert.NoError(t, err)

	valid, err = VerifyECDSA(rawPayload, publicKey, signature)
	assert.NoError(t, err)

	assert.True(t, valid, "signature was not valid")
}
