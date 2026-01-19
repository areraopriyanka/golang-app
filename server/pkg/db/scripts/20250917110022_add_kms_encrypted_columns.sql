-- +goose Up

ALTER TABLE master_user_records
ADD COLUMN kms_encrypted_ledger_password bytea;

ALTER TABLE user_public_keys
ADD COLUMN kms_encrypted_api_key bytea;

ALTER TABLE plaid_items ALTER encrypted_access_token DROP NOT NULL;

-- TODO: Need to add NOT NULL constraints once all tokens are migrated
ALTER TABLE plaid_items
ADD COLUMN kms_encrypted_access_token bytea;



-- +goose Down

ALTER TABLE master_user_records
DROP COLUMN IF EXISTS kms_encrypted_ledger_password;

ALTER TABLE user_public_keys
DROP COLUMN IF EXISTS kms_encrypted_api_key;

ALTER TABLE plaid_items
DROP COLUMN IF EXISTS kms_encrypted_access_token;
