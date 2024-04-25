-- +goose Up
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

-- +goose Down
DROP TABLE IF EXISTS operations;