# Mesh Data Model and Schema

## Entity-Relationship Diagram

```
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│   nodes     │       │  node_tags  │       │    tags      │
├─────────────┤       ├─────────────┤       ├─────────────┤
│ id (PK)     │──────<│ node_id(FK) │>──────│ id (PK)     │
│ type        │       │ tag_id (FK) │       │ name (UQ)   │
│ title       │       └─────────────┘       └─────────────┘
│ content     │
│ summary     │       ┌─────────────────┐
│ source_url  │       │     edges       │
│ image_key   │       ├─────────────────┤
│ embedding   │──────<│ source_id (FK)  │
│ version     │       │ target_id (FK)  │>──── nodes.id
│ created_at  │       │ rel_type        │
│ updated_at  │       │ weight          │
│ updated_at  │       │ created_at      │
└─────────────┘       └─────────────────┘
      │               
      │               
      │               ┌─────────────────┐
      └──────────────<│ review_schedule │
                      ├─────────────────┤
                      │ node_id (PK,FK) │
                      │ stability       │
                      │ difficulty      │
                      │ due_date        │
                      │ last_review     │
                      │ reps            │
                      │ lapses          │
                      └─────────────────┘

┌─────────────────┐
│      jobs       │
├─────────────────┤
│ id (PK)         │
│ type            │
│ payload (JSONB) │
│ status          │
│ claimed_at      │
│ completed_at    │
│ error           │
│ created_at      │
└─────────────────┘
```

## SQL Schema

```sql
-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";        -- pgvector
CREATE EXTENSION IF NOT EXISTS "pg_trgm";       -- fuzzy text search

-- Core node representing any knowledge entity
CREATE TABLE nodes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type        TEXT NOT NULL CHECK (type IN (
                    'article', 'book', 'hobby', 'thought',
                    'journal', 'wildcard', 'image'
                )),
    title       TEXT NOT NULL,
    content     TEXT,
    summary     TEXT,
    source_url  TEXT,
    image_key   TEXT,
    embedding   vector(768),
    version     INTEGER NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Tags / concepts extracted by NLP
CREATE TABLE tags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE
);

-- Many-to-many: nodes <-> tags
CREATE TABLE node_tags (
    node_id     UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    tag_id      UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    confidence  REAL DEFAULT 1.0,
    PRIMARY KEY (node_id, tag_id)
);

-- Edges between nodes
CREATE TABLE edges (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id   UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    target_id   UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    rel_type    TEXT NOT NULL CHECK (rel_type IN (
                    'tag_shared', 'manual', 'semantic', 'bridge', 'wildcard'
                )),
    weight      REAL NOT NULL DEFAULT 1.0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(source_id, target_id, rel_type)
);

-- FSRS spaced repetition scheduling
CREATE TABLE review_schedule (
    node_id     UUID PRIMARY KEY REFERENCES nodes(id) ON DELETE CASCADE,
    stability   REAL NOT NULL DEFAULT 0.4,
    difficulty  REAL NOT NULL DEFAULT 5.0,
    due_date    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_review TIMESTAMPTZ,
    reps        INTEGER NOT NULL DEFAULT 0,
    lapses      INTEGER NOT NULL DEFAULT 0,
    state       TEXT NOT NULL DEFAULT 'new' CHECK (state IN (
                    'new', 'learning', 'review', 'relearning'
                ))
);

-- Background job queue
CREATE TABLE jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type            TEXT NOT NULL CHECK (type IN (
                        'process_url', 'process_text', 'process_image',
                        'generate_embedding', 'build_edges',
                        'discovery_run', 'wildcard_inject',
                        'reembed_batch'
                    )),
    payload         JSONB NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN (
                        'pending', 'running', 'done', 'failed', 'dead'
                    )),
    attempts        INTEGER NOT NULL DEFAULT 0,
    max_attempts    INTEGER NOT NULL DEFAULT 3,
    claimed_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    scheduled_for   TIMESTAMPTZ DEFAULT now(),
    error           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Discovery run history
CREATE TABLE discovery_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_type        TEXT NOT NULL CHECK (run_type IN (
                        'cluster_analysis', 'bridge_detection', 'wildcard_injection'
                    )),
    results         JSONB NOT NULL,
    nodes_affected  INTEGER NOT NULL DEFAULT 0,
    edges_created   INTEGER NOT NULL DEFAULT 0,
    executed_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## Key Query Patterns

### BFS Graph Traversal
```sql
WITH RECURSIVE graph_walk AS (
    SELECT target_id AS node_id, 1 AS depth, ARRAY[source_id, target_id] AS path
    FROM edges WHERE source_id = $1
    UNION ALL
    SELECT e.target_id, gw.depth + 1, gw.path || e.target_id
    FROM edges e
    JOIN graph_walk gw ON e.source_id = gw.node_id
    WHERE gw.depth < $2 AND NOT (e.target_id = ANY(gw.path))
)
SELECT DISTINCT ON (n.id) n.*, gw.depth
FROM graph_walk gw
JOIN nodes n ON n.id = gw.node_id
ORDER BY n.id, gw.depth;
```

### Semantic Similarity Search
```sql
SELECT id, title, type, summary, 1 - (embedding <=> $1) AS similarity
FROM nodes WHERE embedding IS NOT NULL AND id != $2
ORDER BY embedding <=> $1 LIMIT $3;
```
