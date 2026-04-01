-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";        -- pgvector
CREATE EXTENSION IF NOT EXISTS "pg_trgm";       -- fuzzy text search

-- ============================================================
-- Core node representing any knowledge entity
-- ============================================================
CREATE TABLE nodes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type        TEXT NOT NULL CHECK (type IN (
                    'article', 'book', 'hobby', 'thought',
                    'journal', 'wildcard', 'image'
                )),
    title       TEXT NOT NULL,
    content     TEXT,                            -- full extracted text
    summary     TEXT,                            -- AI-generated summary
    source_url  TEXT,                            -- original URL if applicable
    image_key   TEXT,                            -- MinIO object key if applicable
    embedding   vector(384),                     -- nomic-embed-text dimension
    version     INTEGER NOT NULL DEFAULT 1,      -- optimistic concurrency control
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- Tags / concepts extracted by NLP
-- ============================================================
CREATE TABLE tags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE
);

-- ============================================================
-- Many-to-many: nodes <-> tags
-- ============================================================
CREATE TABLE node_tags (
    node_id     UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    tag_id      UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    confidence  REAL DEFAULT 1.0,                -- NLP confidence score
    PRIMARY KEY (node_id, tag_id)
);

-- ============================================================
-- Edges between nodes (multiple relationship types)
-- ============================================================
CREATE TABLE edges (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id   UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    target_id   UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    rel_type    TEXT NOT NULL CHECK (rel_type IN (
                    'tag_shared',   -- auto-created from shared tags
                    'manual',       -- user-created link
                    'semantic',     -- vector similarity bridge
                    'bridge',       -- cross-cluster discovery
                    'wildcard'      -- wildcard injection link
                )),
    weight      REAL NOT NULL DEFAULT 1.0,       -- relationship strength
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(source_id, target_id, rel_type)
);

-- ============================================================
-- FSRS spaced repetition scheduling state per node
-- ============================================================
CREATE TABLE review_schedule (
    node_id     UUID PRIMARY KEY REFERENCES nodes(id) ON DELETE CASCADE,
    stability   REAL NOT NULL DEFAULT 0.4,       -- FSRS stability parameter
    difficulty  REAL NOT NULL DEFAULT 5.0,       -- FSRS difficulty (1-10)
    due_date    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_review TIMESTAMPTZ,
    reps        INTEGER NOT NULL DEFAULT 0,      -- total successful reviews
    lapses      INTEGER NOT NULL DEFAULT 0,      -- times forgotten (rated "Again")
    state       TEXT NOT NULL DEFAULT 'new' CHECK (state IN (
                    'new', 'learning', 'review', 'relearning'
                ))
);

-- ============================================================
-- Background job queue (replaces external message queue)
-- ============================================================
CREATE TABLE jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type            TEXT NOT NULL CHECK (type IN (
                        'process_url', 'process_text', 'process_image',
                        'generate_embedding', 'build_edges',
                        'discovery_run', 'wildcard_inject',
                        'reembed_batch'
                    )),
    payload         JSONB NOT NULL,              -- job-specific parameters
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN (
                        'pending', 'running', 'done', 'failed', 'dead'
                    )),
    attempts        INTEGER NOT NULL DEFAULT 0,
    max_attempts    INTEGER NOT NULL DEFAULT 3,
    claimed_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    scheduled_for   TIMESTAMPTZ DEFAULT now(),   -- for delayed/scheduled jobs
    error           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- Discovery run history (for tracking and metrics)
-- ============================================================
CREATE TABLE discovery_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_type        TEXT NOT NULL CHECK (run_type IN (
                        'cluster_analysis', 'bridge_detection', 'wildcard_injection'
                    )),
    results         JSONB NOT NULL,              -- structured output of the run
    nodes_affected  INTEGER NOT NULL DEFAULT 0,
    edges_created   INTEGER NOT NULL DEFAULT 0,
    executed_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- Indexes
-- ============================================================

-- Vector similarity search (HNSW for fast approximate nearest neighbor)
CREATE INDEX idx_nodes_embedding ON nodes
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- Node lookups
CREATE INDEX idx_nodes_type ON nodes(type);
CREATE INDEX idx_nodes_created ON nodes(created_at DESC);
CREATE INDEX idx_nodes_title_trgm ON nodes USING gin (title gin_trgm_ops);
CREATE INDEX idx_nodes_content_trgm ON nodes USING gin (content gin_trgm_ops);

-- Edge traversal
CREATE INDEX idx_edges_source ON edges(source_id);
CREATE INDEX idx_edges_target ON edges(target_id);
CREATE INDEX idx_edges_rel_type ON edges(rel_type);

-- Job queue performance
CREATE INDEX idx_jobs_pending ON jobs(scheduled_for, created_at)
    WHERE status = 'pending';
CREATE INDEX idx_jobs_status ON jobs(status);

-- Review scheduling
CREATE INDEX idx_review_due ON review_schedule(due_date)
    WHERE due_date <= now();

-- Tag lookups
CREATE INDEX idx_tags_name ON tags(name);
