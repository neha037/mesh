# Project Mesh: Production-Ready Architectural Blueprint

**Version:** 1.0  
**Date:** March 31, 2026  
**Author:** neha037  
**Status:** Pre-Implementation Planning Complete

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Market and Feature Analysis](#2-market-and-feature-analysis)
3. [Technical Stack Recommendation](#3-technical-stack-recommendation)
4. [Architectural Design](#4-architectural-design)
5. [Data Model and Schema](#5-data-model-and-schema)
6. [Implementation Roadmap](#6-implementation-roadmap)
7. [Risk Assessment and Mitigation](#7-risk-assessment-and-mitigation)
8. [Project Structure](#8-project-structure)
9. [Docker Compose Topology](#9-docker-compose-topology)
10. [Appendix: Key Algorithms](#10-appendix-key-algorithms)
11. [References and Research Sources](#11-references-and-research-sources)

---

## 1. Executive Summary

### Mission

Project Mesh is a **localized, private Personal Growth Engine** that maps both structured technical knowledge (system design, algorithms, cryptography) and fluid creative pursuits (acrylic painting, reading, journaling) into a unified, interactive topological space. It actively combats intellectual stagnation by visualizing knowledge clusters and algorithmically injecting serendipitous discovery.

### Core Differentiator

Unlike passive PKM (Personal Knowledge Management) tools that serve as repositories, Mesh is an **active cognitive partner**. It operationalizes Stuart Kauffman's concept of the "Adjacent Possible" — the set of novel cognitive leaps exactly one step from the user's current knowledge boundary — to autonomously bridge disparate domains and surface surprising connections.

### Design Constraints

- **Zero ongoing cloud costs** — all compute and storage runs locally
- **Ephemeral compute** — spin up only when actively used, safe to shut down at any time
- **Single developer** — 6-8 hours/week of sustainable development effort
- **Absolute data sovereignty** — no data leaves the local machine
- **Standard consumer hardware** — must run on a 16 GB RAM developer workstation

---

## 2. Market and Feature Analysis

### 2.1 Competitive Landscape

| Product | Paradigm | Strengths | Weaknesses | Storage | Price |
|---------|----------|-----------|------------|---------|-------|
| **Obsidian** | Document-first, networked thought | 2,600+ plugins, customizable graph view, handles 5K+ notes, fast | Graph clutters above ~500 notes; no active discovery; no serendipity | Local Markdown vaults | Free (Sync $8/mo) |
| **Logseq** | Outliner-first, block references | Open source (AGPL), built-in spaced repetition, daily journals | 2-3s load time on large vaults; passive linking only; smaller plugin ecosystem (~300) | Local-first | Free |
| **Heptabase** | Visual-first infinite canvas | Spatial sense-making, deep PDF extraction, intuitive organization | Cloud-synced subscription; no cross-domain discovery engine | Cloud-synced | Subscription |
| **Tana** | Supertag ontology | Deep data automation, structured semantic schemas | Cloud-locked; no serendipity; requires continuous connectivity | Cloud-synced | Subscription |
| **Mem.ai** | AI-first auto-organization | Auto-categorization, contextual recommendations, low friction | Optimizes for engagement similarity (echo chamber risk); proprietary AI; cloud-dependent | Cloud-synced | Subscription |

### 2.2 The Critical Gap: Algorithmic Serendipity

**Every existing tool is passive.** They rely entirely on the user to manually create connections, tag content, and review older material. Even AI-enhanced tools primarily focus on auto-tagging or summarization — they suggest more of the same, which reinforces echo chambers.

**Serendipity** in knowledge discovery is defined as finding valuable, relevant information that was **not explicitly sought** (Kotkov et al.). Pure recommendation algorithms optimize for immediate engagement by suggesting highly similar content, which leads to over-specialization.

**Mesh's approach:**

- Treat the knowledge graph as a **dynamic ecosystem**, not a static repository
- Use **vector embeddings** (pgvector, cosine similarity) to find semantically related but lexically distinct nodes
- Run **cluster density analysis** to identify over-saturated vs. isolated knowledge regions
- Inject **"Wildcard" topics** from orthogonal domains weekly
- Measure success via **Kotkov's Serendipity Metric**: the intersection of recommended items that are highly relevant to the user but fundamentally dissimilar to items the user has previously interacted with

### 2.3 Feature Stratification: MVP vs. Delighters

#### Must-Have MVP (Phases 1-4)

| Feature | Description | Justification |
|---------|-------------|---------------|
| REST Ingestion API | Endpoints accepting URLs and raw text payloads | Foundational sensory input mechanism for all subsequent processing |
| HTML Stripping + NLP Tagging | Automated concept extraction from unstructured text | Eliminates manual data entry fatigue, keeps system frictionless |
| Graph Storage + Traversal | Nodes, edges, BFS/DFS queries via recursive CTEs | Core memory and relational structure for networked thought |
| Global Interactive Visualization | Physics-based web rendering of the full graph topology | Validates the UI and provides immediate visual feedback of clusters |
| Search and Filter | Query by tag, date, node type | Essential navigation for a growing knowledge base |

#### Long-Term Delighters (Phases 5-7)

| Feature | Description | Justification |
|---------|-------------|---------------|
| Multi-Modal Expansion | Image uploads (paintings, whiteboards) via MinIO | Bridges digital text and physical creative pursuits |
| Manual Journaling | Rich text brain-dump UI, auto-tagged | Supports unstructured thought without requiring a URL source |
| Anti-Echo Chamber Engine | Cluster density analysis + wildcard topic injection | Core serendipity value proposition |
| FSRS Spaced Repetition | Advanced scheduling to resurface forgotten nodes | Prevents knowledge decay using memory retention curves |
| Semantic Cross-Pollination | Vector embedding similarity search across unconnected nodes | Simulates human intuition by connecting conceptually related, lexically distinct nodes |

---

## 3. Technical Stack Recommendation

### 3.1 Backend: Go (Golang)

**Justification:**

- **Compiled binary**: ~10 MB memory footprint per service, no runtime dependency
- **Goroutine concurrency**: Spin up hundreds of async workers for scraping and NLP at negligible cost (goroutines start at ~2 KB stack vs. OS threads at ~1 MB)
- **Native `context` package**: Built-in timeout enforcement for external HTTP calls
- **First-class Kubernetes tooling**: `client-go`, Helm charts, operator SDKs
- **No GIL bottleneck**: Unlike Python, true parallel execution for CPU-bound embedding generation
- **Developer alignment**: Matches the project owner's primary specialization

**Key Go Libraries:**

| Library | Purpose | Notes |
|---------|---------|-------|
| `go-chi/chi` | HTTP router | Lightweight, idiomatic, middleware-composable |
| `sqlc-dev/sqlc` | Type-safe SQL codegen | Generates Go structs and methods from SQL queries |
| `jackc/pgx` | PostgreSQL driver | High-performance, connection pooling, pgvector support |
| `gocolly/colly` | Web scraping | Async, rate-limiting built-in, respects robots.txt |
| `PuerkitoBio/goquery` | HTML parsing | jQuery-like syntax for DOM traversal |
| `sony/gobreaker` | Circuit breaker | Prevents cascading failures on external API calls |
| `pgvector/pgvector-go` | Vector operations | Native pgvector types for pgx/GORM/sqlx |
| `xyproto/ollamaclient` | Ollama LLM client | Text generation, summarization, embeddings |
| `jonathanhecl/gollama` | Ollama wrapper | Structured outputs, function calling, vision support |
| `jdkato/prose` | Pure-Go NLP | Tokenization, POS tagging, NER (fallback when Ollama unavailable) |
| `dgraph-io/ristretto` | In-process cache | Concurrent, admission/eviction policies |
| `minio/minio-go` | MinIO S3 client | Upload, download, presigned URLs |
| `golang-migrate/migrate` | DB migrations | SQL file-based, reversible migrations |
| `testcontainers/testcontainers-go` | Integration testing | Spin up real PostgreSQL/MinIO in tests |

### 3.2 Database: PostgreSQL 16 + pgvector

**Decision: PostgreSQL over Dgraph (dedicated graph DB)**

This is the single most important architectural decision. Research strongly favors PostgreSQL for personal-scale knowledge graphs:

**Evidence:**
- **Capacities** (a tool-for-thought platform with production traffic) migrated FROM Dgraph TO PostgreSQL in 2026, achieving a **70% infrastructure cost reduction**. Dgraph consumed unpredictably high CPU even with modest datasets, forcing difficult scaling decisions.
- Multiple teams have independently built knowledge graphs with PostgreSQL recursive CTEs, reporting microsecond-latency traversals for graphs under 10,000 nodes.
- PostgreSQL's `pg_trgm` extension provides fuzzy text search, `pgvector` provides HNSW-indexed vector similarity, and recursive CTEs provide graph traversal — **all in one process**.

**Why NOT Dgraph:**
- Additional container, separate query language (DQL/GraphQL), separate backup strategy
- Unpredictable CPU usage even at small scale
- Limited Go ORM support compared to PostgreSQL ecosystem
- Personal knowledge graph will realistically stay under 10,000 nodes and 50,000 edges — well within PostgreSQL's sweet spot for recursive queries

**Why NOT Neo4j:**
- Java-based, heavy memory footprint (minimum 2 GB heap)
- Community edition has clustering limitations
- Cypher query language adds cognitive overhead when the team already knows SQL
- Overkill for personal-scale graph with <10K nodes

**pgvector capabilities:**
- HNSW (Hierarchical Navigable Small World) indexing for approximate nearest neighbor search
- Supports `vector(384)` type natively (matches `nomic-embed-text` embedding dimension)
- Cosine distance operator `<=>` for similarity queries
- Integrates with standard `sqlc`/`pgx` tooling via `pgvector-go`

### 3.3 Caching Layer: In-Process (ristretto) — No Redis

**Justification:** For a **single-user local system**, Redis adds:
- An extra container consuming ~30 MB base RAM
- Network hop latency for every cache operation
- Operational complexity (persistence config, eviction tuning)

The `dgraph-io/ristretto` library provides:
- Thread-safe concurrent cache with admission policy (TinyLFU)
- Configurable max cost (memory budget)
- ~10ns read latency (vs. ~100μs for Redis over loopback)

**Migration path:** If multi-user or distributed deployment ever becomes necessary, replacing ristretto with Redis is a straightforward interface swap behind a `cache.Store` interface.

### 3.4 Object Storage: MinIO

- S3-compatible API (100% compatible with AWS SDK)
- Written in Go, single container, minimal resource usage
- Stores: photos of physical canvases, whiteboard diagrams, PDF attachments
- Go client: `minio-go` SDK for upload, download, presigned URL generation
- Docker volume-backed for persistence

### 3.5 NLP / AI Layer: Ollama + Local Models

**Architecture:** Ollama runs as a Docker container serving local LLMs. The Go backend communicates via HTTP API.

| Model | Purpose | Size (Q4) | RAM Required |
|-------|---------|-----------|-------------|
| `nomic-embed-text` | Generate 384-dim embeddings for semantic search | ~270 MB | ~500 MB |
| `mistral:7b-instruct-q4_0` | Concept extraction, summarization, tag generation | ~4 GB | ~6 GB |

**Why local models over cloud APIs:**
- **Privacy**: Zero data leaves the machine — non-negotiable for personal knowledge
- **Cost**: No per-token charges; unlimited inference after one-time model download
- **Latency**: ~200ms for embeddings, ~2-5s for tag extraction (acceptable for async background processing)
- **Availability**: Works offline, no API key management

**Fallback strategy:** When Ollama is not running (to save RAM), the system falls back to:
- `jdkato/prose` for tokenization and named entity recognition
- TF-IDF-based keyword extraction implemented in pure Go
- Embeddings are skipped and generated in batch when Ollama is next available

**Ollama Docker profile:** Configured with Docker Compose `profiles: ["ai"]` so it only starts when explicitly requested (`docker compose --profile ai up`), freeing RAM for other tasks.

### 3.6 Frontend: React + TypeScript + Cytoscape.js

**React** (not Vue) — chosen for:
- Larger component ecosystem and community support
- TypeScript-first tooling with Vite
- TanStack Query for server state management (caching, background refetching)
- Better Cytoscape.js React wrapper (`react-cytoscapejs`)

**Cytoscape.js** — chosen over D3.js and React Flow for:
- **Built-in graph algorithms**: BFS, DFS, shortest path, betweenness centrality, community detection (Louvain), PageRank — useful for cluster analysis on the client side
- **Multiple layout engines**: Force-directed (CoSE), hierarchical (dagre), circular, grid, concentric — switchable at runtime
- **Performance**: Canvas rendering handles 2-5K nodes; WebGL extension (`cytoscape-webgl`) available for larger graphs
- **Rich styling**: CSS-like selectors for node/edge appearance based on data attributes
- **Event system**: Click, hover, drag events with full node data access

**Tailwind CSS** — utility-first styling for rapid, consistent UI development without maintaining a separate CSS architecture.

### 3.7 Orchestration: Docker Compose (MVP) → K3s (Optional Future)

**Phase 1-5: Docker Compose**

Docker Compose handles the 5-6 service topology comfortably for a single-node personal project:
- Simple `docker compose up` / `docker compose down` lifecycle
- Named volumes for persistent data
- Service dependency management
- Profile-based optional services (Ollama)
- Handles <20 services without issue

**Phase 6+ (Optional): K3s Migration**

Only migrate to K3s when:
- You want zero-downtime rolling deployments for the web UI
- You want to practice production Kubernetes patterns with real workloads
- You need CronJob resources for scheduled discovery tasks
- Service count exceeds 15

K3s specifics:
- ~512 MB RAM overhead, single binary <100 MB
- Includes Traefik ingress, CoreDNS, local-path-provisioner out of the box
- Installs in <30 seconds on Linux
- Full Kubernetes API compatibility

---

## 4. Architectural Design

### 4.1 System Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend (React)                         │
│  ┌──────────┐  ┌──────────┐  ┌───────────┐  ┌───────────────┐  │
│  │ Graph    │  │ Search/  │  │ Node      │  │ Review        │  │
│  │ Canvas   │  │ Filter   │  │ Detail    │  │ Card          │  │
│  │(Cytoscape│  │ Panel    │  │ Side Panel│  │ (FSRS)        │  │
│  └──────────┘  └──────────┘  └───────────┘  └───────────────┘  │
│                          │ REST / JSON                          │
└──────────────────────────┼──────────────────────────────────────┘
                           │
┌──────────────────────────┼──────────────────────────────────────┐
│                    Ingestion API (Go)              Port 8080    │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐    │
│  │ POST /ingest │  │ GET /graph   │  │ GET /search        │    │
│  │ POST /upload │  │ GET /nodes   │  │ GET /review/today  │    │
│  └──────┬───────┘  └──────────────┘  └────────────────────┘    │
│         │ enqueue job                                           │
└─────────┼───────────────────────────────────────────────────────┘
          │
┌─────────┼───────────────────────────────────────────────────────┐
│         ▼           Background Workers (Go)                     │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐    │
│  │ Scraper      │  │ NLP Tagger   │  │ Embedding          │    │
│  │ (colly)      │  │ (Ollama)     │  │ Generator          │    │
│  └──────────────┘  └──────────────┘  └────────────────────┘    │
│  ┌──────────────┐  ┌──────────────┐                             │
│  │ Edge Builder │  │ Catch-Up     │                             │
│  │              │  │ Cron Runner  │                             │
│  └──────────────┘  └──────────────┘                             │
└─────────────────────────────────────────────────────────────────┘
          │                    │                    │
          ▼                    ▼                    ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────────┐
│ PostgreSQL   │    │ MinIO        │    │ Ollama           │
│ + pgvector   │    │ (S3 Objects) │    │ (Local LLM)      │
│              │    │              │    │                   │
│ • nodes      │    │ • images     │    │ • nomic-embed     │
│ • edges      │    │ • canvases   │    │ • mistral:7b      │
│ • tags       │    │ • PDFs       │    │                   │
│ • jobs       │    │              │    │                   │
│ • reviews    │    │              │    │                   │
└──────────────┘    └──────────────┘    └──────────────────┘
     Volume:             Volume:             Volume:
     pgdata/           minio_data/        ollama_models/
```

### 4.2 Service Interaction Flow

#### Ingestion Flow (URL → Knowledge Node)

```
1. User submits URL via UI or curl
2. API validates input, creates raw node (status: pending)
3. API inserts job into `jobs` table (type: "process_url")
4. Worker claims job via SELECT ... FOR UPDATE SKIP LOCKED
5. Worker fetches URL content (colly, 30s timeout, circuit breaker)
6. Worker strips HTML → clean text (goquery)
7. Worker calls Ollama: extract concepts → JSON array of tags
8. Worker calls Ollama: generate embedding → 384-dim vector
9. Worker UPSERTs tags (ON CONFLICT DO NOTHING)
10. Worker creates node_tags associations
11. Worker auto-generates edges to nodes sharing 2+ tags
12. Worker updates node status to "processed", stores summary
13. Worker marks job as "done"
```

#### Graph Query Flow

```
1. Frontend requests GET /api/v1/graph?depth=2&center=<node_id>
2. API executes recursive CTE traversal (depth-limited BFS)
3. API enriches nodes with tag lists and edge metadata
4. API returns JSON: { nodes: [...], edges: [...] }
5. Frontend renders via Cytoscape.js with CoSE layout
```

#### Discovery Flow (Anti-Echo Chamber)

```
1. Discovery engine runs on schedule (or catch-up on startup)
2. Cluster density analysis:
   a. Group nodes by dominant tags
   b. Count nodes per cluster
   c. Identify over-saturated clusters (>N nodes)
   d. Identify isolated nodes (<M connections)
3. Adjacent Possible detection:
   a. For each pair of distinct clusters, find the pair of nodes
      (one from each) with highest vector similarity but no direct edge
   b. Suggest these as "bridge reading"
4. Wildcard injection:
   a. Pick a random high-quality topic from external sources
   b. Create a seed node with type "wildcard"
   c. Generate embedding, find nearest existing nodes
   d. Create lightweight "wildcard" edges
```

### 4.3 API Endpoint Specification

#### Ingestion Endpoints

| Method | Path | Description | Request Body |
|--------|------|-------------|-------------|
| `POST` | `/api/v1/ingest/url` | Submit a URL for scraping and processing | `{ "url": "https://...", "type": "article" }` |
| `POST` | `/api/v1/ingest/text` | Submit raw text (thought, note) | `{ "title": "...", "content": "...", "type": "thought" }` |
| `POST` | `/api/v1/ingest/image` | Upload an image file | `multipart/form-data` with image + metadata JSON |
| `POST` | `/api/v1/ingest/journal` | Submit a journal entry | `{ "content": "...", "mood": "..." }` |

#### Query Endpoints

| Method | Path | Description | Query Params |
|--------|------|-------------|-------------|
| `GET` | `/api/v1/graph` | Full graph or subgraph | `?center=<uuid>&depth=<int>&types=<csv>` |
| `GET` | `/api/v1/nodes` | List nodes with pagination | `?page=&limit=&type=&tag=&from=&to=` |
| `GET` | `/api/v1/nodes/:id` | Get single node with edges | — |
| `GET` | `/api/v1/nodes/:id/similar` | Semantic similarity search | `?limit=<int>&threshold=<float>` |
| `GET` | `/api/v1/search` | Full-text + semantic search | `?q=<text>&mode=text|semantic|hybrid` |
| `GET` | `/api/v1/tags` | List all tags with counts | `?sort=count|alpha` |
| `GET` | `/api/v1/clusters` | Cluster density report | — |

#### FSRS / Review Endpoints

| Method | Path | Description | Request Body |
|--------|------|-------------|-------------|
| `GET` | `/api/v1/review/today` | Get today's review node | — |
| `POST` | `/api/v1/review/:node_id` | Submit review rating | `{ "rating": 1-4 }` |
| `GET` | `/api/v1/review/stats` | Review history and metrics | — |

#### Discovery Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/discovery/bridges` | Get suggested bridge readings |
| `GET` | `/api/v1/discovery/wildcards` | Get recent wildcard injections |
| `POST` | `/api/v1/discovery/trigger` | Manually trigger discovery run |

#### Node CRUD

| Method | Path | Description |
|--------|------|-------------|
| `PUT` | `/api/v1/nodes/:id` | Update node (title, content, tags) |
| `DELETE` | `/api/v1/nodes/:id` | Delete node and associated edges |
| `POST` | `/api/v1/nodes/:id/edges` | Manually create an edge |
| `DELETE` | `/api/v1/edges/:id` | Delete an edge |

---

## 5. Data Model and Schema

### 5.1 Entity-Relationship Diagram

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
└─────────────┘       │ created_at      │
      │               └─────────────────┘
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

### 5.2 Complete SQL Schema

```sql
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
```

### 5.3 Key Query Patterns

#### BFS Graph Traversal (N-hop neighborhood)

```sql
WITH RECURSIVE graph_walk AS (
    -- Base case: direct neighbors
    SELECT
        target_id AS node_id,
        1 AS depth,
        ARRAY[source_id, target_id] AS path
    FROM edges
    WHERE source_id = $1  -- starting node UUID

    UNION ALL

    -- Recursive case: walk further
    SELECT
        e.target_id,
        gw.depth + 1,
        gw.path || e.target_id
    FROM edges e
    JOIN graph_walk gw ON e.source_id = gw.node_id
    WHERE gw.depth < $2                       -- max depth parameter
      AND NOT (e.target_id = ANY(gw.path))    -- prevent cycles
)
SELECT DISTINCT ON (n.id)
    n.*,
    gw.depth
FROM graph_walk gw
JOIN nodes n ON n.id = gw.node_id
ORDER BY n.id, gw.depth;
```

#### Semantic Similarity Search

```sql
SELECT
    id, title, type, summary,
    1 - (embedding <=> $1) AS similarity  -- cosine similarity
FROM nodes
WHERE embedding IS NOT NULL
  AND id != $2                             -- exclude the query node
ORDER BY embedding <=> $1                  -- order by cosine distance
LIMIT $3;                                  -- top N results
```

#### Job Queue Claim (Lock-Free)

```sql
UPDATE jobs
SET status = 'running',
    claimed_at = now(),
    attempts = attempts + 1
WHERE id = (
    SELECT id FROM jobs
    WHERE status = 'pending'
      AND scheduled_for <= now()
      AND attempts < max_attempts
    ORDER BY created_at
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING *;
```

#### Cluster Density Analysis

```sql
WITH tag_clusters AS (
    SELECT
        t.id AS tag_id,
        t.name AS tag_name,
        COUNT(DISTINCT nt.node_id) AS node_count
    FROM tags t
    JOIN node_tags nt ON nt.tag_id = t.id
    GROUP BY t.id, t.name
),
cluster_stats AS (
    SELECT
        AVG(node_count) AS avg_size,
        STDDEV(node_count) AS stddev_size
    FROM tag_clusters
)
SELECT
    tc.tag_name,
    tc.node_count,
    CASE
        WHEN tc.node_count > cs.avg_size + cs.stddev_size THEN 'over_saturated'
        WHEN tc.node_count < cs.avg_size - cs.stddev_size THEN 'under_explored'
        ELSE 'balanced'
    END AS cluster_health
FROM tag_clusters tc
CROSS JOIN cluster_stats cs
ORDER BY tc.node_count DESC;
```

#### Find Bridge Candidates (Adjacent Possible)

```sql
WITH cluster_a_nodes AS (
    SELECT nt.node_id FROM node_tags nt
    JOIN tags t ON t.id = nt.tag_id
    WHERE t.name = $1  -- cluster A tag
),
cluster_b_nodes AS (
    SELECT nt.node_id FROM node_tags nt
    JOIN tags t ON t.id = nt.tag_id
    WHERE t.name = $2  -- cluster B tag
)
SELECT
    a.id AS node_a_id,
    a.title AS node_a_title,
    b.id AS node_b_id,
    b.title AS node_b_title,
    1 - (a.embedding <=> b.embedding) AS similarity
FROM nodes a
JOIN cluster_a_nodes ca ON ca.node_id = a.id
CROSS JOIN nodes b
JOIN cluster_b_nodes cb ON cb.node_id = b.id
WHERE a.embedding IS NOT NULL
  AND b.embedding IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM edges e
      WHERE e.source_id = a.id AND e.target_id = b.id
  )
ORDER BY a.embedding <=> b.embedding  -- closest in embedding space
LIMIT 5;
```

---

## 6. Implementation Roadmap

### Overview

The project is divided into **7 phases** spanning approximately **30 weeks** at a sustainable pace of **6-8 hours per week**. Each phase delivers a working, standalone increment — no phase depends on later phases for utility.

```
Phase 1 ──► Phase 2 ──► Phase 3 ──► Phase 4 ──► Phase 5 ──► Phase 6 ──► Phase 7
 Senses      Brain      Memory       Eyes       Human     Discovery    Slow Burn
 Wk 1-3     Wk 4-6     Wk 7-9     Wk 10-14   Wk 15-18   Wk 19-24    Wk 25-30
   │           │           │           │           │           │           │
   ▼           ▼           ▼           ▼           ▼           ▼           ▼
  MVP ─────────────────────────────────┘     Delighters ────────────────────┘
```

---

### Phase 1: Foundation and Ingestion — "The Senses"

**Timeline:** Weeks 1-3 (18-24 hours)  
**Goal:** A working Go API that accepts data and persists it, running in Docker Compose.

#### Week 1: Project Scaffolding

- [ ] Initialize Go module: `go mod init github.com/neha037/mesh`
- [ ] Create project directory structure (see Section 8)
- [ ] Set up Docker Compose with PostgreSQL 16 + pgvector
- [ ] Write initial SQL migration (`001_initial_schema.up.sql`) with `nodes`, `tags`, `node_tags`, `jobs` tables
- [ ] Configure `golang-migrate` for migration management
- [ ] Set up `Makefile` with targets: `build`, `run`, `migrate-up`, `migrate-down`, `test`, `docker-up`, `docker-down`
- [ ] Create `Dockerfile.api` (multi-stage build: Go builder → scratch/alpine)
- [ ] Initialize Git repository, add `.gitignore`

#### Week 2: Ingestion API

- [ ] Implement HTTP server with `chi` router
- [ ] `POST /api/v1/ingest/url` endpoint:
  - Validate URL format
  - Create node with status `pending`
  - Enqueue `process_url` job
  - Return `202 Accepted` with node ID
- [ ] `POST /api/v1/ingest/text` endpoint:
  - Validate required fields (title, content)
  - Create node directly with status `raw`
  - Enqueue `process_text` job
  - Return `201 Created` with node ID
- [ ] Implement PostgreSQL repository layer using `pgx` + `sqlc`
- [ ] Add request logging middleware
- [ ] Add CORS middleware (for frontend development)

#### Week 3: Resilience and Testing

- [ ] Implement web scraper using `colly`:
  - `context.WithTimeout` (30 seconds) for all external HTTP calls
  - User-Agent rotation
  - Respect `robots.txt`
- [ ] Integrate `sony/gobreaker` circuit breaker:
  - Open after 5 consecutive failures
  - Half-open probe after 60 seconds
  - Configurable per-domain
- [ ] Job queue claim logic: `SELECT ... FOR UPDATE SKIP LOCKED`
- [ ] Write unit tests for handlers (table-driven)
- [ ] Write integration tests using `testcontainers-go` (real PostgreSQL)
- [ ] Verify end-to-end: `curl` a URL → job created → scraper fetches content → raw text stored

**Phase 1 Deliverable:** Running `curl -X POST localhost:8080/api/v1/ingest/url -d '{"url":"https://example.com"}'` creates a node and enqueues a processing job.

---

### Phase 2: Processing and Intelligence — "The Brain"

**Timeline:** Weeks 4-6 (18-24 hours)  
**Goal:** Background workers that extract content, generate tags and embeddings automatically.

#### Week 4: Worker Pool and HTML Processing

- [ ] Implement worker pool:
  - Configurable worker count (default: 4 goroutines)
  - Each goroutine runs a claim-process-complete loop
  - Graceful shutdown via `context.Done()`
  - Exponential backoff when no jobs available (1s → 2s → 4s → 8s → 30s cap)
- [ ] HTML stripping pipeline:
  - Fetch raw HTML (if URL job)
  - Parse with `goquery`
  - Extract `<title>`, `<article>`, `<main>`, or `<body>` content
  - Strip script/style tags
  - Clean whitespace, decode HTML entities
  - Store clean text in `nodes.content`
- [ ] Add `Dockerfile.worker` for the worker service
- [ ] Add worker service to Docker Compose

#### Week 5: NLP Integration via Ollama

- [ ] Add Ollama container to Docker Compose (with `profiles: ["ai"]`)
- [ ] Implement Ollama client in Go:
  - HTTP client to `http://ollama:11434/api/generate` for tag extraction
  - HTTP client to `http://ollama:11434/api/embeddings` for vector generation
  - Retry logic with circuit breaker
- [ ] Tag extraction prompt engineering:

  ```
  Extract 3-8 core concepts/topics from this text as a JSON array of strings.
  Focus on domain-specific terms, not generic words.
  Return ONLY the JSON array, no explanation.

  Text: {content}
  ```

- [ ] Embedding generation:
  - Model: `nomic-embed-text` (384 dimensions)
  - Store result in `nodes.embedding` column
- [ ] UPSERT logic for tags:

  ```sql
  INSERT INTO tags (name) VALUES ($1)
  ON CONFLICT (name) DO NOTHING
  RETURNING id;
  ```

- [ ] Create `node_tags` associations from extracted tags

#### Week 6: Edge Building and Concurrency Control

- [ ] Auto-edge generation algorithm:
  - After processing a node, find all other nodes sharing 2+ tags
  - Create `tag_shared` edges with weight proportional to shared tag count
  - Use UPSERT to avoid duplicate edges
- [ ] Optimistic Concurrency Control:
  - All node updates include `WHERE version = $expected_version`
  - On conflict, re-read node and retry (max 3 attempts)
  - Log warning on version conflicts for monitoring
- [ ] Fallback NLP path (when Ollama is unavailable):
  - Use `jdkato/prose` for named entity recognition
  - Extract nouns and proper nouns as tags
  - Skip embedding generation (queue for later batch processing)
- [ ] Integration tests for the full pipeline
- [ ] Monitor: log worker throughput (jobs/minute), Ollama latency

**Phase 2 Deliverable:** Ingest a URL → worker scrapes it → AI extracts tags and embedding → edges auto-created to related nodes.

---

### Phase 3: Graph Traversal and Query API — "The Memory"

**Timeline:** Weeks 7-9 (18-24 hours)  
**Goal:** Rich query API exposing graph traversal, search, and similarity.

#### Week 7: Graph Traversal Endpoints

- [ ] Implement recursive CTE-based traversal:
  - `GET /api/v1/graph?center=<uuid>&depth=<int>` — BFS subgraph
  - `GET /api/v1/graph` (no center) — full graph export (with pagination for large graphs)
  - Cycle detection via path tracking in CTE
  - Maximum depth limit: 5 (configurable, prevents runaway queries)
- [ ] Response format:

  ```json
  {
    "nodes": [
      {
        "id": "uuid",
        "type": "article",
        "title": "...",
        "summary": "...",
        "tags": ["tag1", "tag2"],
        "created_at": "2026-03-31T...",
        "connection_count": 5
      }
    ],
    "edges": [
      {
        "id": "uuid",
        "source": "uuid",
        "target": "uuid",
        "rel_type": "tag_shared",
        "weight": 2.0
      }
    ],
    "meta": {
      "total_nodes": 150,
      "total_edges": 340,
      "depth": 2
    }
  }
  ```

#### Week 8: Search and Similarity

- [ ] Full-text search: `GET /api/v1/search?q=<text>&mode=text`
  - Uses `pg_trgm` trigram index for fuzzy matching
  - Searches title and content fields
  - Returns ranked results with similarity score
- [ ] Semantic search: `GET /api/v1/search?q=<text>&mode=semantic`
  - Generate embedding for query text (via Ollama)
  - Find nearest neighbors via pgvector cosine distance
  - Return top N results with similarity score
- [ ] Hybrid search: `GET /api/v1/search?q=<text>&mode=hybrid`
  - Run both text and semantic search
  - Reciprocal Rank Fusion (RRF) to merge results
- [ ] Node similarity: `GET /api/v1/nodes/:id/similar?limit=10`
  - Uses the node's stored embedding
  - Cosine distance against all other node embeddings

#### Week 9: CRUD and Filtering

- [ ] Node CRUD operations:
  - `GET /api/v1/nodes/:id` — single node with all edges and tags
  - `PUT /api/v1/nodes/:id` — update node (with OCC version check)
  - `DELETE /api/v1/nodes/:id` — soft delete or cascade
- [ ] Edge management:
  - `POST /api/v1/nodes/:id/edges` — manually create edge
  - `DELETE /api/v1/edges/:id` — remove edge
- [ ] Filtering and pagination:
  - `GET /api/v1/nodes?type=article&tags=golang,kubernetes&from=2026-01-01&limit=20&offset=0`
  - `GET /api/v1/tags?sort=count` — all tags with node counts
- [ ] API documentation (OpenAPI/Swagger spec)

**Phase 3 Deliverable:** Full REST API serving graph data, search results, and similarity queries.

---

### Phase 4: Frontend Visualization — "The Eyes"

**Timeline:** Weeks 10-14 (30-40 hours)  
**Goal:** Interactive knowledge graph web UI.

#### Week 10-11: React Scaffolding and Graph Canvas

- [ ] Scaffold React project: `npm create vite@latest web -- --template react-ts`
- [ ] Install dependencies:
  - `cytoscape`, `react-cytoscapejs` — graph rendering
  - `@tanstack/react-query` — server state management
  - `tailwindcss` — utility-first CSS
  - `lucide-react` — icon library
  - `react-router-dom` — client-side routing
- [ ] Create API client module (`web/src/lib/api.ts`):
  - Typed fetch wrappers for all backend endpoints
  - Error handling, loading states
- [ ] Global Graph View:
  - Fetch full graph from `/api/v1/graph`
  - Render with Cytoscape.js using CoSE (Compound Spring Embedder) layout
  - Color-code nodes by type (articles=blue, books=green, hobbies=orange, thoughts=purple, wildcards=red)
  - Size nodes by connection count (more connections = larger)
  - Edge thickness by weight

#### Week 12: Local View and Side Panel

- [ ] Click a node → transition to Local View:
  - Re-fetch subgraph centered on clicked node (depth=2)
  - Animate layout transition
  - Highlight clicked node, dim distant nodes
- [ ] Side Panel (right drawer):
  - Node title, type badge, creation date
  - AI-generated summary
  - Tag chips (clickable to filter)
  - Source URL (clickable link)
  - Image thumbnail (if image node)
  - List of connected nodes with relationship types
  - "Similar nodes" section (semantic similarity)
- [ ] Breadcrumb navigation: Global → Local → Node Detail

#### Week 13: Search and Filter

- [ ] Search bar (top of page):
  - Debounced input (300ms)
  - Autocomplete suggestions from `/api/v1/search`
  - Toggle between text/semantic/hybrid search modes
- [ ] Filter panel (left sidebar):
  - Checkbox filters by node type
  - Date range picker
  - Tag cloud (size by frequency, clickable to filter)
  - "Reset filters" button
- [ ] Visual feedback: filtered-out nodes fade to 10% opacity, matching nodes glow

#### Week 14: Polish and Deployment

- [ ] Responsive layout (desktop-first, functional on tablet)
- [ ] Dark mode support (Tailwind `dark:` classes, system preference detection)
- [ ] Loading skeletons for graph and side panel
- [ ] Error boundary with retry
- [ ] Add `web/Dockerfile` (Vite build → nginx serve)
- [ ] Add web service to Docker Compose
- [ ] End-to-end smoke test: ingest 10 articles → view graph → search → click through

**Phase 4 Deliverable:** Interactive web UI showing the full knowledge graph with search, filter, and node detail views.

---

### Phase 5: Multi-Modal and Journaling — "The Human Element"

**Timeline:** Weeks 15-18 (24-32 hours)  
**Goal:** Support images and unstructured manual entries alongside URL-sourced content.

#### Week 15-16: Image Upload and Storage

- [ ] MinIO bucket initialization on startup:
  - Create `mesh-images` bucket if not exists
  - Set lifecycle policy (no auto-deletion for personal use)
- [ ] `POST /api/v1/ingest/image` endpoint:
  - Accept `multipart/form-data`
  - Validate file type (JPEG, PNG, WebP, HEIC)
  - Generate UUID-based object key
  - Upload to MinIO via `minio-go`
  - Create node with type `image`, store object key in `image_key`
  - Optionally link to existing node via `parent_id` parameter
- [ ] Image serving: `GET /api/v1/images/:key`
  - Generate MinIO presigned URL (1-hour expiry)
  - Redirect or proxy
- [ ] If Ollama available: generate image description using vision model
  - Use as node content for embedding generation
  - Extract tags from description

#### Week 17-18: Journal Entry UI

- [ ] Journal entry page in frontend:
  - Rich text editor (lightweight: `@tiptap/react` or `react-quill`)
  - Title field (optional, auto-generated from first line if empty)
  - Tag suggestions as you type (autocomplete from existing tags)
  - "Save" creates node via `/api/v1/ingest/journal`
- [ ] Journal entries auto-processed:
  - NLP tag extraction (same pipeline as articles)
  - Embedding generation
  - Edge auto-creation to related nodes
- [ ] Gallery view for image nodes:
  - Grid layout of image thumbnails
  - Click to expand with linked nodes shown
- [ ] Timeline view: chronological list of all journal entries and image uploads

**Phase 5 Deliverable:** Upload a photo of a painting → it links to the "acrylic painting" cluster. Write a brain-dump → auto-tagged and connected.

---

### Phase 6: Anti-Echo Chamber Engine — "Discovery"

**Timeline:** Weeks 19-24 (36-48 hours)  
**Goal:** Automated knowledge gap detection and serendipitous injection.

#### Week 19-20: Cluster Density Analysis

- [ ] Implement cluster analysis service:
  - Group nodes by their dominant tag clusters
  - Calculate per-cluster metrics: node count, average inter-node similarity, edge density
  - Classify clusters: `over_saturated` (> mean + 1σ), `under_explored` (< mean - 1σ), `balanced`
  - Store results in `discovery_runs` table
- [ ] Cluster health API:
  - `GET /api/v1/clusters` returns cluster health report
  - Include: cluster name, node count, health status, trending direction
- [ ] Discovery dashboard in frontend:
  - Treemap or bubble chart of cluster sizes
  - Color-coded by health status (red = over-saturated, blue = under-explored, green = balanced)
  - Click cluster → filter graph to that cluster

#### Week 21-22: Adjacent Possible and Bridge Detection

- [ ] Bridge detection algorithm:
  - For each pair of distinct clusters (tags with different node sets):
    - Find the pair of nodes (one from each cluster) with highest vector cosine similarity but no existing edge
    - Score: `similarity * (1 / shared_tag_count)` — high similarity + low tag overlap = best bridge
  - Create `bridge` edges for top candidates
  - Surface as "Suggested Explorations"
- [ ] Adjacent Possible API:
  - `GET /api/v1/discovery/bridges` — returns top bridge candidates with explanation
  - Each bridge includes: the two nodes, their clusters, the similarity score, a one-sentence AI-generated explanation of why they connect
- [ ] Frontend: "Discovery" tab showing bridge suggestions as cards with "Explore" and "Dismiss" actions

#### Week 23-24: Wildcard Injector and Catch-Up Cron

- [ ] Wildcard Injector:
  - Weekly scheduled job (configurable frequency)
  - Sources: Wikipedia Random Article API, Hacker News top stories, arXiv random paper
  - Selection criteria: pick topic maximally distant from all existing cluster centroids in embedding space
  - Create `wildcard` node with scraped content, tags, embedding
  - Auto-create `wildcard` edges to nearest existing nodes
  - Notify via discovery dashboard
- [ ] Catch-Up Cron logic:
  - On container startup, query `discovery_runs` for last execution timestamp
  - If more than N days since last run, execute missed runs in sequence
  - Prevent duplicate runs using advisory locks (`pg_advisory_lock`)
- [ ] External API resilience:
  - Circuit breaker per external source
  - Graceful degradation: if Wikipedia is down, skip and retry next cycle
  - Configurable API keys for sources that require them
- [ ] Serendipity metrics:
  - Track user interactions with bridge/wildcard nodes (clicked, dismissed, explored further)
  - Calculate Kotkov metric: relevance × dissimilarity of accepted suggestions

**Phase 6 Deliverable:** System actively suggests cross-domain reading, injects new seed topics weekly, and tracks discovery effectiveness.

---

### Phase 7: Spaced Repetition and Semantic Depth — "The Slow Burn"

**Timeline:** Weeks 25-30 (36-48 hours)  
**Goal:** Fight knowledge decay and surface deep semantic connections.

#### Week 25-26: FSRS Algorithm Implementation

- [ ] Port FSRS algorithm to Go:
  - Reference: `open-spaced-repetition/fsrs-rs` (Rust, BSD-3)
  - Core functions:
    - `CalculateStability(difficulty, stability, rating) float64`
    - `CalculateDifficulty(oldDifficulty, rating) float64`
    - `CalculateRetrievability(stability, elapsedDays) float64`
    - `ScheduleReview(card ReviewSchedule, rating int) ReviewSchedule`
  - Rating scale: 1 (Again), 2 (Hard), 3 (Good), 4 (Easy)
  - Default parameters (from FSRS v5):
    - Initial stability: [0.4, 0.6, 2.4, 5.8] (per first rating)
    - Initial difficulty: 5.0
    - Decay: -0.5
    - Factor: 19/81
- [ ] Auto-enroll new nodes into review schedule:
  - On node creation, insert into `review_schedule` with `state: 'new'`
  - First due date: creation date + 1 day
- [ ] Unit tests for FSRS with known input/output pairs from reference implementation

#### Week 27-28: Review Queue and UI

- [ ] Review API:
  - `GET /api/v1/review/today` — returns the single most overdue node (or one disconnected node if no reviews due)
  - `POST /api/v1/review/:node_id` — submit rating, update FSRS parameters, calculate next due date
  - `GET /api/v1/review/stats` — review history: total reviewed, streak, average rating, upcoming schedule
- [ ] Review Card UI:
  - Full-screen card with node title, summary, tags
  - "Show Content" expandable section
  - Connected nodes preview
  - Rating buttons: Again (red), Hard (orange), Good (green), Easy (blue)
  - After rating: show next due date, animate card exit
  - Daily streak counter
- [ ] Review notification on dashboard: badge showing count of due reviews

#### Week 29-30: Semantic Cross-Pollination

- [ ] Nightly batch job: semantic edge builder
  - For all node pairs without existing edges:
    - Compute cosine similarity from stored embeddings
    - If similarity > 0.75 (configurable threshold): create `semantic` edge
  - Optimization: only process nodes created/updated since last run
  - Use pgvector's built-in nearest-neighbor for efficiency (not brute-force all pairs)
- [ ] Surface semantic connections in UI:
  - "Surprisingly similar" section on node detail panel
  - Distinct edge styling (dashed purple lines) on graph canvas
  - Tooltip showing similarity score and shared semantic concepts
- [ ] Serendipity metrics dashboard:
  - Kotkov metric trend over time
  - Total bridges explored vs. dismissed
  - Knowledge diversity score (entropy of tag distribution)
  - Graph connectivity evolution (average degree over time)
- [ ] Optional: K3s migration
  - Convert Docker Compose services to Kubernetes Deployments
  - Use K3s CronJob for scheduled discovery and review tasks
  - Traefik ingress for web UI
  - PersistentVolumeClaims with local-path-provisioner

**Phase 7 Deliverable:** Daily review card surfacing forgotten knowledge + deep semantic connections revealing hidden relationships between disparate knowledge domains.

---

## 7. Risk Assessment and Mitigation

### 7.1 Local Hardware Constraints (RAM/CPU)

| Component | Base RAM | Peak RAM | Notes |
|-----------|----------|----------|-------|
| PostgreSQL 16 | ~50 MB | ~200 MB | With shared_buffers=128MB |
| MinIO | ~30 MB | ~100 MB | Minimal for single-bucket use |
| Ollama (nomic-embed-text) | ~500 MB | ~700 MB | Embedding model only |
| Ollama (mistral:7b Q4) | ~4 GB | ~6 GB | Only during inference |
| Go API | ~10 MB | ~50 MB | Compiled binary |
| Go Worker (x4) | ~40 MB | ~200 MB | 4 goroutines |
| React Dev Server | ~100 MB | ~200 MB | Development only (nginx: ~5 MB) |
| **Total (with Mistral)** | **~4.7 GB** | **~7.5 GB** | Fits in 16 GB |
| **Total (without Mistral)** | **~730 MB** | **~1.5 GB** | Fits in 8 GB |

**Mitigation strategies:**
- Use **quantized models** (Q4_0): reduces Mistral from ~14 GB to ~4 GB RAM
- Make Ollama an **optional dependency** via Docker Compose profiles
- Fall back to `jdkato/prose` for basic NER when memory is tight
- Configure **Docker memory limits** per container to prevent OOM
- Ollama starts **on-demand** (not always-on): `docker compose --profile ai up ollama`

### 7.2 External API Rate Limiting and Scraping Failures

**Risk:** Web scraping targets may rate-limit, block, CAPTCHAs, or change HTML structure.

**Mitigation:**
- Circuit breaker (`sony/gobreaker`): opens after 5 consecutive failures, half-open probe after 60s
- `context.WithTimeout(ctx, 30*time.Second)` on all external HTTP calls
- Job retry with configurable max attempts (3) and exponential backoff
- Dead-letter queue: failed jobs moved to `status='dead'` after max retries for manual review
- Respect `robots.txt` via colly's built-in support
- Configurable inter-request delay per domain (default: 2 seconds)
- User-Agent rotation from a predefined list

### 7.3 Data Privacy and Sovereignty

**Risk:** Using cloud LLM APIs would leak personal knowledge data.

**Mitigation:**
- **All AI inference runs locally via Ollama** — zero data leaves the machine
- No cloud sync, no telemetry, no external analytics, no tracking
- PostgreSQL and MinIO data stored on **local Docker volumes only**
- All containers bind to `127.0.0.1` (localhost only, not exposed to network)
- Optional: encrypted backup script to local external drive
- No third-party JavaScript analytics in the frontend

### 7.4 Concurrency and Race Conditions

**Risk:** Multiple workers creating duplicate tags or corrupting edges when processing overlapping content.

**Mitigation:**
- PostgreSQL `ON CONFLICT (name) DO NOTHING` for tag UPSERT — idempotent, no duplicates
- `ON CONFLICT (source_id, target_id, rel_type) DO UPDATE SET weight = GREATEST(edges.weight, EXCLUDED.weight)` for edge UPSERT
- Optimistic Concurrency Control: `WHERE version = $expected_version` on all node updates
- `FOR UPDATE SKIP LOCKED` for job queue claims — zero deadlocks, zero contention
- PostgreSQL advisory locks (`pg_advisory_lock`) for singleton cron jobs

### 7.5 Scope Creep and Burnout

**Risk:** 7 phases over 30 weeks is ambitious for a side project with 6-8 hours per week.

**Mitigation:**
- Each phase delivers a **standalone, usable increment** — Phase 1 is useful by itself
- **Phase 1-4 (the MVP) is the primary commitment**; Phases 5-7 are optional enhancements
- "Catch-Up Cron" design means the system **tolerates arbitrary downtime** — no pressure to keep it running
- If a phase takes longer than estimated, **extend the timeline rather than cut corners**
- Monthly retrospective: re-evaluate priorities, drop low-value features
- No self-imposed deadlines — this is a personal growth project, not a startup

### 7.6 Embedding Model Drift and Quality

**Risk:** Switching embedding models later invalidates all existing vectors (different dimensions, different semantic spaces).

**Mitigation:**
- Standardize on `nomic-embed-text` (384-dim) from Phase 2 onward
- Store model name/version as metadata (consider adding `embedding_model TEXT` column to `nodes`)
- If migration is needed: batch re-embed script processes all nodes overnight (feasible at <10K nodes)
- pgvector HNSW index rebuild is fast at personal scale (~seconds for 10K vectors)

### 7.7 PostgreSQL as Sole Data Store (Single Point of Failure)

**Risk:** All critical data in one database. Corruption or volume loss means total data loss.

**Mitigation:**
- **Automated daily backups**: cron job running `pg_dump` to timestamped file
- Backup retention: keep last 7 daily + 4 weekly + 3 monthly dumps
- Store backups in a separate directory from the Docker volume
- Optional: rsync backups to a second physical drive
- MinIO data backed up separately (sync bucket to local directory)
- Test restore procedure once during Phase 1 setup

---

## 8. Project Structure

```
mesh/
├── cmd/
│   ├── api/                    # API server entrypoint
│   │   └── main.go
│   ├── worker/                 # Background worker entrypoint
│   │   └── main.go
│   └── discovery/              # Discovery engine entrypoint
│       └── main.go
├── internal/
│   ├── api/                    # HTTP handlers, middleware, routing
│   │   ├── handler/
│   │   │   ├── ingest.go       # Ingestion endpoints
│   │   │   ├── graph.go        # Graph query endpoints
│   │   │   ├── search.go       # Search endpoints
│   │   │   ├── review.go       # FSRS review endpoints
│   │   │   ├── discovery.go    # Discovery endpoints
│   │   │   └── nodes.go        # Node CRUD endpoints
│   │   ├── middleware/
│   │   │   ├── logging.go
│   │   │   ├── cors.go
│   │   │   └── recovery.go
│   │   └── router.go           # Route registration
│   ├── config/                 # Configuration loading (env vars)
│   │   └── config.go
│   ├── domain/                 # Core domain types
│   │   ├── node.go             # Node, NodeType
│   │   ├── edge.go             # Edge, RelationType
│   │   ├── tag.go              # Tag
│   │   ├── job.go              # Job, JobType, JobStatus
│   │   └── review.go           # ReviewSchedule, Rating
│   ├── graph/                  # Graph traversal logic
│   │   ├── traverse.go         # BFS/DFS via recursive CTEs
│   │   ├── cluster.go          # Cluster density analysis
│   │   └── bridge.go           # Bridge candidate detection
│   ├── ingest/                 # Content acquisition
│   │   ├── scraper.go          # URL scraping (colly)
│   │   ├── html.go             # HTML stripping (goquery)
│   │   └── circuit.go          # Circuit breaker wrapper
│   ├── nlp/                    # NLP and AI integration
│   │   ├── ollama.go           # Ollama client (tags, embeddings, summaries)
│   │   ├── fallback.go         # prose-based fallback NER
│   │   └── prompt.go           # Prompt templates
│   ├── discovery/              # Discovery engine
│   │   ├── engine.go           # Main discovery orchestration
│   │   ├── wildcard.go         # Wildcard topic injection
│   │   └── sources.go          # External source adapters (Wikipedia, HN, arXiv)
│   ├── fsrs/                   # FSRS algorithm implementation
│   │   ├── algorithm.go        # Core FSRS calculations
│   │   ├── scheduler.go        # Review scheduling logic
│   │   └── algorithm_test.go   # Tests against reference values
│   ├── storage/                # PostgreSQL data access layer
│   │   ├── postgres.go         # Connection pool setup
│   │   ├── nodes.go            # Node repository (sqlc-generated + custom)
│   │   ├── edges.go            # Edge repository
│   │   ├── tags.go             # Tag repository
│   │   ├── jobs.go             # Job queue repository
│   │   ├── reviews.go          # Review schedule repository
│   │   └── queries/            # sqlc SQL query files
│   │       ├── nodes.sql
│   │       ├── edges.sql
│   │       ├── tags.sql
│   │       ├── jobs.sql
│   │       └── reviews.sql
│   ├── objstore/               # MinIO client wrapper
│   │   └── minio.go
│   ├── queue/                  # Job queue orchestration
│   │   ├── worker.go           # Worker pool manager
│   │   └── processor.go        # Job type → handler dispatch
│   └── cache/                  # Caching interface
│       ├── cache.go            # Cache interface definition
│       └── ristretto.go        # ristretto implementation
├── migrations/                 # SQL migration files
│   ├── 001_initial_schema.up.sql
│   ├── 001_initial_schema.down.sql
│   ├── 002_add_review_schedule.up.sql
│   ├── 002_add_review_schedule.down.sql
│   └── ...
├── web/                        # React frontend application
│   ├── src/
│   │   ├── components/
│   │   │   ├── GraphCanvas.tsx     # Cytoscape.js wrapper
│   │   │   ├── SidePanel.tsx       # Node detail panel
│   │   │   ├── SearchBar.tsx       # Search with autocomplete
│   │   │   ├── FilterPanel.tsx     # Type/tag/date filters
│   │   │   ├── ReviewCard.tsx      # FSRS review card
│   │   │   ├── DiscoveryDash.tsx   # Discovery dashboard
│   │   │   ├── JournalEditor.tsx   # Rich text journal entry
│   │   │   └── ImageGallery.tsx    # Image node gallery
│   │   ├── hooks/
│   │   │   ├── useGraph.ts         # Graph data fetching
│   │   │   ├── useSearch.ts        # Search with debounce
│   │   │   └── useReview.ts        # Review queue state
│   │   ├── lib/
│   │   │   ├── api.ts              # Typed API client
│   │   │   ├── cytoscape.ts        # Cytoscape.js config and styles
│   │   │   └── types.ts            # TypeScript type definitions
│   │   ├── pages/
│   │   │   ├── GraphPage.tsx       # Main graph view
│   │   │   ├── SearchPage.tsx      # Search results
│   │   │   ├── ReviewPage.tsx      # Daily review
│   │   │   ├── DiscoveryPage.tsx   # Discovery dashboard
│   │   │   └── JournalPage.tsx     # Journal entry
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── tailwind.config.js
│   └── vite.config.ts
├── deploy/
│   ├── docker-compose.yml          # Full service topology
│   ├── docker-compose.dev.yml      # Development overrides
│   ├── Dockerfile.api              # Multi-stage Go build for API
│   ├── Dockerfile.worker           # Multi-stage Go build for Worker
│   ├── Dockerfile.web              # Vite build → nginx
│   ├── nginx.conf                  # Frontend nginx config
│   └── k3s/                        # Future K3s manifests
│       ├── namespace.yaml
│       ├── postgres.yaml
│       ├── api.yaml
│       ├── worker.yaml
│       └── ingress.yaml
├── scripts/
│   ├── backup.sh                   # PostgreSQL + MinIO backup script
│   ├── restore.sh                  # Restore from backup
│   ├── seed.sh                     # Seed database with sample data
│   └── pull-models.sh              # Download Ollama models
├── .env.example                    # Environment variable template
├── .gitignore
├── go.mod
├── go.sum
├── sqlc.yaml                       # sqlc configuration
├── Makefile                        # Build, test, run targets
└── README.md
```

---

## 9. Docker Compose Topology

### Full Docker Compose Configuration

```yaml
# deploy/docker-compose.yml
version: "3.9"

services:
  # ─── PostgreSQL with pgvector ───────────────────────────
  postgres:
    image: pgvector/pgvector:pg16
    container_name: mesh-postgres
    restart: unless-stopped
    volumes:
      - pgdata:/var/lib/postgresql/data
    environment:
      POSTGRES_DB: mesh
      POSTGRES_USER: mesh
      POSTGRES_PASSWORD: ${PG_PASSWORD:?Set PG_PASSWORD in .env}
    ports:
      - "127.0.0.1:5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U mesh"]
      interval: 5s
      timeout: 5s
      retries: 5

  # ─── MinIO Object Storage ──────────────────────────────
  minio:
    image: minio/minio:latest
    container_name: mesh-minio
    restart: unless-stopped
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data
    environment:
      MINIO_ROOT_USER: ${MINIO_USER:-meshadmin}
      MINIO_ROOT_PASSWORD: ${MINIO_PASSWORD:?Set MINIO_PASSWORD in .env}
    ports:
      - "127.0.0.1:9000:9000"
      - "127.0.0.1:9001:9001"
    healthcheck:
      test: ["CMD", "mc", "ready", "local"]
      interval: 10s
      timeout: 5s
      retries: 3

  # ─── Ollama Local LLM (optional) ──────────────────────
  ollama:
    image: ollama/ollama:latest
    container_name: mesh-ollama
    restart: unless-stopped
    volumes:
      - ollama_models:/root/.ollama
    ports:
      - "127.0.0.1:11434:11434"
    profiles: ["ai"]
    deploy:
      resources:
        limits:
          memory: 8G

  # ─── Go API Server ────────────────────────────────────
  api:
    build:
      context: ..
      dockerfile: deploy/Dockerfile.api
    container_name: mesh-api
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
      minio:
        condition: service_healthy
    environment:
      DATABASE_URL: postgres://mesh:${PG_PASSWORD}@postgres:5432/mesh?sslmode=disable
      MINIO_ENDPOINT: minio:9000
      MINIO_ACCESS_KEY: ${MINIO_USER:-meshadmin}
      MINIO_SECRET_KEY: ${MINIO_PASSWORD}
      MINIO_BUCKET: mesh-images
      OLLAMA_HOST: http://ollama:11434
      LOG_LEVEL: info
    ports:
      - "127.0.0.1:8080:8080"

  # ─── Go Background Worker ─────────────────────────────
  worker:
    build:
      context: ..
      dockerfile: deploy/Dockerfile.worker
    container_name: mesh-worker
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      DATABASE_URL: postgres://mesh:${PG_PASSWORD}@postgres:5432/mesh?sslmode=disable
      MINIO_ENDPOINT: minio:9000
      MINIO_ACCESS_KEY: ${MINIO_USER:-meshadmin}
      MINIO_SECRET_KEY: ${MINIO_PASSWORD}
      OLLAMA_HOST: http://ollama:11434
      WORKER_COUNT: 4
      LOG_LEVEL: info

  # ─── React Frontend ───────────────────────────────────
  web:
    build:
      context: ../web
      dockerfile: ../deploy/Dockerfile.web
    container_name: mesh-web
    restart: unless-stopped
    ports:
      - "127.0.0.1:3000:80"
    depends_on:
      - api

volumes:
  pgdata:
    driver: local
  minio_data:
    driver: local
  ollama_models:
    driver: local
```

### Development Override

```yaml
# deploy/docker-compose.dev.yml
# Usage: docker compose -f docker-compose.yml -f docker-compose.dev.yml up
services:
  api:
    build:
      target: builder
    volumes:
      - ../:/app
    command: ["go", "run", "./cmd/api"]
    environment:
      LOG_LEVEL: debug

  worker:
    build:
      target: builder
    volumes:
      - ../:/app
    command: ["go", "run", "./cmd/worker"]
    environment:
      LOG_LEVEL: debug

  web:
    image: node:20-alpine
    working_dir: /app
    volumes:
      - ../web:/app
    command: ["npm", "run", "dev", "--", "--host"]
    ports:
      - "127.0.0.1:5173:5173"
```

### Multi-Stage Dockerfile (API Example)

```dockerfile
# deploy/Dockerfile.api
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /mesh-api ./cmd/api

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /mesh-api /usr/local/bin/mesh-api
EXPOSE 8080
ENTRYPOINT ["mesh-api"]
```

### Environment Template

```bash
# .env.example
PG_PASSWORD=change-me-to-something-secure
MINIO_USER=meshadmin
MINIO_PASSWORD=change-me-to-something-secure
```

### Makefile Targets

```makefile
.PHONY: build run test migrate-up migrate-down docker-up docker-down lint

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker
	go build -o bin/discovery ./cmd/discovery

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

test:
	go test ./... -v -race -count=1

test-integration:
	go test ./... -v -race -tags=integration

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

docker-up:
	cd deploy && docker compose up -d

docker-up-ai:
	cd deploy && docker compose --profile ai up -d

docker-down:
	cd deploy && docker compose --profile ai down

docker-logs:
	cd deploy && docker compose logs -f

lint:
	golangci-lint run ./...

sqlc:
	sqlc generate

seed:
	./scripts/seed.sh

backup:
	./scripts/backup.sh

pull-models:
	docker exec mesh-ollama ollama pull nomic-embed-text
	docker exec mesh-ollama ollama pull mistral:7b-instruct-q4_0
```

---

## 10. Appendix: Key Algorithms

### 10.1 FSRS Core Calculations (Pseudocode)

```
// FSRS v5 parameters (defaults, optimizable from review logs)
DECAY = -0.5
FACTOR = 19.0 / 81.0
INITIAL_STABILITY = [0.4, 0.6, 2.4, 5.8]  // indexed by first rating (1-4)

function calculateRetrievability(stability, elapsedDays):
    return (1 + FACTOR * elapsedDays / stability) ^ DECAY

function calculateNextStability(difficulty, stability, retrievability, rating):
    if rating == 1:  // Again (forgot)
        return stability * exp(-0.5 * difficulty) *
               ((retrievability + 1) ^ 0.2 - 1) *
               max(1, exp(1 - stability))
    else:  // Hard, Good, Easy
        hardPenalty = if rating == 2 then 0.85 else 1.0
        easyBonus = if rating == 4 then 1.3 else 1.0
        return stability *
               (1 + exp(0.5 - difficulty) *
                (stability ^ -0.2) *
                (retrievability ^ -0.5 - 1) *
                hardPenalty * easyBonus)

function calculateNextDifficulty(difficulty, rating):
    newDiff = difficulty - 0.7 * (rating - 3)
    return clamp(newDiff, 1.0, 10.0)

function scheduleReview(card, rating):
    elapsed = daysSince(card.lastReview)
    retrievability = calculateRetrievability(card.stability, elapsed)
    card.stability = calculateNextStability(
        card.difficulty, card.stability, retrievability, rating)
    card.difficulty = calculateNextDifficulty(card.difficulty, rating)
    card.dueDate = now() + card.stability days
    card.reps += 1
    if rating == 1: card.lapses += 1
    return card
```

### 10.2 Kotkov Serendipity Metric

```
Serendipity(recommendations, userHistory) =
    |{item ∈ recommendations : isRelevant(item, user) AND isDissimilar(item, userHistory)}|
    / |recommendations|

where:
    isRelevant(item, user) = cosineSimilarity(item.embedding, user.centroid) > relevanceThreshold
    isDissimilar(item, history) = max(cosineSimilarity(item.embedding, h.embedding) for h in history) < dissimilarityThreshold

Higher serendipity score = more items that are relevant yet surprising
Target: 0.3 - 0.5 (30-50% of suggestions should be serendipitous)
```

### 10.3 Cluster Density Scoring

```
For each tag cluster C:
    density(C) = |edges within C| / (|nodes in C| * (|nodes in C| - 1) / 2)
    centroid(C) = mean(embedding for node in C)
    isolation(C) = min(distance(centroid(C), centroid(D)) for D != C)

Health classification:
    over_saturated: |nodes| > mean + 1σ AND density > 0.7
    under_explored: |nodes| < mean - 1σ OR isolation > 0.8
    balanced: otherwise
```

---

## 11. References and Research Sources

### Database Selection
- "Scaling Capacities: Why we swapped Dgraph for PostgreSQL" — Capacities blog, 2026 (70% cost reduction post-migration)
- "Building a personal knowledge graph with just PostgreSQL (no Neo4j needed)" — micelclaw.com (recursive CTEs at personal scale)
- "SQLite as a Graph Database: Recursive CTEs, Semantic Search" — dev.to (further validation of relational DBs for small graphs)

### Vector Search
- pgvector-go: github.com/pgvector/pgvector-go (native Go support for pgvector)
- goformersearch: Pure Go vector similarity search, 10K-50K documents, HNSW indexing
- "Building a Semantic Search Engine with OpenAI, Go, and PostgreSQL (pgvector)" — dev.to

### Spaced Repetition
- open-spaced-repetition/fsrs-rs: Reference FSRS v5 implementation (Rust, BSD-3)
- open-spaced-repetition/fsrs-optimizer: Python-based parameter optimizer from review logs

### Graph Visualization
- Cytoscape.js: ~500K weekly npm downloads, built-in graph algorithms, multiple layouts
- Sigma.js: WebGL rendering for 100K+ node graphs (future scaling option)

### Go Architecture
- "Go Microservices Architecture: Patterns and Best Practices 2026" — reintech.io
- sony/gobreaker: Circuit breaker for Go
- dgraph-io/ristretto: High-performance concurrent cache

### NLP / AI
- ollamaclient (Go): Ollama wrapper with embedding support
- gollama (Go): Structured outputs, function calling, vision
- jdkato/prose: Pure-Go tokenization, POS tagging, NER

### PKM Market Analysis
- Obsidian vs Logseq comparisons (softpicker.com, trybuildpilot.com, aiproductivity.ai)
- "Knowledge Graph Tools Compared (2026)" — atlasworkspace.ai
- "Algorithmic serendipity — can AI bring back discovery?" — ManageEngine

### Orchestration
- "I Migrated from Docker Compose to K3s on a Single Server" — kgabeci, Medium, 2026
- "Docker Compose vs Kubernetes: When to Use Each in 2026" — dev.to
- K3s documentation: k3s.io

---

*This document is a living blueprint. Update it as architectural decisions evolve during implementation.*
