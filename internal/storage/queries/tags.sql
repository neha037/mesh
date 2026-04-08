-- name: UpsertTag :one
-- Insert tag, return existing if name conflict.
INSERT INTO tags (name) VALUES (lower(trim($1)))
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id, name;

-- name: AssociateNodeTag :exec
-- Link node to tag with confidence score, keep highest confidence on conflict.
INSERT INTO node_tags (node_id, tag_id, confidence)
VALUES ($1, $2, $3)
ON CONFLICT (node_id, tag_id) DO UPDATE
SET confidence = GREATEST(node_tags.confidence, EXCLUDED.confidence);

-- name: GetNodeTags :many
-- Get all tags for a node ordered by confidence.
SELECT t.id, t.name, nt.confidence
FROM tags t JOIN node_tags nt ON t.id = nt.tag_id
WHERE nt.node_id = $1
ORDER BY nt.confidence DESC;

-- name: BulkAssociateNodeTags :exec
INSERT INTO node_tags (node_id, tag_id, confidence)
SELECT @node_id::uuid, unnest(@tag_ids::uuid[]), @confidence::real
ON CONFLICT (node_id, tag_id) DO UPDATE
SET confidence = GREATEST(node_tags.confidence, EXCLUDED.confidence);
