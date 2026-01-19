package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"process-api/pkg/logging"
	"strings"

	"braces.dev/errtrace"
)

type ECDSASignature struct {
	R, S *big.Int
}

func DecodePrivateKey(privateKey string) (*ecdsa.PrivateKey, error) {
	// convert private key in expected format
	bts := CorrectPEMPrivateKeyFormat(privateKey)
	block, _ := pem.Decode([]byte(bts))
	if block == nil {
		return nil, errtrace.Wrap(errors.New("failed to decode PEM private key"))
	}

	x509Encoded := block.Bytes
	parsedPrivateKey, err := x509.ParsePKCS8PrivateKey(x509Encoded)
	if err != nil {
		return nil, errtrace.Wrap(errors.New("failed to parse ECDSA private key"))
	}
	switch parsedPrivateKey := parsedPrivateKey.(type) {
	case *ecdsa.PrivateKey:
		return parsedPrivateKey, nil
	}
	return nil, errtrace.Wrap(errors.New("unsupported public key type"))
}

func DecodePublicKey(publicKey string) (*ecdsa.PublicKey, error) {
	// convert private key in expected format
	bts := CorrectPEMPublicKeyFormat(publicKey)
	block, _ := pem.Decode([]byte(bts))
	if block == nil {
		return nil, errtrace.Wrap(errors.New("failed to decode PEM public key"))
	}

	x509Encoded := block.Bytes
	parsedPublicKey, err := x509.ParsePKIXPublicKey(x509Encoded)
	if err != nil {
		return nil, errtrace.Wrap(errors.New("failed to parse ECDSA public key"))
	}
	switch parsedPublicKey := parsedPublicKey.(type) {
	case *ecdsa.PublicKey:
		return parsedPublicKey, nil
	}
	return nil, errtrace.Wrap(errors.New("unsupported public key type"))
}

func CorrectPEMPrivateKeyFormat(key string) string {
	// Remove any leading or trailing whitespace
	key = strings.TrimSpace(key)

	// Ensure the key starts with "-----BEGIN PRIVATE KEY-----"
	if !strings.HasPrefix(key, "-----BEGIN PRIVATE KEY-----") {
		key = "-----BEGIN PRIVATE KEY-----\n" + key
	}

	// Ensure the key ends with "-----END PRIVATE KEY-----"
	if !strings.HasSuffix(key, "-----END PRIVATE KEY-----") {
		key = key + "\n-----END PRIVATE KEY-----"
	}

	// Ensure there is a newline character between BEGIN and the key content
	key = strings.Replace(key, "-----BEGIN PRIVATE KEY----- ", "-----BEGIN PRIVATE KEY-----\n ", 1)

	// Ensure there is a newline character between the key content and END
	key = strings.Replace(key, " -----END PRIVATE KEY-----", " \n-----END PRIVATE KEY-----", 1)

	// Normalize line endings to use Unix-style ("\n")
	key = strings.ReplaceAll(key, "\r\n", "\n")

	return key
}

func CorrectPEMPublicKeyFormat(key string) string {
	// Remove any leading or trailing whitespace
	key = strings.TrimSpace(key)

	// Ensure the key starts with "-----BEGIN PUBLIC KEY-----"
	if !strings.HasPrefix(key, "-----BEGIN PUBLIC KEY-----") {
		key = "-----BEGIN PUBLIC KEY-----\n" + key
	}

	// Ensure the key ends with "-----END PUBLIC KEY-----"
	if !strings.HasSuffix(key, "-----END PUBLIC KEY-----") {
		key = key + "\n-----END PUBLIC KEY-----"
	}

	// Ensure there is a newline character between BEGIN and the key content
	key = strings.Replace(key, "-----BEGIN PUBLIC KEY----- ", "-----BEGIN PUBLIC KEY-----\n ", 1)

	// Ensure there is a newline character between the key content and END
	key = strings.Replace(key, " -----END PUBLIC KEY-----", " \n-----END PUBLIC KEY-----", 1)

	// Normalize line endings to use Unix-style ("\n")
	key = strings.ReplaceAll(key, "\r\n", "\n")

	return key
}

func Sign(payload []byte, privateKey string) (string, error) {
	pk, derr := DecodePrivateKey(privateKey)
	if derr != nil {
		logging.Logger.Error("Error while decode PEM private key", "error", derr.Error())
		return "", errtrace.Wrap(derr)
	}

	hash := sha256.Sum256(payload)
	r, s, err := ecdsa.Sign(rand.Reader, pk, hash[:])
	if err != nil {
		logging.Logger.Error(err.Error())
		return "", errtrace.Wrap(err)
	}
	b, ee := asn1.Marshal(ECDSASignature{r, s})
	if ee != nil {
		logging.Logger.Error(ee.Error())
		return "", errtrace.Wrap(ee)
	}
	signature := base64.StdEncoding.EncodeToString(b)
	return signature, nil
}

func Verify(payload []byte, publicKey string, signature string) (bool, error) {
	pk, err := DecodePublicKey(publicKey)
	if err != nil {
		logging.Logger.Error("Error while decode PEM private key", "error", err)
		return false, errtrace.Wrap(err)
	}

	rawSignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		logging.Logger.Error("Verify failed to decode signature", "error", err)
		return false, errtrace.Wrap(err)
	}

	hash := sha256.Sum256(payload)

	var sig ECDSASignature
	_, err = asn1.Unmarshal(rawSignature, &sig)
	if err != nil {
		logging.Logger.Error("Error while unmarshaling ASN.1 signature", "error", err)
		return false, errtrace.Wrap(err)
	}

	return ecdsa.Verify(pk, hash[:], sig.R, sig.S), nil
}

func CreateKeys() (string, string, error) {
	// > P256 returns a [Curve] which implements NIST P-256 (FIPS 186-3, section D.2.3),
	// > also known as secp256r1 or prime256v1. The CurveParams.Name of this [Curve] is "P-256".
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", errtrace.Wrap(fmt.Errorf("failed to generate private key: %v", err))
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", errtrace.Wrap(fmt.Errorf("failed to marshal private key: %v", err))
	}

	privateKeyPem := fmt.Sprintf(
		"-----BEGIN PRIVATE KEY-----\n%s\n-----END PRIVATE KEY-----",
		base64.StdEncoding.EncodeToString(privateKeyBytes),
	)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", errtrace.Wrap(fmt.Errorf("failed to marshal public key: %v", err))
	}

	publicKeyPem := fmt.Sprintf(
		"-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----",
		base64.StdEncoding.EncodeToString(publicKeyBytes),
	)

	return publicKeyPem, privateKeyPem, nil
}

func SignECDSA(payload []byte, privateKeyPEM string) (string, error) {
	privateKey, err := DecodePrivateKey(privateKeyPEM)
	if err != nil {
		return "", errtrace.Wrap(err)
	}

	hash := sha256.Sum256(payload)

	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		return "", errtrace.Wrap(err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func VerifyECDSA(payload []byte, publicKeyPEM string, signatureB64 string) (bool, error) {
	publicKey, err := DecodePublicKey(publicKeyPEM)
	if err != nil {
		return false, errtrace.Wrap(err)
	}

	hash := sha256.Sum256(payload)

	signature, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return false, errtrace.Wrap(err)
	}

	return ecdsa.VerifyASN1(publicKey, hash[:], signature), nil
}
