-- name: CreateAgent :one
INSERT INTO agents
    (created_at, number_of_parallel_calculations, last_ping, status)
VALUES
    ($1, $2, $3, $4)
RETURNING
    agent_id, number_of_parallel_calculations,
    last_ping, status, created_at,
    number_of_active_calculations;

-- name: GetAgents :many
SELECT
    agent_id, number_of_parallel_calculations,
    last_ping, status, created_at,
    number_of_active_calculations
FROM agents
ORDER BY created_at DESC;

-- name: UpdateAgentLastPing :exec
UPDATE agents
SET last_ping = $1
WHERE agent_id = $2;

-- name: UpdateAgentStatus :exec
UPDATE agents
SET status = $1
WHERE agent_id = $2;

-- name: UpdateTerminateAgentByID :exec
UPDATE agents
SET status = 'terminated', number_of_active_calculations = 0
WHERE agent_id = $1;

-- name: DecrementNumberOfActiveCalculations :exec
UPDATE agents
SET number_of_active_calculations = number_of_active_calculations - 1
WHERE agent_id = $1;

-- name: IncrementNumberOfActiveCalculations :exec
UPDATE agents
SET number_of_active_calculations = number_of_active_calculations + 1
WHERE agent_id = $1;

-- name: TerminateAgents :many
UPDATE agents
SET status = 'terminated', number_of_active_calculations = 0
WHERE EXTRACT(SECOND FROM NOW()::timestamp - agents.last_ping) > $1::numeric
RETURNING agent_id;

-- name: TerminateOldAgents :exec
DELETE FROM agents
WHERE status = 'terminated';

-- name: GetBusyAgents :many
SELECT agent_id
FROM agents
WHERE number_of_active_calculations != 0;