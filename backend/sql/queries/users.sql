-- name: GetUser :one
SELECT user_id, email, password_hash
FROM users
WHERE user_id = $1;

-- name: SaveUser :one
INSERT INTO users
    (email, password_hash)
VALUES
    ($1, $2)
RETURNING user_id;