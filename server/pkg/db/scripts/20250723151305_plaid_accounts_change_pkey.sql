-- +goose Up

ALTER TABLE plaid_accounts
  DROP CONSTRAINT plaid_accounts_pkey,
  DROP CONSTRAINT plaid_accounts_plaid_account_id_unique;

ALTER TABLE plaid_accounts
  DROP COLUMN id;

ALTER TABLE plaid_accounts
  ADD CONSTRAINT plaid_accounts_pkey PRIMARY KEY (plaid_account_id);

-- +goose Down

ALTER TABLE plaid_accounts
  DROP CONSTRAINT plaid_accounts_pkey;

ALTER TABLE plaid_accounts
  ADD COLUMN id uuid NOT NULL DEFAULT gen_random_uuid();

ALTER TABLE plaid_accounts
  ADD CONSTRAINT plaid_accounts_pkey PRIMARY KEY (id);

ALTER TABLE plaid_accounts
  ADD CONSTRAINT plaid_accounts_plaid_account_id_unique UNIQUE (plaid_account_id);
