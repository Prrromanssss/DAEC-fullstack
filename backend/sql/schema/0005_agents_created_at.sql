-- +goose Up
ALTER TABLE agents ADD COLUMN created_at timestamp NOT NULL;

-- +goose Down
ALTER TABLE agents DROP COLUMN created_at;