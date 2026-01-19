-- +goose Up
SELECT 'up SQL query';

ALTER TABLE user_account_card
ADD card_expiration_date character varying(6) DEFAULT '';

-- +goose Down
SELECT 'down SQL query';

ALTER TABLE user_account_card
DROP COLUMN card_expiration_date;
