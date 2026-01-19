-- +goose Up

ALTER TABLE ledger_transaction_events
  ADD COLUMN transaction_number text;

-- +goose Down

ALTER TABLE ledger_transaction_events
  DROP COLUMN transaction_number;
