-- +goose Up
ALTER TABLE plaid_items ADD COLUMN item_error TEXT;

-- +goose Down
ALTER TABLE plaid_items DROP COLUMN item_error;