package agreements

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

var Agreements *AgreementStore

func PrepareAgreements() error {
	esign, err := prepareAgreement("e-sign")
	if err != nil {
		return err
	}

	dreamfi_ach_authorization, err := prepareAgreement("dreamfi-ach-authorization")
	if err != nil {
		return err
	}

	card_and_deposit, err := prepareAgreement("card-and-deposit")
	if err != nil {
		return err
	}

	privacy_notice, err := prepareAgreement("privacy-notice")
	if err != nil {
		return err
	}

	terms_of_service, err := prepareAgreement("terms-of-service")
	if err != nil {
		return err
	}

	debtwise, err := prepareAgreement("debtwise")
	if err != nil {
		return err
	}

	Agreements = &AgreementStore{
		ESign:                   *esign,
		DreamfiAchAuthorization: *dreamfi_ach_authorization,
		CardAndDeposit:          *card_and_deposit,
		PrivacyNotice:           *privacy_notice,
		TermsOfService:          *terms_of_service,
		Debtwise:                *debtwise,
	}

	return nil
}

func prepareAgreement(agreementName string) (*AgreementStoreEntry, error) {
	data, err := readAgreementJson(agreementName)
	if err != nil {
		return nil, err
	}

	if !json.Valid(data) {
		return nil, fmt.Errorf("agreement '%s' has invalid JSON encoding", agreementName)
	}

	buffer := bytes.Buffer{}
	err = json.Compact(&buffer, data)
	if err != nil {
		return nil, fmt.Errorf("agreement '%s' failed to compact JSON: %w", agreementName, err)
	}
	compacted_data := buffer.Bytes()

	hash, err := GetJSONHash(compacted_data)
	if err != nil {
		return nil, fmt.Errorf("agreement '%s' failed to hash JSON: %w", agreementName, err)
	}

	return &AgreementStoreEntry{
		Name: agreementName,
		JSON: string(compacted_data),
		Hash: *hash,
	}, nil
}

func readAgreementJson(agreementName string) ([]byte, error) {
	return os.ReadFile(fmt.Sprintf("pkg/resource/agreements/%s.json", agreementName))
}

func GetJSONHash(data []byte) (*string, error) {
	hash := sha256.New()
	_, err := hash.Write(data)
	if err != nil {
		return nil, err
	}
	result := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	return &result, nil
}

func GetAgreementStore(agreementName string) (*AgreementStoreEntry, error) {
	switch agreementName {
	case "e-sign":
		return &Agreements.ESign, nil
	case "dreamfi-ach-authorization":
		return &Agreements.DreamfiAchAuthorization, nil
	case "card-and-deposit":
		return &Agreements.CardAndDeposit, nil
	case "privacy-notice":
		return &Agreements.PrivacyNotice, nil
	case "terms-of-service":
		return &Agreements.TermsOfService, nil
	case "debtwise":
		return &Agreements.Debtwise, nil
	default:
		return nil, fmt.Errorf("unexpected agreement name: %s", agreementName)
	}
}

type AgreementStore struct {
	ESign                   AgreementStoreEntry
	DreamfiAchAuthorization AgreementStoreEntry
	CardAndDeposit          AgreementStoreEntry
	PrivacyNotice           AgreementStoreEntry
	TermsOfService          AgreementStoreEntry
	Debtwise                AgreementStoreEntry
}

type AgreementStoreEntry struct {
	Name string
	JSON string
	Hash string
}
