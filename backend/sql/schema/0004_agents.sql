-- +goose Up
CREATE TYPE agent_status AS ENUM ('running', 'waiting', 'sleeping', 'terminated');

CREATE TABLE agents (
    id UUID PRIMARY KEY,
    number_of_parallel_calculations INTEGER NOT NULL DEFAULT 5,
    last_ping TIMESTAMP NOT NULL,
    status agent_status NOT NULL
);


-- +goose Down
DROP TABLE agents;
DROP TYPE agent_status;