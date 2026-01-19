package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"log/slog"
	"process-api/pkg/config"
	"process-api/pkg/logging"

	"braces.dev/errtrace"
)

func Decrypt(cipherTextBytes []byte) (string, error) {
	encryptionKey, err := DecryptKms(config.Config.Encrypt.EncryptionKey)
	if err != nil {
		return "", errtrace.Wrap(err)
	}
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return "", errtrace.Wrap(err)
	}
	if len(cipherTextBytes) < aes.BlockSize {
		return "", errtrace.Wrap(fmt.Errorf("cipherText too short"))
	}

	iv := cipherTextBytes[:aes.BlockSize]
	cipherTextBytes = cipherTextBytes[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(cipherTextBytes, cipherTextBytes)

	return string(cipherTextBytes), nil
}

// DecryptApiKeyAndLedgerPassword provides backward compatibility during migration from AES to KMS encryption
// It tries KMS decryption first, then falls back to the old AES method
// TODO: Remove after 2025-10-01 (after ledgerPasswords and apiKeys have been migrated)
func DecryptApiKeyAndLedgerPassword(oldPassword, kmsEncryptedPassword, oldApiKey, kmsEncryptedApiKey []byte, logger *slog.Logger) (string, string, error) {
	var decryptedLedgerPassword, decryptedApiKey string
	var err error
	if len(kmsEncryptedPassword) == 0 {
		// Use old ledgerPassword value for decryption as kmsEncryptedPassword is not set
		decryptedLedgerPassword, err = Decrypt(oldPassword)
		if err != nil {
			return "", "", errtrace.Wrap(fmt.Errorf("error decrypting stored ledgerPassword: %s", err.Error()))
		}
		logger.Warn("Decrypted old ledgerPassword as kmsEncryptedLedgerPassword is not set")

	} else {
		decryptedLedgerPassword, err = DecryptKmsBinary(kmsEncryptedPassword)
		if err != nil {
			return "", "", errtrace.Wrap(fmt.Errorf("error decrypting stored kmsEncryptedLedgerPassword: %s", err.Error()))
		}
		logger.Debug("Decrypted kmsEncryptedLedgerPassword using kmsDecryption")
	}

	if len(kmsEncryptedApiKey) == 0 {
		// Use old apiKey value for decryption as kmsEncryptedApiKey is not set
		decryptedApiKey, err = Decrypt(oldApiKey)
		if err != nil {
			return "", "", errtrace.Wrap(fmt.Errorf("error decrypting stored apiKey: %s", err.Error()))
		}

		logger.Warn("Decrypted old apiKey as kmsEncryptedApiKey is not set")

	} else {
		decryptedApiKey, err = DecryptKmsBinary(kmsEncryptedApiKey)
		if err != nil {
			return "", "", errtrace.Wrap(fmt.Errorf("error decrypting stored apiKey: %s", err.Error()))
		}
		logger.Debug("Decrypted kmsEncryptedApiKey using kmsDecryption")
	}

	return decryptedLedgerPassword, decryptedApiKey, nil
}

// DecryptPlaidAccessToken provides backward compatibility during migration from AES to KMS encryption
// It tries KMS decryption first, then falls back to the old AES method
// TODO: Remove after 2025-10-01 (after access tokens have been migrated)
func DecryptPlaidAccessToken(oldEncryptedToken, kmsEncryptedAccessToken []byte) (string, error) {
	var decryptedAccessToken string
	var err error
	if len(kmsEncryptedAccessToken) == 0 {
		// Use old encryptedAccessToken value for decryption as kmsEncryptedAccessToken is not set
		decryptedAccessToken, err = Decrypt(oldEncryptedToken)
		if err != nil {
			return "", errtrace.Wrap(fmt.Errorf("error decrypting stored accessToken: %s", err.Error()))
		}

		logging.Logger.Warn("Decrypted old accessToken as kmsEncryptedAccessToken is not set")

	} else {
		decryptedAccessToken, err = DecryptKmsBinary(kmsEncryptedAccessToken)
		if err != nil {
			return "", errtrace.Wrap(fmt.Errorf("error decrypting stored accessToken: %s", err.Error()))
		}
		logging.Logger.Debug("Decrypted kmsEncryptedAccessToken using kmsDecryption")
	}
	return decryptedAccessToken, nil
}
