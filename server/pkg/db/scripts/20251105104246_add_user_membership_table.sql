-- +goose Up

CREATE TYPE membership_status AS ENUM ('subscribed', 'unsubscribed');

CREATE TABLE public.user_membership (
    id uuid NOT NULL PRIMARY KEY,
    user_id uuid  NOT NULL,
    membership_status membership_status NOT NULL,
    created_at              timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at              timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_membership_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.master_user_records (id)
);


-- +goose Down
DROP TABLE IF EXISTS public.user_membership;
DROP TYPE IF EXISTS membership_status;