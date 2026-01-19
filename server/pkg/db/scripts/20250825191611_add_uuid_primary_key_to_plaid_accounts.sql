-- +goose Up

-- Drop the current primary key constraint
ALTER TABLE plaid_accounts
  DROP CONSTRAINT plaid_accounts_pkey;

-- Add UUID column as primary key
ALTER TABLE plaid_accounts
  ADD COLUMN id uuid NOT NULL DEFAULT gen_random_uuid();

-- Set the new UUID column as primary key
ALTER TABLE plaid_accounts
  ADD CONSTRAINT plaid_accounts_pkey PRIMARY KEY (id);

-- Add composite unique constraint to prevent Plaid account ID collisions between items
ALTER TABLE plaid_accounts
  ADD CONSTRAINT plaid_accounts_item_account_unique UNIQUE (plaid_item_id, plaid_account_id);

-- +goose Down

-- Drop the composite unique constraint
ALTER TABLE plaid_accounts
  DROP CONSTRAINT plaid_accounts_item_account_unique;

-- Drop the UUID primary key
ALTER TABLE plaid_accounts
  DROP CONSTRAINT plaid_accounts_pkey;

-- Remove the UUID column
ALTER TABLE plaid_accounts
  DROP COLUMN id;

-- Restore plaid_account_id as primary key
ALTER TABLE plaid_accounts
  ADD CONSTRAINT plaid_accounts_pkey PRIMARY KEY (plaid_account_id);
