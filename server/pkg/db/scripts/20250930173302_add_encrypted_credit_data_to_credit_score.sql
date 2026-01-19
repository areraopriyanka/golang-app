-- +goose Up
ALTER TABLE user_credit_score 
ADD COLUMN encrypted_credit_data BYTEA;

-- +goose Down
ALTER TABLE user_credit_score 
DROP COLUMN encrypted_credit_data;
