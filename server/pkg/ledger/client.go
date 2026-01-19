package ledger

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/crypto"

	"braces.dev/errtrace"
)

type NetXDApiClient struct {
	paramsBuilder ParamsBuilder
	url           string
}

func (apiClient *NetXDApiClient) BuildParams(payload interface{}) (*Params, error) {
	if apiClient == nil {
		return nil, fmt.Errorf("apiClient is nil for BuildParams call")
	}

	return apiClient.paramsBuilder.BuildParams(payload)
}

type ParamsBuilder interface {
	BuildParams(payload interface{}) (*Params, error)
}

type SigningParamsBuilder struct {
	privateKey string
	credential string
	keyId      string
	apiKey     string
}

func NewSigningParamsBuilder(privateKey string, username string, password string, keyId string, apiKey string) *SigningParamsBuilder {
	userCredential := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
	return &SigningParamsBuilder{
		privateKey: privateKey,
		credential: userCredential,
		keyId:      keyId,
		apiKey:     apiKey,
	}
}

func NewLedgerSigningParamsBuilderFromConfig(config config.LedgerConfigs) *SigningParamsBuilder {
	return &SigningParamsBuilder{
		privateKey: config.PrivateKey,
		credential: config.Credential,
		keyId:      config.KeyId,
		apiKey:     config.ApiKey,
	}
}

func (c *SigningParamsBuilder) BuildParams(payload interface{}) (*Params, error) {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("marshaling payload: %w", err))
	}

	signature, err := crypto.Sign(rawPayload, c.privateKey)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("signing payload: %w", err))
	}

	api := Api{
		KeyId:      c.keyId,
		ApiKey:     c.apiKey,
		Signature:  signature,
		Credential: c.credential,
	}

	return &Params{
		Api:     api,
		Payload: rawPayload,
	}, nil
}

type PreSignedParamsBuilder struct {
	publicKey  string
	signature  string
	rawPayload string
	username   string
	password   string
	keyId      string
	apiKey     string
}

func NewPreSignedParamsBuilder(publicKey string, signature string, rawPayload string, username string, password string, keyId string, apiKey string) *PreSignedParamsBuilder {
	return &PreSignedParamsBuilder{
		publicKey:  publicKey,
		signature:  signature,
		rawPayload: rawPayload,
		username:   username,
		password:   password,
		keyId:      keyId,
		apiKey:     apiKey,
	}
}

func (c *PreSignedParamsBuilder) BuildParams(payload interface{}) (*Params, error) {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to marshal payload: %s", err.Error()))
	}

	if c.rawPayload != string(rawPayload) {
		return nil, errtrace.Wrap(fmt.Errorf("rawPayload did not equal the marshalled payload: %s %s", c.rawPayload, string(rawPayload)))
	}

	api := Api{
		KeyId:      c.keyId,
		ApiKey:     c.apiKey,
		Signature:  c.signature,
		Credential: "Basic " + base64.StdEncoding.EncodeToString([]byte(c.username+":"+c.password)),
	}

	return &Params{
		Api:     api,
		Payload: json.RawMessage(c.rawPayload),
	}, nil
}

type NetXDCardApiConfig struct {
	cardProduct string
	cardChannel string
	cardProgram string
	publicKey   string
}

func NewNetXDCardApiConfig(config config.LedgerConfigs) *NetXDCardApiConfig {
	return &NetXDCardApiConfig{
		cardProduct: config.CardsProduct,
		cardChannel: config.CardsChannel,
		cardProgram: config.CardsProgram,
		publicKey:   config.CardsPublicKey,
	}
}

type NetXDCardApiClient struct {
	NetXDApiClient
	NetXDCardApiConfig
}

func NewNetXDCardApiClient(config config.LedgerConfigs, paramsBuilder ParamsBuilder) *NetXDCardApiClient {
	return &NetXDCardApiClient{
		NetXDApiClient: NetXDApiClient{
			url:           config.CardsEndpoint,
			paramsBuilder: paramsBuilder,
		},
		NetXDCardApiConfig: *NewNetXDCardApiConfig(config),
	}
}

type NetXDLedgerApiClient struct {
	NetXDApiClient
	ledgerCategory string
}

func NewNetXDLedgerApiClient(config config.LedgerConfigs, paramsBuilder ParamsBuilder) *NetXDLedgerApiClient {
	return &NetXDLedgerApiClient{
		NetXDApiClient{
			url:           config.Endpoint,
			paramsBuilder: paramsBuilder,
		},
		config.LedgerCategory,
	}
}

type NetXDPaymentApiClient struct {
	NetXDApiClient
}

func NewNetXDPaymentApiClient(config config.LedgerConfigs, paramsBuilder ParamsBuilder) *NetXDPaymentApiClient {
	return &NetXDPaymentApiClient{NetXDApiClient{
		url:           config.PaymentsEndpoint,
		paramsBuilder: paramsBuilder,
	}}
}

type NetXDApiResponse[T any] struct {
	Result *T               `json:"result"`
	Error  *MaybeInnerError `json:"error"`
	Id     string           `json:"id"`
}

func (c *NetXDApiClient) call(method string, endpoint string, payload interface{}, response interface{}) error {
	params, err := c.BuildParams(payload)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("building params: %w", err))
	}
	request := &Request{
		Id:     constant.LEDGER_REQUEST_ID,
		Method: method,
		Params: *params,
	}

	statusCode, respBody, err := CallLedgerAPIWithUrlAndGetRawResponse(nil, request, endpoint)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("calling ledger API: %w", err))
	}

	if statusCode != http.StatusOK {
		return errtrace.Wrap(fmt.Errorf("ledger returned error status %d: %s", statusCode, respBody))
	}

	err = json.Unmarshal(respBody, response)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed parsing body: %w", err))
	}

	return nil
}

func CreateLedgerApiClient(ledgerConfig config.LedgerConfigs) *NetXDLedgerApiClient {
	ledgerParamsBuilder := NewLedgerSigningParamsBuilderFromConfig(ledgerConfig)
	ledgerClient := NewNetXDLedgerApiClient(ledgerConfig, ledgerParamsBuilder)
	return ledgerClient
}

func CreateCardApiClient(publicKey, signature, payload, email, password, keyId, apiKey string) *NetXDCardApiClient {
	userParamsBuilder := NewPreSignedParamsBuilder(publicKey, signature, payload, email, password, keyId, apiKey)
	userClient := NewNetXDCardApiClient(config.Config.Ledger, userParamsBuilder)
	return userClient
}

func CreatePaymentApiClient(publicKey, signature, payload, email, password, keyId, apiKey string) *NetXDPaymentApiClient {
	userParamsBuilder := NewPreSignedParamsBuilder(publicKey, signature, payload, email, password, keyId, apiKey)
	userClient := NewNetXDPaymentApiClient(config.Config.Ledger, userParamsBuilder)
	return userClient
}
