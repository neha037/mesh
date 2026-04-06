-- name: UpsertEdge :exec
-- Create or update edge, keeping the higher weight.
INSERT INTO edges (source_id, target_id, rel_type, weight)
VALUES ($1, $2, $3, $4)
ON CONFLICT (source_id, target_id, rel_type) DO UPDATE
SET weight = GREATEST(edges.weight, EXCLUDED.weight);

-- name: BuildTagSharedEdges :exec
-- Create tag_shared edges for nodes sharing 2+ tags with the given node.
-- Weight = shared_count / total_tags_on_source_node (normalized 0-1).
INSERT INTO edges (source_id, target_id, rel_type, weight)
SELECT $1::uuid, nt2.node_id, 'tag_shared',
       COUNT(*)::real / NULLIF((SELECT COUNT(*) FROM node_tags WHERE node_id = $1), 0)
FROM node_tags nt1
JOIN node_tags nt2 ON nt1.tag_id = nt2.tag_id
WHERE nt1.node_id = $1 AND nt2.node_id != $1
GROUP BY nt2.node_id
HAVING COUNT(*) >= 2
ON CONFLICT (source_id, target_id, rel_type) DO UPDATE
SET weight = GREATEST(edges.weight, EXCLUDED.weight);

-- name: FindSimilarNodes :many
-- Find nodes with similar embeddings using pgvector cosine distance.
SELECT id, title, (1 - (embedding <=> sqlc.arg('embedding')::vector))::real AS similarity
FROM nodes
WHERE embedding IS NOT NULL AND id != sqlc.arg('exclude_id')::uuid
ORDER BY embedding <=> sqlc.arg('embedding')::vector
LIMIT $1;
