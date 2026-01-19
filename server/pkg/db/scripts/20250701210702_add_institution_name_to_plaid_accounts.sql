-- +goose Up
ALTER TABLE plaid_accounts ADD COLUMN institution_name text;

-- +goose Down
ALTER TABLE plaid_accounts DROP COLUMN institution_name;
