-- +goose Up

-- plaid.ItemAuthMethod
-- From ~/go/pkg/mod/github.com/plaid/plaid-go/v34@v34.0.0/plaid/model_item_auth_method.go
CREATE TYPE plaid_auth_method AS ENUM (
    'INSTANT_AUTH',
    'INSTANT_MATCH',
    'AUTOMATED_MICRODEPOSITS',
    'SAME_DAY_MICRODEPOSITS',
    'INSTANT_MICRODEPOSITS',
    'DATABASE_MATCH',
    'DATABASE_INSIGHTS',
    'TRANSFER_MIGRATED',
    'INVESTMENTS_FALLBACK'
);

ALTER TABLE plaid_accounts 
ADD COLUMN auth_method plaid_auth_method;

-- +goose Down
ALTER TABLE plaid_accounts 
DROP COLUMN IF EXISTS auth_method;

DROP TYPE IF EXISTS plaid_auth_method;