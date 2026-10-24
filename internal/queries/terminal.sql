-- name: ListTerminals :many
SELECT * FROM terminals ORDER BY id;

-- name: GetTerminal :one
SELECT * FROM terminals WHERE id = $1;

-- name: CreateTerminal :exec
INSERT INTO terminals (name)
VALUES ($1)
    ON CONFLICT (name) DO NOTHING;

-- name: CountTerminalsByStatus :many
SELECT status, COUNT(*) AS count FROM terminals GROUP BY status;

-- name: RecordHeartbeat :one
UPDATE terminals
SET uptime_seconds = uptime_seconds + CASE
        WHEN status = 'online' AND last_seen_at IS NOT NULL
            THEN EXTRACT(EPOCH FROM NOW() - last_seen_at)::BIGINT
        ELSE 0
    END,
    status = sqlc.arg(status),
    last_seen_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    RETURNING *;

-- name: AddDataSent :one
UPDATE terminals
SET data_sent_bytes = data_sent_bytes + sqlc.arg(bytes),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    RETURNING *;
