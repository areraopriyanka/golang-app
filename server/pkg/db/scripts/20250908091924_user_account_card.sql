-- +goose Up
ALTER TABLE user_account_card 
    ADD COLUMN card_mask_number TEXT;


-- +goose Down
ALTER TABLE user_account_card 
  DROP COLUMN  card_mask_number;
