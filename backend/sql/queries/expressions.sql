-- name: CreateExpression :one
INSERT INTO expressions (id, created_at, updated_at, data, parse_data, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetExpressions :many
SELECT * FROM expressions;

-- name: GetExpressionByID :one
SELECT * FROM expressions
WHERE id = $1;