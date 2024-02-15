-- +goose Up
CREATE TABLE operations (
    id UUID PRIMARY KEY,
    operation_type VARCHAR(1) UNIQUE NOT NULL,
    execution_time INT NOT NULL DEFAULT 100
);

-- +goose Down
DROP TABLE operations;