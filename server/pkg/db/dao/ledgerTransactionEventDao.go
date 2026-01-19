package dao

import (
	"errors"
	"time"

	"braces.dev/errtrace"
	"github.com/jinzhu/gorm"
)

type LedgerTransactionEventDao struct {
	EventId                          string    `gorm:"column:event_id;primaryKey"`
	Channel                          string    `gorm:"column:channel"`
	TransactionType                  string    `gorm:"column:transaction_type"`
	TransactionNumber                string    `gorm:"column:transaction_number"`
	BinNumber                        string    `gorm:"column:bin_number"`
	CardId                           string    `gorm:"column:card_id"`
	UserId                           string    `gorm:"column:user_id"`
	AccountNumber                    string    `gorm:"column:account_number"`
	AccountRoutingNumber             string    `gorm:"column:account_routing_number"`
	ExternalBankAccountName          string    `gorm:"column:external_bank_account_name"`
	ExternalBankAccountRoutingNumber string    `gorm:"column:external_bank_account_routing_number"`
	ExternalBankAccountNumber        []byte    `gorm:"column:external_bank_account_number"`
	CardPayeeId                      string    `gorm:"column:card_payee_id"`
	CardPayeeName                    string    `gorm:"column:card_payee_name"`
	InstructedAmount                 int       `gorm:"column:instructed_amount"`
	InstructedCurrency               string    `gorm:"column:instructed_currency"`
	IsOutward                        bool      `gorm:"column:is_outward"`
	Mcc                              string    `gorm:"column:mcc"`
	RawPayload                       []byte    `gorm:"column:raw_payload"`
	CreatedAt                        time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt                        time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (LedgerTransactionEventDao) FindOneByEventId(db *gorm.DB, eventId string) (*LedgerTransactionEventDao, error) {
	var ledgerEventRecord LedgerTransactionEventDao
	result := db.Where("event_id=?", eventId).Find(&ledgerEventRecord)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errtrace.Wrap(result.Error)
	}

	return &ledgerEventRecord, nil
}

func (LedgerTransactionEventDao) TableName() string {
	return "ledger_transaction_events"
}
