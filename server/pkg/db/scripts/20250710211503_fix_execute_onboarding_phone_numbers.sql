-- +goose Up
UPDATE master_user_records
SET mobile_no = '+1' || REPLACE(mobile_no, '-', '')
WHERE mobile_no LIKE '%-%';
