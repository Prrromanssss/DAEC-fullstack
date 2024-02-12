-- +goose Up
CREATE TABLE expressions (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    data TEXT NOT NULL,
    status VARCHAR(32) NOT NULL
);


-- +goose Down
DROP TABLE expressions;