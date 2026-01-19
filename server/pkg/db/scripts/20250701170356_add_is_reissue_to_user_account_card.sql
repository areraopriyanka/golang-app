-- +goose Up
SELECT 'up SQL query';

ALTER TABLE user_account_card
ADD is_replace boolean DEFAULT false,
ADD is_reissue boolean DEFAULT false;

-- +goose Down
SELECT 'down SQL query';

ALTER TABLE user_account_card
DROP COLUMN is_reissue,
DROP COLUMN is_replace;
