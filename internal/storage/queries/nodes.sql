-- name: UpsertRawNode :one
INSERT INTO nodes (type, title, content, source_url)
VALUES ($1, $2, $3, $4)
ON CONFLICT (source_url) WHERE source_url IS NOT NULL
DO UPDATE SET title = EXCLUDED.title,
             content = EXCLUDED.content,
             updated_at = now()
RETURNING id, type, title, source_url, created_at, updated_at,
          (xmax = 0) AS created;

-- name: ListRecentNodes :many
SELECT id, title, source_url, created_at
FROM nodes
ORDER BY created_at DESC
LIMIT $1;

-- name: ListNodes :many
SELECT id, title, source_url, created_at
FROM nodes
WHERE (sqlc.narg('cursor')::TIMESTAMPTZ IS NULL OR created_at < sqlc.narg('cursor'))
ORDER BY created_at DESC
LIMIT $1;

-- name: DeleteNode :exec
DELETE FROM nodes WHERE id = $1;
