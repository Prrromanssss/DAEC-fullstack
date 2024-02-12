-- +goose Up
CREATE TABLE operations (
    id UUID PRIMARY KEY,
    operation_type VARCHAR(1) UNIQUE NOT NULL,
    execution_time INTERVAL NOT NULL DEFAULT '100 seconds'
);

-- +goose Down
DROP TABLE operations;