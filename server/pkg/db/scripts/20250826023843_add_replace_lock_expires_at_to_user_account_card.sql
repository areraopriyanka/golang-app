-- +goose Up
ALTER TABLE user_account_card
ADD COLUMN replace_lock_expires_at timestamptz;

-- +goose Down
ALTER TABLE user_account_card
DROP COLUMN IF EXISTS replace_lock_expires_at;
