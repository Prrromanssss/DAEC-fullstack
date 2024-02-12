-- +goose Up
ALTER TABLE expressions ADD COLUMN parse_data TEXT NOT NULL;

-- +goose Down
ALTER TABLE expressions DROP COLUMN parse_data;