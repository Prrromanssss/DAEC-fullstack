-- +goose Up
CREATE TABLE IF NOT EXISTS operations (
    operation_id int GENERATED ALWAYS AS IDENTITY,
    operation_type varchar(1) UNIQUE NOT NULL,
    execution_time int NOT NULL DEFAULT 100,

    PRIMARY KEY(operation_id)
);

-- +goose Down
DROP TABLE IF EXISTS operations;