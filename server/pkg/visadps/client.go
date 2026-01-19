package visadps

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"process-api/pkg/clock"
	"process-api/pkg/logging"

	"braces.dev/errtrace"
	"github.com/go-jose/go-jose/v4"
)

type VisaDPSClient struct {
	*Client
	secret VisaDpsSecret
}

func (client *VisaDPSClient) GenerateCvv2(cardExternalId, last4PrimaryAccountNumber, expirationMM, expirationYY string) (string, error) {
	body := GenerateCvv2ForCardAliasId2JSONRequestBody{
		Last4PrimaryAccountNumber: last4PrimaryAccountNumber,
		ExpirationDate: ExpirationDate{
			Mm: expirationMM,
			Yy: expirationYY,
		},
	}

	cvv2response, err := mleRequest[CommonResponseGenerateCvv2Response](client, body, func(bodyReader io.Reader) (*http.Response, error) {
		return client.GenerateCvv2ForCardAliasId2WithBody(context.Background(), cardExternalId, nil, "application/json", bodyReader)
	})
	if err != nil {
		return "", errtrace.Wrap(fmt.Errorf("failed to request cvv2generation: %w", err))
	}

	return cvv2response.Resource.Cvv2, nil
}

func mleRequest[T any](client *VisaDPSClient, requestBody interface{}, performRequest func(body io.Reader) (*http.Response, error)) (*T, error) {
	reqJson, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}
	logging.Logger.Info("mleRequest", "request", string(reqJson))
	encryptedPayload, err := createEncryptedPayload(client.secret, requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt request payload: %w", err)
	}

	encryptedBody := MlePayload{
		EncData: encryptedPayload,
	}

	buf, err := json.Marshal(encryptedBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encode encrypted payload: %w", err)
	}
	bodyReader := bytes.NewReader(buf)

	response, err := performRequest(bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to make visa dps MLE request: %w", err)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	defer func() { _ = response.Body.Close() }()
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if response.StatusCode >= 300 {
		return nil, fmt.Errorf("failure status code %d from API: %s", response.StatusCode, string(bodyBytes))
	}

	// VisaDPS requires Message Level Encryption
	// This means the request payload is encrypted with a Visa public key
	// and the response payload is encrypted with our public key. That key
	// exchange occurs in the VisaDPS developer console.
	// https://developer.visa.com/pages/encryption_guide
	var encryptedResponse MlePayload
	err = json.Unmarshal(bodyBytes, &encryptedResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal EncData: %w", err)
	}

	decrypted, err := decryptPayload(client.secret, encryptedResponse.EncData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt response: %w", err)
	}
	var val T
	if err := json.Unmarshal([]byte(decrypted), &val); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response into %T: %w", val, err)
	}
	return &val, nil
}

// createEncryptedPayload creates a JWE encrypted payload using RSA-OAEP-256 and A128GCM
func createEncryptedPayload(visaDpsSecret VisaDpsSecret, payload interface{}) (string, error) {
	var payloadBytes []byte
	switch p := payload.(type) {
	case string:
		payloadBytes = []byte(p)
	default:
		var err error
		payloadBytes, err = json.Marshal(p)
		if err != nil {
			return "", fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	serverCertificate := []byte(visaDpsSecret.MleServerPublicKey)

	serverCertificateBlock, _ := pem.Decode(serverCertificate)
	if serverCertificateBlock == nil {
		return "", fmt.Errorf("failed to decode server certificate")
	}

	certificate, err := x509.ParseCertificate(serverCertificateBlock.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse server certificate: %w", err)
	}

	encrypter, err := jose.NewEncrypter(
		jose.A128GCM,
		jose.Recipient{
			Algorithm: jose.RSA_OAEP_256,
			Key:       certificate.PublicKey,
			KeyID:     visaDpsSecret.MleKeyId,
		},
		(&jose.EncrypterOptions{}).WithType("JWE").
			WithHeader("iat", clock.Now().Unix()*1000),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create encrypter: %w", err)
	}

	jwe, err := encrypter.Encrypt(payloadBytes)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt payload: %w", err)
	}

	serialized, err := jwe.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize JWE: %w", err)
	}

	return serialized, nil
}

// decryptPayload decrypts a JWE encrypted payload using RSA-OAEP-256 and A128GCM
func decryptPayload(visaDpsSecret VisaDpsSecret, encryptedPayload string) ([]byte, error) {
	privateKeyPEM := []byte(visaDpsSecret.MlePrivateKey)

	privateKeyDER, _ := pem.Decode(privateKeyPEM)

	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyDER.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	decKey := jose.JSONWebKey{
		Key:       privateKey,
		KeyID:     visaDpsSecret.MleKeyId,
		Algorithm: string(jose.RSA_OAEP_256),
		Use:       "enc",
	}

	jwe, err := jose.ParseEncrypted(encryptedPayload, []jose.KeyAlgorithm{jose.RSA_OAEP_256}, []jose.ContentEncryption{jose.A128GCM})
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted payload: %w", err)
	}

	decrypted, err := jwe.Decrypt(&decKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt payload: %w", err)
	}

	return decrypted, nil
}

type VisaDpsSecret struct {
	VisaUsername             string `json:"VisaUsername"`
	VisaPassword             string `json:"VisaPassword"`
	ClientCertificateFile    string `json:"ClientCertificateFile"`
	CACertificateFile        string `json:"CACertificateFile"`
	ClientCertificateKeyFile string `json:"ClientCertificateKeyFile"`
	MleKeyId                 string `json:"MleKeyId"`
	MleServerPublicKey       string `json:"MleServerPublicKey"`
	MlePrivateKey            string `json:"MlePrivateKey"`
}

func CreateClient(visaDpsSecret VisaDpsSecret) (*VisaDPSClient, error) {
	cert, err := tls.X509KeyPair([]byte(visaDpsSecret.ClientCertificateFile), []byte(visaDpsSecret.ClientCertificateKeyFile))
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to load client certificate: %w", err))
	}

	caCert := []byte(visaDpsSecret.CACertificateFile)
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, errtrace.Wrap(fmt.Errorf("failed to parse CA certificate"))
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpClient := &http.Client{
		Transport: transport,
	}

	username := visaDpsSecret.VisaUsername
	password := visaDpsSecret.VisaPassword

	authEditor := func(ctx context.Context, req *http.Request) error {
		req.SetBasicAuth(username, password)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		// directly assign `keyId` to avoid Set upper-casing the value to `KeyId`.
		// Visa expects `keyId` case sensitive
		req.Header["keyId"] = []string{visaDpsSecret.MleKeyId}
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return errtrace.Wrap(fmt.Errorf("failed to call DumpRequestOut: %w", err))
		}
		logging.Logger.Info("mleRequest dump", "dump", string(dump))

		return nil
	}

	client, err := NewClient(
		"https://cert.api.visa.com",
		WithHTTPClient(httpClient),
		WithRequestEditorFn(authEditor),
	)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to create Visa DPS client: %w", err))
	}

	return &VisaDPSClient{
		Client: client,
		secret: visaDpsSecret,
	}, nil
}

type MlePayload struct {
	EncData string `json:"encData"`
}
