-- +goose Up
ALTER TABLE plaid_accounts ALTER COLUMN balance_refreshed_at DROP NOT NULL;

-- +goose Down
ALTER TABLE plaid_accounts ALTER COLUMN balance_refreshed_at SET NOT NULL;
