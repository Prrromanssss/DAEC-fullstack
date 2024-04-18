-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    user_id int GENERATED ALWAYS AS IDENTITY,
    email text UNIQUE NOT NULL,
    password_hash bytea NOT NULL,

    PRIMARY KEY(user_id) 
);

-- +goose Down
DROP TABLE IF EXISTS users;