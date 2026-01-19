-- +goose Up
SELECT 'up SQL query';

ALTER TABLE user_account_card
ADD account_closure_reason character varying(255) DEFAULT NULL;

-- +goose Down
SELECT 'down SQL query';

ALTER TABLE user_account_card
DROP COLUMN account_closure_reason;
