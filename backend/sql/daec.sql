DROP TYPE IF EXISTS agent_status;
CREATE TYPE agent_status AS ENUM ('running', 'waiting', 'sleeping', 'terminated');

CREATE TABLE IF NOT EXISTS agents (
    agent_id int GENERATED ALWAYS AS IDENTITY,
    number_of_parallel_calculations int NOT NULL DEFAULT 5,
    last_ping timestamp NOT NULL,
    status agent_status NOT NULL,

    PRIMARY KEY(agent_id)
);

ALTER TABLE agents ADD COLUMN created_at timestamp NOT NULL;

CREATE TABLE IF NOT EXISTS users (
    user_id int GENERATED ALWAYS AS IDENTITY,
    email text UNIQUE NOT NULL,
    password_hash bytea NOT NULL,

    PRIMARY KEY(user_id) 
);

DROP TYPE IF EXISTS expression_status;
CREATE TYPE expression_status AS ENUM ('ready_for_computation', 'computing', 'result', 'terminated');

CREATE TABLE IF NOT EXISTS expressions (
    expression_id int GENERATED ALWAYS AS IDENTITY,
    user_id int NOT NULL,
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
	    ON DELETE SET NULL,
    FOREIGN KEY(user_id)
      REFERENCES users(user_id)
      ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS operations (
    operation_id int GENERATED ALWAYS AS IDENTITY,
    operation_type varchar(1) NOT NULL,
    execution_time int NOT NULL DEFAULT 100,
    user_id int NOT NULL,

    PRIMARY KEY(operation_id),
    CONSTRAINT operation_type_user_id UNIQUE(operation_type, user_id),
    FOREIGN KEY(user_id) 
	  REFERENCES users(user_id)
	  ON DELETE CASCADE
);

ALTER TABLE agents ADD COLUMN number_of_active_calculations int NOT NULL DEFAULT 0;