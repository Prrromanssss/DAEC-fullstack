-- name: CreateExpression :one
INSERT INTO expressions
    (created_at, updated_at, data, parse_data, status, user_id)
VALUES
    ($1, $2, $3, $4, $5, $6)
RETURNING
    expression_id, user_id, agent_id,
    created_at, updated_at, data, parse_data,
    status, result, is_ready;

-- name: GetExpressions :many
SELECT
    expression_id, user_id, agent_id,
    created_at, updated_at, data, parse_data,
    status, result, is_ready
FROM expressions
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetExpressionByID :one
SELECT
    expression_id, user_id, agent_id,
    created_at, updated_at, data, parse_data,
    status, result, is_ready
FROM expressions
WHERE expression_id = $1;

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
SELECT
    expression_id, user_id, agent_id,
    created_at, updated_at, data, parse_data,
    status, result, is_ready
FROM expressions
WHERE status IN ('ready_for_computation', 'computing', 'terminated')
ORDER BY created_at DESC;

-- name: MakeExpressionsTerminated :exec
UPDATE expressions
SET status = 'terminated'
WHERE agent_id = $1 AND is_ready = false;

-- name: GetTerminatedExpressions :many
SELECT
    expression_id, user_id, agent_id,
    created_at, updated_at, data, parse_data,
    status, result, is_ready
FROM expressions
WHERE status = 'terminated'
ORDER BY created_at DESC;

-- name: AssignExpressionToAgent :exec
UPDATE expressions
SET agent_id = $1
WHERE expression_id = $2;