package ledger

import (
	"encoding/json"
	"process-api/pkg/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestPayload struct {
	Value   string `json:"value"`
	Another string `json:"another"`
}

func TestNewSigningParamsBuilder(t *testing.T) {
	publicKey, privateKey, err := crypto.CreateKeys()
	assert.NoError(t, err)

	payload := TestPayload{
		Value:   "test",
		Another: "foobar",
	}

	paramsBuilder := NewSigningParamsBuilder(privateKey, "username", "password", "1234", "apikey1234")
	params, err := paramsBuilder.BuildParams(payload)
	assert.NoError(t, err)

	rawPayload, err := json.Marshal(payload)
	assert.NoError(t, err)

	valid, err := crypto.Verify(rawPayload, publicKey, params.Api.Signature)
	assert.NoError(t, err)

	assert.True(t, valid, "signature was not valid")
}

func TestNewPreSignedParamsBuilder(t *testing.T) {
	publicKey, privateKey, err := crypto.CreateKeys()
	assert.NoError(t, err)

	payload := TestPayload{
		Value:   "test",
		Another: "foobar",
	}

	rawPayload, err := json.Marshal(payload)
	assert.NoError(t, err)

	signature, err := crypto.Sign(rawPayload, privateKey)
	assert.NoError(t, err)

	paramsBuilder := NewPreSignedParamsBuilder(publicKey, signature, string(rawPayload), "username", "password", "1234", "apikey1234")
	params, err := paramsBuilder.BuildParams(payload)
	if !assert.NoError(t, err) {
		return
	}

	valid, err := crypto.Verify(rawPayload, publicKey, params.Api.Signature)
	assert.NoError(t, err)

	assert.True(t, valid, "signature was not valid")
}
