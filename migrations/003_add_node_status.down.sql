DROP INDEX IF EXISTS idx_nodes_status;

DROP INDEX IF EXISTS idx_review_due;
CREATE INDEX idx_review_due ON review_schedule(due_date);

ALTER TABLE nodes DROP COLUMN status;
