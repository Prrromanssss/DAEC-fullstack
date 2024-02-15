-- +goose Up
ALTER TABLE expressions
ADD COLUMN result INT NOT NULL DEFAULT 0,
ADD COLUMN is_ready BOOLEAN NOT NULL DEFAULT False;
    
-- +goose Down
ALTER TABLE expressions
DROP COLUMN result,
DROP COLUMN is_ready;