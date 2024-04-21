-- name: UpdateOperationTime :one
UPDATE operations
SET execution_time = $1
WHERE operation_type = $2 AND user_id = $3
RETURNING operation_id, operation_type, execution_time, user_id;

-- name: GetOperations :many
SELECT
    operation_id, operation_type, execution_time, user_id
FROM operations
WHERE user_id = $1
ORDER BY operation_type DESC;

-- name: GetOperationTimeByType :one
SELECT execution_time
FROM operations
WHERE operation_type = $1 AND user_id = $2;

-- name: NewOperationsForUser :exec
INSERT INTO operations (operation_type, user_id) VALUES 
('+', $1),
('-', $1),
('*', $1),
('/', $1);