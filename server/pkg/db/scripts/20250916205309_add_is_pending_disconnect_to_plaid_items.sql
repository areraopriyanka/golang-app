-- +goose Up
ALTER TABLE plaid_items ADD COLUMN is_pending_disconnect BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE plaid_items DROP COLUMN is_pending_disconnect;
