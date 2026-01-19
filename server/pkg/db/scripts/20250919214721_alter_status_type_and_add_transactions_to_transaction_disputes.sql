-- +goose Up
ALTER TABLE transaction_disputes
  ALTER COLUMN status TYPE text
    USING (status::text);

CREATE TYPE new_transaction_disputes_status AS ENUM ('pending', 'credited', 'rejected', 'voided');
UPDATE transaction_disputes SET status = 'credited' WHERE status = 'accepted';

ALTER TABLE transaction_disputes
  ADD COLUMN provisional_credit_transaction text,
  ADD COLUMN void_credit_transaction text,
  ADD COLUMN credited_at timestamptz,
  ADD COLUMN voided_at timestamptz,
  ALTER COLUMN status TYPE new_transaction_disputes_status
    USING (status::new_transaction_disputes_status);

DROP TYPE transaction_disputes_status;
ALTER TYPE new_transaction_disputes_status RENAME TO transaction_disputes_status;


-- +goose Down
ALTER TABLE transaction_disputes
  ALTER COLUMN status TYPE text
    USING (status::text);

CREATE TYPE old_transaction_disputes_status AS ENUM ('pending', 'accepted', 'rejected');
UPDATE transaction_disputes SET status = 'accepted' WHERE status = 'credited';
UPDATE transaction_disputes SET status = 'rejected' WHERE status = 'voided';

ALTER TABLE transaction_disputes
  DROP COLUMN provisional_credit_transaction,
  DROP COLUMN void_credit_transaction,
  DROP COLUMN credited_at,
  DROP COLUMN voided_at,
  ALTER COLUMN status TYPE old_transaction_disputes_status
    USING (status::text::old_transaction_disputes_status);

DROP TYPE transaction_disputes_status;
ALTER TYPE old_transaction_disputes_status RENAME TO transaction_disputes_status;
