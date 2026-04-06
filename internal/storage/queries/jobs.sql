-- name: CreateJob :one
INSERT INTO jobs (type, payload, max_attempts)
VALUES ($1, $2, $3)
RETURNING id, type, payload, status, attempts, max_attempts, error,
          created_at, claimed_at, completed_at, scheduled_for;

-- name: ClaimJob :one
UPDATE jobs
SET status = 'running',
    claimed_at = now(),
    attempts = attempts + 1
WHERE id = (
    SELECT id FROM jobs
    WHERE status = 'pending'
      AND scheduled_for <= now()
      AND attempts < max_attempts
    ORDER BY created_at ASC
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING id, type, payload, status, attempts, max_attempts, error,
          created_at, claimed_at, completed_at, scheduled_for;

-- name: CompleteJob :exec
UPDATE jobs
SET status = 'done', completed_at = now(), error = NULL
WHERE id = $1;

-- name: FailJob :exec
UPDATE jobs
SET status = CASE WHEN attempts >= max_attempts THEN 'dead' ELSE 'failed' END,
    error = $2,
    completed_at = now()
WHERE id = $1;

-- name: RetryJob :exec
UPDATE jobs
SET status = 'pending',
    scheduled_for = now() + make_interval(secs => @backoff_seconds::double precision)
WHERE id = $1 AND status = 'failed';

-- name: InsertPendingNode :one
INSERT INTO nodes (type, title, content, source_url, status, version)
VALUES ($1, $2, $3, $4, 'pending', 1)
RETURNING id, type, title, content, source_url, status, version, created_at, updated_at;
