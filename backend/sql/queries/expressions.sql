-- name: CreateExpression :one
INSERT INTO expressions (created_at, updated_at, data, parse_data, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetExpressions :many
SELECT * FROM expressions
ORDER BY created_at DESC;

-- name: GetExpressionByID :one
SELECT * FROM expressions
WHERE expression_id = $1;

-- name: UpdateExpressionData :exec
UPDATE expressions
SET data = $1
WHERE expression_id = $2;

-- name: UpdateExpressionParseData :exec
UPDATE expressions
SET parse_data = $1
WHERE expression_id = $2;

-- name: MakeExpressionReady :exec
UPDATE expressions
SET parse_data = $1, result = $2, updated_at = $3, is_ready = True, status = 'result'
WHERE expression_id = $4;

-- name: UpdateExpressionStatus :exec
UPDATE expressions
SET status = $1
WHERE expression_id = $2;

-- name: GetComputingExpressions :many
SELECT * FROM expressions
WHERE status = 'computing'
ORDER BY created_at DESC;