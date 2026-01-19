-- +goose Up

-- Plaid accounts come in many more flavors, but for the moment, we're only supporting "checking" and "savings"
CREATE TYPE plaid_account_subtype AS ENUM ('checking', 'savings');

CREATE TABLE plaid_accounts (
    id                      uuid                     NOT NULL PRIMARY KEY,
    user_id                 character varying(36)    NOT NULL,
    plaid_item_id           character varying(255)   NOT NULL, -- the associated plaid item
    plaid_account_id        text                     NOT NULL,
    name                    text                     NOT NULL,
    subtype                 plaid_account_subtype    NOT NULL,
    mask                    text,                              -- the last 2-4 digits of the account number
    institution_id          text,
    available_balance_cents bigint,
    balance_refreshed_at    timestamp                NOT NULL, -- last time /accounts/balance/get was called
    created_at              timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at              timestamp with time zone DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT plaid_accounts_plaid_account_id_unique UNIQUE (plaid_account_id),

    CONSTRAINT fk_plaid_accounts_plaid_items FOREIGN KEY (plaid_item_id) REFERENCES plaid_items (plaid_item_id),

    CONSTRAINT fk_plaid_accounts_users FOREIGN KEY (user_id) REFERENCES master_user_records (id)
);

-- +goose Down
DROP TABLE IF EXISTS plaid_accounts;
DROP TYPE IF EXISTS plaid_account_subtype;
