-- +goose Up

--  Step 1 - Drop fkey constraints from child tables
ALTER TABLE demographic_updates DROP CONSTRAINT demographic_updates_user_id_fkey;
ALTER TABLE ledger_transaction_events DROP CONSTRAINT ledger_transaction_events_pkey;
ALTER TABLE signable_payloads DROP CONSTRAINT signable_payloads_pkey;
ALTER TABLE master_user_otp DROP CONSTRAINT master_user_otp_user_id_fkey;
ALTER TABLE plaid_items DROP CONSTRAINT plaid_items_user_id_fkey;
ALTER TABLE sardine_kyc_data DROP CONSTRAINT sardine_kyc_data_user_id_fkey;
ALTER TABLE user_credit_score DROP CONSTRAINT user_credit_score_user_id_fkey;
ALTER TABLE user_public_keys DROP CONSTRAINT user_public_keys_user_id_fkey;
ALTER TABLE user_account_card DROP CONSTRAINT user_account_card_user_id_fkey;
ALTER TABLE plaid_accounts DROP CONSTRAINT fk_plaid_accounts_users;
ALTER TABLE master_user_logins DROP CONSTRAINT master_user_logins_user_id_fkey;
ALTER TABLE transaction_disputes DROP CONSTRAINT transaction_disputes_user_id_fkey;

-- Step 2 - Alter child tables user_id column type from varchar to uuid
ALTER TABLE demographic_updates ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
ALTER TABLE ledger_transaction_events ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
ALTER TABLE signable_payloads ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
ALTER TABLE master_user_otp ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
ALTER TABLE plaid_items ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
ALTER TABLE sardine_kyc_data ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
ALTER TABLE user_credit_score ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
ALTER TABLE user_public_keys ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
ALTER TABLE user_account_card ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
ALTER TABLE plaid_accounts ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
ALTER TABLE master_user_logins ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
ALTER TABLE transaction_disputes ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
 
-- Step 3 - Alter master_user_records table's id column type from varchar to uuid
ALTER TABLE master_user_records ALTER COLUMN id TYPE uuid USING id::uuid;

-- Step 4 - Add fkey constraint back to child tables
ALTER TABLE demographic_updates
ADD CONSTRAINT demographic_updates_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE ledger_transaction_events
ADD CONSTRAINT ledger_transaction_events_pkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE signable_payloads
ADD CONSTRAINT signable_payloads_pkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE master_user_otp
ADD CONSTRAINT master_user_otp_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE plaid_items
ADD CONSTRAINT plaid_items_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE sardine_kyc_data
ADD CONSTRAINT sardine_kyc_data_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE user_credit_score
ADD CONSTRAINT user_credit_score_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE user_public_keys
ADD CONSTRAINT user_public_keys_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE user_account_card
ADD CONSTRAINT user_account_card_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE plaid_accounts
ADD CONSTRAINT fk_plaid_accounts_users FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE master_user_logins
ADD CONSTRAINT master_user_logins_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE transaction_disputes
ADD CONSTRAINT transaction_disputes_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);

-- +goose Down

-- Step 1 - Drop fkey constraints from child tables
ALTER TABLE demographic_updates DROP CONSTRAINT demographic_updates_user_id_fkey;
ALTER TABLE ledger_transaction_events DROP CONSTRAINT ledger_transaction_events_pkey;
ALTER TABLE signable_payloads DROP CONSTRAINT signable_payloads_pkey;
ALTER TABLE master_user_otp DROP CONSTRAINT master_user_otp_user_id_fkey;
ALTER TABLE plaid_items DROP CONSTRAINT plaid_items_user_id_fkey;
ALTER TABLE sardine_kyc_data DROP CONSTRAINT sardine_kyc_data_user_id_fkey;
ALTER TABLE user_credit_score DROP CONSTRAINT user_credit_score_user_id_fkey;
ALTER TABLE user_public_keys DROP CONSTRAINT user_public_keys_user_id_fkey;
ALTER TABLE user_account_card DROP CONSTRAINT user_account_card_user_id_fkey;
ALTER TABLE plaid_accounts DROP CONSTRAINT fk_plaid_accounts_users;
ALTER TABLE master_user_logins DROP CONSTRAINT master_user_logins_user_id_fkey;
ALTER TABLE transaction_disputes DROP CONSTRAINT transaction_disputes_user_id_fkey;

-- Step 2 - Alter master_user_records table's id column type from uuid back to varchar
ALTER TABLE master_user_records ALTER COLUMN id TYPE varchar(36) USING id::text;

-- Step 3 - Alter child tables user_id column type from uuid back to varchar
ALTER TABLE demographic_updates ALTER COLUMN user_id TYPE varchar(36) USING user_id::text;
ALTER TABLE ledger_transaction_events ALTER COLUMN user_id TYPE varchar(36) USING user_id::text;
ALTER TABLE signable_payloads ALTER COLUMN user_id TYPE varchar(36) USING user_id::text;
ALTER TABLE master_user_otp ALTER COLUMN user_id TYPE varchar(36) USING user_id::text;
ALTER TABLE plaid_items ALTER COLUMN user_id TYPE varchar(36) USING user_id::text;
ALTER TABLE sardine_kyc_data ALTER COLUMN user_id TYPE varchar(36) USING user_id::text;
ALTER TABLE user_credit_score ALTER COLUMN user_id TYPE varchar(36) USING user_id::text;
ALTER TABLE user_public_keys ALTER COLUMN user_id TYPE varchar(36) USING user_id::text;
ALTER TABLE user_account_card ALTER COLUMN user_id TYPE varchar(36) USING user_id::text;
ALTER TABLE plaid_accounts ALTER COLUMN user_id TYPE varchar(36) USING user_id::text;
ALTER TABLE master_user_logins ALTER COLUMN user_id TYPE varchar(36) USING user_id::text;
ALTER TABLE transaction_disputes ALTER COLUMN user_id TYPE varchar(36) USING user_id::text;

-- Step 4 - Add fkey constraint back to child tables
ALTER TABLE demographic_updates
ADD CONSTRAINT demographic_updates_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE ledger_transaction_events
ADD CONSTRAINT ledger_transaction_events_pkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE signable_payloads
ADD CONSTRAINT signable_payloads_pkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE master_user_otp
ADD CONSTRAINT master_user_otp_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE plaid_items
ADD CONSTRAINT plaid_items_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE sardine_kyc_data
ADD CONSTRAINT sardine_kyc_data_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE user_credit_score
ADD CONSTRAINT user_credit_score_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE user_public_keys
ADD CONSTRAINT user_public_keys_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE user_account_card
ADD CONSTRAINT user_account_card_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE plaid_accounts
ADD CONSTRAINT fk_plaid_accounts_users FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE master_user_logins
ADD CONSTRAINT master_user_logins_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);
ALTER TABLE transaction_disputes
ADD CONSTRAINT transaction_disputes_user_id_fkey FOREIGN KEY (user_id) REFERENCES master_user_records(id);