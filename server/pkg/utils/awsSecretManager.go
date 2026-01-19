package utils

import (
	"encoding/json"
	"fmt"
	"process-api/pkg/config"

	"braces.dev/errtrace"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

func GetSecret(region string, secretId string) (string, error) {
	// Connection for aws
	var awsConfig *aws.Config
	if config.Config.AwsSecretManager.AccessKeyId != "" {
		awsConfig = &aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewStaticCredentials(config.Config.AwsSecretManager.AccessKeyId, config.Config.AwsSecretManager.SecreteKeyId, ""),
		}
	} else {
		awsConfig = &aws.Config{
			Region: aws.String(region),
		}
	}
	awsSession := session.Must(session.NewSession(awsConfig))
	sm := secretsmanager.New(awsSession)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretId),
	}
	result, err := sm.GetSecretValue(input)
	if err != nil {
		return "", errtrace.Wrap(fmt.Errorf("error retrieving secret value from secret manager: %w", err))
	}
	return *result.SecretString, nil
}

func UnmarshalSecret[T any](region string, secretId string) (*T, error) {
	jsonString, err := GetSecret(region, secretId)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	var val T
	if err := json.Unmarshal([]byte(jsonString), &val); err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to unmarshal secret into %T: %w", val, err))
	}
	return &val, nil
}
