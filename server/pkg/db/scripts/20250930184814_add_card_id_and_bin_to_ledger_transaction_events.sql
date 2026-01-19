-- +goose Up

ALTER TABLE ledger_transaction_events
  ADD COLUMN bin_number text,
  ADD COLUMN card_id text;

-- +goose Down

ALTER TABLE ledger_transaction_events
  DROP COLUMN bin_number,
  DROP COLUMN card_id;
