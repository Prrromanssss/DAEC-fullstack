-- +goose Up
DROP TYPE IF EXISTS agent_status;
CREATE TYPE agent_status AS ENUM ('running', 'waiting', 'sleeping', 'terminated');

CREATE TABLE IF NOT EXISTS agents (
    agent_id int GENERATED ALWAYS AS IDENTITY,
    number_of_parallel_calculations int NOT NULL DEFAULT 5,
    last_ping timestamp NOT NULL,
    status agent_status NOT NULL,

    PRIMARY KEY(agent_id)
);

-- +goose Down
DROP TABLE IF EXISTS agents;
DROP TYPE IF EXISTS agent_status;