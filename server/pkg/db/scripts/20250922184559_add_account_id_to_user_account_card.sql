-- +goose Up
ALTER TABLE user_account_card
ADD COLUMN account_id text;

-- +goose Down
ALTER TABLE user_account_card
DROP COLUMN IF EXISTS account_id;
