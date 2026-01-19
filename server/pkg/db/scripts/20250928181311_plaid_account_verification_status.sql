-- +goose Up

CREATE TYPE plaid_verification_status AS ENUM (
  -- From ~/go/pkg/mod/github.com/plaid/plaid-go/v34@v34.0.0/plaid/model_link_delivery_verification_status.go
	'automatically_verified',
	'pending_automatic_verification',
	'pending_manual_verification',
	'manually_verified',
	'verification_expired',
	'verification_failed',
	'database_matched',
	'database_insights_pending'
);

ALTER TABLE plaid_accounts
ADD COLUMN verification_status plaid_verification_status;

-- +goose Down
ALTER TABLE plaid_accounts
DROP COLUMN IF EXISTS verification_status;

DROP TYPE IF EXISTS plaid_verification_status;