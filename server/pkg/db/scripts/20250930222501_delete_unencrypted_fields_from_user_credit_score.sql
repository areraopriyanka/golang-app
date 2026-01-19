-- +goose Up
ALTER TABLE user_credit_score 
ALTER COLUMN encrypted_credit_data SET NOT NULL;

ALTER TABLE user_credit_score 
DROP COLUMN score,
DROP COLUMN increase,
DROP COLUMN debtwise_customer_number,
DROP COLUMN payment_history_amount,
DROP COLUMN payment_history_factor,
DROP COLUMN credit_utilization_amount,
DROP COLUMN credit_utilization_factor,
DROP COLUMN derogatory_marks_amount,
DROP COLUMN derogatory_marks_factor,
DROP COLUMN credit_age_amount,
DROP COLUMN credit_age_factor,
DROP COLUMN credit_mix_amount,
DROP COLUMN credit_mix_factor,
DROP COLUMN new_credit_amount,
DROP COLUMN new_credit_factor,
DROP COLUMN total_accounts_amount,
DROP COLUMN total_accounts_factor;

-- +goose Down
ALTER TABLE user_credit_score 
ADD COLUMN score INT,
ADD COLUMN increase INT,
ADD COLUMN debtwise_customer_number INT,
ADD COLUMN payment_history_amount FLOAT8,
ADD COLUMN payment_history_factor VARCHAR(255),
ADD COLUMN credit_utilization_amount FLOAT8,
ADD COLUMN credit_utilization_factor VARCHAR(255),
ADD COLUMN derogatory_marks_amount INT,
ADD COLUMN derogatory_marks_factor VARCHAR(255),
ADD COLUMN credit_age_amount FLOAT8,
ADD COLUMN credit_age_factor VARCHAR(255),
ADD COLUMN credit_mix_amount INT,
ADD COLUMN credit_mix_factor VARCHAR(255),
ADD COLUMN new_credit_amount INT,
ADD COLUMN new_credit_factor VARCHAR(255),
ADD COLUMN total_accounts_amount INT,
ADD COLUMN total_accounts_factor VARCHAR(255);

ALTER TABLE user_credit_score 
ALTER COLUMN encrypted_credit_data DROP NOT NULL;