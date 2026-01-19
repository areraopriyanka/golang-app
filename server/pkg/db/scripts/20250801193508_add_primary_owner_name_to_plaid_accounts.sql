-- +goose Up
ALTER TABLE plaid_accounts
ADD COLUMN primary_owner_name text;

-- +goose Down
ALTER TABLE plaid_accounts
DROP COLUMN primary_owner_name;
