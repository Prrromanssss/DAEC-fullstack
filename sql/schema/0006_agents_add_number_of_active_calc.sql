-- +goose Up
ALTER TABLE agents ADD COLUMN number_of_active_calculations int NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE agents DROP COLUMN number_of_active_calculations;