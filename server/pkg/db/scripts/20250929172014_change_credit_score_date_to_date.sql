-- +goose Up

ALTER TABLE user_credit_score 
ALTER COLUMN date TYPE DATE 
using to_date(date, 'YYYY-MM-DD');

-- +goose Down

ALTER TABLE user_credit_score 
ALTER COLUMN date TYPE CHARACTER VARYING(40) 
USING TO_CHAR(date, 'YYYY-MM-DD');
