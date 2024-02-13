-- name: CreateAgent :one
INSERT INTO agents (id, created_at, number_of_parallel_calculations, last_ping, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAgents :many
SELECT * FROM agents;

-- name: GetAgentByID :one
SELECT * FROM agents
WHERE id = $1;

-- name: UpdateAgentLastPing :one
UPDATE agents
SET last_ping = $1
WHERE id = $2
RETURNING *;

-- name: UpdateAgentStatus :one
UPDATE agents
SET status = $1
WHERE id = $2
RETURNING *;