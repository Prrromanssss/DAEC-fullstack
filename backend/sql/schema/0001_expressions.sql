-- +goose Up
DROP TYPE IF EXISTS expression_status;
CREATE TYPE expression_status AS ENUM ('ready for computation', 'computing', 'result', 'terminated');

CREATE TABLE IF NOT EXISTS expressions (
    expression_id int GENERATED ALWAYS AS IDENTITY,
    agent_id int,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    data text NOT NULL,
    parse_data text NOT NULL,
    status expression_status NOT NULL,
    result int NOT NULL DEFAULT 0,
    is_ready boolean NOT NULL DEFAULT false,

    PRIMARY KEY(expression_id),
    FOREIGN KEY(agent_id) 
	  REFERENCES agents(agent_id)
	  ON DELETE SET NULL
);


-- +goose Down
DROP TABLE IF EXISTS expressions;
DROP TYPE IF EXISTS expression_status;