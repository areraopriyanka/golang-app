-- +goose Up

ALTER TABLE ledger_transaction_events
  ADD COLUMN card_payee_id text,
  ADD COLUMN card_payee_name text;

-- +goose Down

ALTER TABLE ledger_transaction_events
  DROP COLUMN card_payee_id,
  DROP COLUMN card_payee_name;
