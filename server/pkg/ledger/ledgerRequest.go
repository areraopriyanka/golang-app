package ledger

import (
	"encoding/json"
	"process-api/pkg/config"
	"process-api/pkg/logging"

	"braces.dev/errtrace"
)

type Request struct {
	Method string `json:"method"`
	Id     string `json:"id"`
	Params Params `json:"params"`
}

type Params struct {
	Api     Api             `json:"api"`
	Payload json.RawMessage `json:"payload"`
}

type Api struct {
	Signature  string `json:"signature"`
	KeyId      string `json:"keyId"`
	ApiKey     string `json:"apiKey"`
	Credential string `json:"credential"`
	Mfp        string `json:"mfp,omitempty"`
}

var (
	middlewareKeyId      string
	middlewareApiKey     string
	middlewareCredential string
)

func (req *Request) setApiParams(signature string) {
	middlewareKeyId = config.Config.Ledger.KeyId
	middlewareApiKey = config.Config.Ledger.ApiKey
	middlewareCredential = config.Config.Ledger.Credential

	req.Params.Api.Signature = signature
	req.Params.Api.KeyId = middlewareKeyId
	if req.Params.Api.ApiKey == "" {
		req.Params.Api.ApiKey = middlewareApiKey
	}
	if req.Params.Api.Credential == "" {
		req.Params.Api.Credential = middlewareCredential
	}
}

func SignPayload(payload json.RawMessage) (string, error) {
	payloadByteArr, err := json.Marshal(payload)
	if err != nil {
		return "", errtrace.Wrap(err)
	}
	return SignWithLedgerKey(payloadByteArr)
}

func (request *Request) SignRequestWithMiddlewareKey() error {
	logging.Logger.Info("Inside SignRequestWithMiddlewareKey for Ledger API", "requestMethod", request.Method)

	// Add signing payload part and set api struct values
	signature, err := SignPayload(request.Params.Payload)
	if err != nil {
		logging.Logger.Error("An error occurred while signing unsigned mobile request payload", "error", err)
		return errtrace.Wrap(err)
	}
	logging.Logger.Info("Signature Created Successfully.")
	request.setApiParams(signature)

	return nil
}
