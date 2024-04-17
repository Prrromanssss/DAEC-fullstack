-- name: CreateAgent :one
INSERT INTO agents (created_at, number_of_parallel_calculations, last_ping, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAgents :many
SELECT * FROM agents
ORDER BY created_at DESC;

-- name: GetAgentByID :one
SELECT * FROM agents
WHERE agent_id = $1;

-- name: UpdateAgentLastPing :exec
UPDATE agents
SET last_ping = $1
WHERE agent_id = $2;

-- name: UpdateAgentStatus :exec
UPDATE agents
SET status = $1
WHERE agent_id = $2;

-- name: DeleteAgents :exec
TRUNCATE agents RESTART IDENTITY;

-- name: DecrementNumberOfActiveCalculations :exec
UPDATE agents
SET number_of_active_calculations = number_of_active_calculations - 1
WHERE agent_id = $1;

-- name: IncrementNumberOfActiveCalculations :exec
UPDATE agents
SET number_of_active_calculations = number_of_active_calculations + 1
WHERE agent_id = $1;