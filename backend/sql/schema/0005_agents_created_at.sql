-- +goose Up
ALTER TABLE agents ADD COLUMN created_at TIMESTAMP NOT NULL;

-- +goose Down
ALTER TABLE agents DROP COLUMN created_at;