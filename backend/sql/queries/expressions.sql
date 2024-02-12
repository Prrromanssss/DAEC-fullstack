-- name: CreateExpression :one
INSERT INTO expressions (id, created_at, updated_at, data, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
