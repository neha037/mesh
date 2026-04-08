CREATE INDEX idx_nodes_status_updated ON nodes (status, updated_at)
WHERE status = 'processing';
