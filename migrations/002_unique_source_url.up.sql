-- Deduplicate existing rows: keep the newest node per source_url.
DELETE FROM nodes WHERE id IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (PARTITION BY source_url ORDER BY created_at DESC) AS rn
        FROM nodes WHERE source_url IS NOT NULL
    ) t WHERE rn > 1
);

-- Prevent future duplicates.
CREATE UNIQUE INDEX idx_nodes_source_url_unique
    ON nodes (source_url)
    WHERE source_url IS NOT NULL;
