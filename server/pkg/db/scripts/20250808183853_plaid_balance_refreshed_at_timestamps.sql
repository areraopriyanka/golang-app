-- +goose Up
ALTER TABLE plaid_accounts
  ALTER COLUMN balance_refreshed_at
  TYPE timestamptz
  USING timezone(current_setting('TimeZone'), balance_refreshed_at);

-- +goose Down
ALTER TABLE plaid_accounts
    ALTER COLUMN balance_refreshed_at
    TYPE timestamp without time zone
    USING timezone(current_setting('TimeZone'), balance_refreshed_at);
