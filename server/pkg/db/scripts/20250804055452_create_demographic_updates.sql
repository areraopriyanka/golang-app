-- +goose Up

CREATE TYPE demographic_update_status AS ENUM ('pending', 'accepted', 'rejected');

CREATE TYPE demographic_update_type AS ENUM ('full_name','email', 'address', 'mobile_no');

CREATE TABLE demographic_updates
(
    id uuid NOT NULL,
    updated_value jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    extra_info text,
    user_id character varying(36) NOT NULL,
    type demographic_update_type NOT NULL,
    status demographic_update_status NOT NULL,
    CONSTRAINT demographic_updates_pkey PRIMARY KEY (id),
    CONSTRAINT demographic_updates_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.master_user_records (id)
);


-- +goose Down
DROP TABLE IF EXISTS demographic_updates;
DROP TYPE IF EXISTS demographic_update_status;
DROP TYPE IF EXISTS demographic_update_type;