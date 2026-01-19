-- +goose Up
ALTER TABLE public.signable_payloads
  ADD COLUMN consumed_at timestamp with time zone,
  ADD COLUMN user_id character varying(36);

-- premature optimization for when we want to bulk remove used payloads
CREATE INDEX idx_signable_payloads_consumed_at ON public.signable_payloads(consumed_at);

-- +goose Down
ALTER TABLE public.signable_payloads
  DROP COLUMN consumed_at,
  DROP COLUMN user_id;

DROP INDEX idx_signable_payloads_consumed_at;
