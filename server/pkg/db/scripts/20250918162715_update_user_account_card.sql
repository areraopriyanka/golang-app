-- +goose Up
ALTER TABLE user_account_card 
    ADD COLUMN previous_card_id TEXT,
    ADD COLUMN previous_card_mask_number TEXT;

-- +goose Down
ALTER TABLE user_account_card 
    DROP COLUMN previous_card_id,
    DROP COLUMN previous_card_mask_number;