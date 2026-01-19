-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE ledger_transaction_events (
  event_id text NOT NULL UNIQUE,
  channel text NOT NULL,
  transaction_type text NOT NULL,
  user_id character varying(36) NOT NULL,
  account_number text NOT NULL,
  account_routing_number text,
  instructed_amount bigint NOT NULL,
  instructed_currency text NOT NULL,
  is_outward boolean NOT NULL,
  external_bank_account_name text,
  external_bank_account_routing_number text,
  external_bank_account_number bytea,
  mcc text,
  raw_payload bytea NOT NULL,
  created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (event_id)
)

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE ledger_transaction_events;
-- +goose StatementEnd
