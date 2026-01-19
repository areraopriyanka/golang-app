-- +goose Up
ALTER TABLE user_account_card ADD COLUMN suspended_at timestamptz;

ALTER TABLE user_account_card ADD COLUMN closed_at timestamptz;

ALTER TABLE user_account_card DROP CONSTRAINT user_account_card_account_status_check;
ALTER TABLE user_account_card ADD CONSTRAINT user_account_card_account_status_check CHECK (account_status IN('ACTIVE','CLOSED','SUSPENDED'));




-- +goose Down
ALTER TABLE user_account_card DROP COLUMN IF EXISTS suspended_at;

ALTER TABLE user_account_card DROP COLUMN IF EXISTS closed_at;

ALTER TABLE user_account_card DROP CONSTRAINT user_account_card_account_status_check;
ALTER TABLE user_account_card ADD CONSTRAINT user_account_card_account_status_check CHECK (account_status IN('ACTIVE','CLOSED'));


