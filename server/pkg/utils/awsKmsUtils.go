package utils

import (
	"encoding/base64"
	"process-api/pkg/config"
	"process-api/pkg/logging"

	"braces.dev/errtrace"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

var kmsClient kmsiface.KMSAPI

func AwsKmsConnection() {
	var awsConfig *aws.Config
	if config.Config.Aws.AccessKeyId != "" {
		awsConfig = &aws.Config{
			Region:      aws.String(config.Config.Aws.Region),
			Credentials: credentials.NewStaticCredentials(config.Config.Aws.AccessKeyId, config.Config.Aws.SecreteKeyId, ""),
		}
	} else {
		awsConfig = &aws.Config{
			Region: aws.String(config.Config.Aws.Region),
		}
	}

	// Connection for aws
	awsSession := session.Must(session.NewSession(awsConfig))
	kmsClient = kms.New(awsSession)
}

func SetKmsClient(client kmsiface.KMSAPI) {
	kmsClient = client
}

func EncryptKmsBinary(plainText string) ([]byte, error) {
	result, err := kmsClient.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String(config.Config.Aws.KmsEncryptionKeyId),
		Plaintext: []byte(plainText),
	})
	if err != nil {
		logging.Logger.Error("Got error encrypting data", "error", err)
		return []byte{}, errtrace.Wrap(err)
	}
	return result.CiphertextBlob, nil
}

func EncryptKms(plainText string) (string, error) {
	// Encrypt the data
	result, err := EncryptKmsBinary(plainText)
	if err != nil {
		return "", errtrace.Wrap(err)
	}

	return base64.URLEncoding.EncodeToString(result), nil
}

func DecryptKms(cipherTextStr string) (string, error) {
	cipherText, err := base64.URLEncoding.DecodeString(cipherTextStr)
	if err != nil {
		return "", errtrace.Wrap(err)
	}

	return DecryptKmsBinary(cipherText)
}

func DecryptKmsBinary(cipherBlob []byte) (string, error) {
	result, err := kmsClient.Decrypt(&kms.DecryptInput{CiphertextBlob: cipherBlob})
	if err != nil {
		logging.Logger.Error("Got error decrypting data", "error", err)
		return "", errtrace.Wrap(err)
	}

	return string(result.Plaintext), nil
}
