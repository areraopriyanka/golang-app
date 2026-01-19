-- +goose Up

CREATE TYPE transaction_disputes_status AS ENUM ('pending', 'accepted', 'rejected');

CREATE TABLE transaction_disputes
(
    id uuid NOT NULL,
    status transaction_disputes_status NOT NULL,
    transaction_identifier character varying(100) NOT NULL,
    reason character varying(255) NOT NULL,
    details text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    extra_info text,
    user_id character varying(36),
    CONSTRAINT transaction_disputes_pkey PRIMARY KEY (id),
    CONSTRAINT transaction_identifier_unique UNIQUE (transaction_identifier),
    CONSTRAINT transaction_disputes_user_id_fkey FOREIGN KEY (user_id)
        REFERENCES master_user_records (id)
);


-- +goose Down
DROP TABLE IF EXISTS transaction_disputes;
DROP TYPE IF EXISTS transaction_disputes_status;