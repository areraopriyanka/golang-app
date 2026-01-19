-- +goose Up

ALTER TABLE master_user_records
  ADD COLUMN agreement_privacy_notice_hash                 TEXT,
  ADD COLUMN agreement_card_and_deposit_hash               TEXT,
  ADD COLUMN agreement_dreamfi_ach_authorization_hash      TEXT,
  ADD COLUMN agreement_e_sign_hash                         TEXT,
  ADD COLUMN agreement_terms_of_service_hash               TEXT,
  ADD COLUMN agreement_privacy_notice_signed_at            TIMESTAMPTZ,
  ADD COLUMN agreement_card_and_deposit_signed_at          TIMESTAMPTZ,
  ADD COLUMN agreement_dreamfi_ach_authorization_signed_at TIMESTAMPTZ,
  ADD COLUMN agreement_e_sign_signed_at                    TIMESTAMPTZ,
  ADD COLUMN agreement_terms_of_service_signed_at          TIMESTAMPTZ;
