package ledger

import (
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/utils"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBuildOutboundAchCreditRequest_WithMissingReason(t *testing.T) {
	sessionId := uuid.New().String()
	userRecord := dao.MasterUserRecordDao{
		Id:                         sessionId,
		FirstName:                  "DEBTOR",
		LastName:                   "Bar",
		Email:                      "email@example.com",
		KmsEncryptedLedgerPassword: nil,
		LedgerCustomerNumber:       "100000000006001",
		Password:                   []byte("password"),
		UserStatus:                 constant.ACTIVE,
	}

	payloadData := BuildOutboundAchCreditRequest(
		&userRecord,
		"50040002699049",
		"000000000",
		"50000",
		"CREDITOR",
		"234567890123456",
		"012345678",
		"CHECKING",
		nil,
		nil,
	)

	assert.Equal(t, "ACH Credit", *payloadData.Reason)

	payloadData = BuildOutboundAchCreditRequest(
		&userRecord,
		"50040002699049",
		"000000000",
		"50000",
		"CREDITOR",
		"234567890123456",
		"012345678",
		"CHECKING",
		utils.Pointer(""),
		nil,
	)

	assert.Equal(t, "ACH Credit", *payloadData.Reason)
}

func TestBuildOutboundAchDeditRequest_WithMissingReason(t *testing.T) {
	sessionId := uuid.New().String()
	userRecord := dao.MasterUserRecordDao{
		Id:                         sessionId,
		FirstName:                  "DEBTOR",
		LastName:                   "Bar",
		Email:                      "email@example.com",
		KmsEncryptedLedgerPassword: nil,
		LedgerCustomerNumber:       "100000000006001",
		Password:                   []byte("password"),
		UserStatus:                 constant.ACTIVE,
	}

	payloadData := BuildOutboundAchDebitRequest(
		&userRecord,
		"50040002699049",
		"50000",
		"DEBTOR",
		"234567890123456",
		"012345678",
		"CHECKING",
		nil,
		nil,
	)

	assert.Equal(t, "ACH Debit", *payloadData.Reason)

	payloadData = BuildOutboundAchDebitRequest(
		&userRecord,
		"50040002699049",
		"50000",
		"DEBTOR",
		"234567890123456",
		"012345678",
		"CHECKING",
		utils.Pointer(""),
		nil,
	)

	assert.Equal(t, "ACH Debit", *payloadData.Reason)
}
