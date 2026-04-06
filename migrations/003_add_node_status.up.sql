-- Add status column to nodes for processing pipeline tracking
ALTER TABLE nodes ADD COLUMN status TEXT NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'processing', 'processed', 'failed'));

-- Fix: use partial index on review_schedule (matches blueprint)
DROP INDEX IF EXISTS idx_review_due;
CREATE INDEX idx_review_due ON review_schedule(due_date)
    WHERE due_date <= now();

-- Index for efficient status filtering
CREATE INDEX idx_nodes_status ON nodes(status);
