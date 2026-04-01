# Mesh Codebase Review Checklist

**Purpose:** This is a structured audit framework for reviewing the Mesh codebase at any point during development. When asked to "review the codebase," walk through each applicable section, assess each criterion, and produce a report with status (`PASS` / `PARTIAL` / `MISSING` / `N/A`), notes, and recommendations.

**How to use:** Only evaluate sections relevant to the current phase. Mark future-phase items as `N/A - Phase X`.

---

## 1. Project Structure

| # | Criterion | Phase | Status | Notes |
|---|-----------|-------|--------|-------|
| 1.1 | Directory layout matches blueprint (cmd/, internal/, migrations/, web/, deploy/, scripts/) | 1 | PASS | Week 1 scaffolding |
| 1.2 | No orphan files in root (only go.mod, go.sum, Makefile, sqlc.yaml, .env.example, .gitignore, README.md) | 1 | PASS | |
| 1.3 | Go module initialized as `github.com/neha037/mesh` | 1 | PASS | |
| 1.4 | `.gitignore` covers binaries, .env, vendor/, node_modules/, IDE files | 1 | PASS | |
| 1.5 | cmd/ has separate entrypoints: api, worker, discovery | 1-6 | | |
| 1.6 | internal/ packages follow single-responsibility (no circular imports) | 1+ | | |
| 1.7 | web/ follows standard React+Vite structure (src/components, hooks, lib, pages) | 4 | | |

## 2. Go Backend Quality

| # | Criterion | Phase | Status | Notes |
|---|-----------|-------|--------|-------|
| 2.1 | All exported functions/types have purpose-clear names (no stuttering: `node.NodeType` → `node.Type`) | 1+ | | |
| 2.2 | Errors are wrapped with context (`fmt.Errorf("doing X: %w", err)`) | 1+ | | |
| 2.3 | No `panic()` in library code; only in main() for unrecoverable setup failures | 1+ | | |
| 2.4 | `context.Context` is the first parameter where applicable | 1+ | | |
| 2.5 | All external HTTP calls use `context.WithTimeout` | 1 | | |
| 2.6 | Graceful shutdown via signal handling (SIGINT/SIGTERM) | 1 | | |
| 2.7 | No global mutable state; dependencies injected via constructors | 1+ | | |
| 2.8 | Interfaces defined where consumed, not where implemented | 1+ | | |
| 2.9 | Goroutines have clear ownership and shutdown paths | 2 | | |
| 2.10 | `go vet` and `golangci-lint` pass with zero warnings | 1+ | | |

## 3. API Completeness

| # | Criterion | Phase | Status | Notes |
|---|-----------|-------|--------|-------|
| 3.1 | `POST /api/v1/ingest/url` — validates URL, creates node, enqueues job, returns 202 | 1 | | |
| 3.2 | `POST /api/v1/ingest/text` — validates fields, creates node, returns 201 | 1 | | |
| 3.3 | `GET /api/v1/graph` — full graph with pagination | 3 | | |
| 3.4 | `GET /api/v1/graph?center=<uuid>&depth=<int>` — BFS subgraph | 3 | | |
| 3.5 | `GET /api/v1/nodes` — list with pagination, type/tag/date filters | 3 | | |
| 3.6 | `GET /api/v1/nodes/:id` — single node with edges and tags | 3 | | |
| 3.7 | `PUT /api/v1/nodes/:id` — update with OCC version check | 3 | | |
| 3.8 | `DELETE /api/v1/nodes/:id` — cascade delete | 3 | | |
| 3.9 | `GET /api/v1/search?q=&mode=text\|semantic\|hybrid` — all three modes | 3 | | |
| 3.10 | `GET /api/v1/nodes/:id/similar` — vector similarity | 3 | | |
| 3.11 | `GET /api/v1/tags` — all tags with counts | 3 | | |
| 3.12 | `POST /api/v1/ingest/image` — multipart upload to MinIO | 5 | | |
| 3.13 | `POST /api/v1/ingest/journal` — journal entry | 5 | | |
| 3.14 | `GET /api/v1/review/today` — FSRS due review | 7 | | |
| 3.15 | `POST /api/v1/review/:node_id` — submit rating | 7 | | |
| 3.16 | `GET /api/v1/clusters` — cluster health report | 6 | | |
| 3.17 | `GET /api/v1/discovery/bridges` — bridge candidates | 6 | | |
| 3.18 | `POST /api/v1/discovery/trigger` — manual discovery run | 6 | | |
| 3.19 | Request logging middleware active | 1 | | |
| 3.20 | CORS middleware configured | 1 | | |
| 3.21 | Consistent JSON error response format | 1+ | | |
| 3.22 | Input validation on all POST/PUT endpoints | 1+ | | |

## 4. Database & Migrations

| # | Criterion | Phase | Status | Notes |
|---|-----------|-------|--------|-------|
| 4.1 | `golang-migrate` configured and working | 1 | PASS | Makefile targets ready |
| 4.2 | Every `.up.sql` has a matching `.down.sql` | 1+ | PASS | |
| 4.3 | Schema matches blueprint: nodes, tags, node_tags, edges, jobs tables | 1 | PASS | All 7 tables in initial migration |
| 4.4 | review_schedule table present | 7 | PASS | Included in initial migration |
| 4.5 | discovery_runs table present | 6 | PASS | Included in initial migration |
| 4.6 | pgvector extension enabled, `embedding vector(384)` column on nodes | 2 | PASS | In initial migration |
| 4.7 | pg_trgm extension enabled, trigram indexes on title/content | 3 | PASS | In initial migration |
| 4.8 | HNSW index on embeddings with appropriate parameters | 2 | PASS | m=16, ef_construction=64 |
| 4.9 | Job queue index: `idx_jobs_pending` with partial index on status='pending' | 1 | PASS | |
| 4.10 | All foreign keys have ON DELETE CASCADE | 1 | PASS | |
| 4.11 | UNIQUE constraints: tags.name, edges(source_id, target_id, rel_type) | 1 | PASS | |
| 4.12 | CHECK constraints on type enums (nodes.type, edges.rel_type, jobs.type, jobs.status) | 1 | PASS | |
| 4.13 | sqlc configured and generating type-safe Go code | 1 | PARTIAL | sqlc.yaml ready, queries not yet written (Week 2) |
| 4.14 | Migrations can be applied to a fresh database without errors | 1+ | | Needs Docker verification |
| 4.15 | Migrations can be rolled back cleanly | 1+ | | Needs Docker verification |

## 5. Worker System

| # | Criterion | Phase | Status | Notes |
|---|-----------|-------|--------|-------|
| 5.1 | Worker pool with configurable goroutine count | 2 | | |
| 5.2 | Job claim via `SELECT ... FOR UPDATE SKIP LOCKED` | 1 | | |
| 5.3 | Exponential backoff when no jobs available (1s → 30s cap) | 2 | | |
| 5.4 | Graceful shutdown: workers finish current job before exiting | 2 | | |
| 5.5 | Job retry with max_attempts (default 3) | 1 | | |
| 5.6 | Dead-letter: failed jobs set to `status='dead'` after max retries | 1 | | |
| 5.7 | HTML stripping pipeline: fetch → parse → strip scripts/styles → clean text | 2 | | |
| 5.8 | Web scraper respects robots.txt, has User-Agent rotation, inter-request delay | 1 | | |
| 5.9 | Circuit breaker on external HTTP calls (open after 5 failures, half-open after 60s) | 1 | | |
| 5.10 | All external calls wrapped with `context.WithTimeout(30s)` | 1 | | |
| 5.11 | Tag UPSERT uses `ON CONFLICT (name) DO NOTHING` | 2 | | |
| 5.12 | Edge UPSERT uses `ON CONFLICT ... DO UPDATE SET weight = GREATEST(...)` | 2 | | |
| 5.13 | Auto-edge generation: creates edges for nodes sharing 2+ tags | 2 | | |
| 5.14 | Optimistic concurrency control: `WHERE version = $expected_version` | 2 | | |

## 6. NLP/AI Integration

| # | Criterion | Phase | Status | Notes |
|---|-----------|-------|--------|-------|
| 6.1 | Ollama HTTP client implemented (generate + embeddings endpoints) | 2 | | |
| 6.2 | Tag extraction prompt produces JSON array of 3-8 domain-specific concepts | 2 | | |
| 6.3 | Embedding generation uses `nomic-embed-text` (384-dim) | 2 | | |
| 6.4 | Retry logic with circuit breaker on Ollama calls | 2 | | |
| 6.5 | Fallback to `jdkato/prose` NER when Ollama unavailable | 2 | | |
| 6.6 | Embeddings skipped and queued for batch when Ollama down | 2 | | |
| 6.7 | Ollama configured as optional Docker profile (`profiles: ["ai"]`) | 2 | | |

## 7. Frontend

| # | Criterion | Phase | Status | Notes |
|---|-----------|-------|--------|-------|
| 7.1 | React + TypeScript + Vite scaffolded | 4 | | |
| 7.2 | Cytoscape.js rendering with CoSE layout | 4 | | |
| 7.3 | Nodes color-coded by type | 4 | | |
| 7.4 | Node size proportional to connection count | 4 | | |
| 7.5 | Click node → local view (depth=2 subgraph) | 4 | | |
| 7.6 | Side panel: title, summary, tags, source URL, connected nodes | 4 | | |
| 7.7 | Search bar with debounce (300ms) and mode toggle | 4 | | |
| 7.8 | Filter panel: node type, date range, tag cloud | 4 | | |
| 7.9 | TanStack Query for server state management | 4 | | |
| 7.10 | Typed API client (`web/src/lib/api.ts`) | 4 | | |
| 7.11 | Dark mode support | 4 | | |
| 7.12 | Loading skeletons and error boundaries | 4 | | |
| 7.13 | Journal editor (tiptap or react-quill) | 5 | | |
| 7.14 | Image gallery view | 5 | | |
| 7.15 | Review card UI with rating buttons | 7 | | |
| 7.16 | Discovery dashboard with cluster visualization | 6 | | |
| 7.17 | TypeScript strict mode enabled | 4 | | |

## 8. Testing

| # | Criterion | Phase | Status | Notes |
|---|-----------|-------|--------|-------|
| 8.1 | Unit tests for HTTP handlers (table-driven) | 1 | | |
| 8.2 | Integration tests using `testcontainers-go` with real PostgreSQL | 1 | | |
| 8.3 | Tests run with `-race` flag | 1+ | | |
| 8.4 | FSRS algorithm tests against reference implementation values | 7 | | |
| 8.5 | Worker pipeline integration test (ingest → process → edges) | 2 | | |
| 8.6 | Search tests covering text, semantic, and hybrid modes | 3 | | |
| 8.7 | Frontend: component tests exist for key components | 4 | | |
| 8.8 | No test files import from other test files | 1+ | | |
| 8.9 | CI-friendly: tests don't depend on external services beyond testcontainers | 1+ | | |

## 9. Docker & Deployment

| # | Criterion | Phase | Status | Notes |
|---|-----------|-------|--------|-------|
| 9.1 | `docker-compose.yml` has all services: postgres, minio, api, worker, web | 1-4 | PARTIAL | Web commented out (Phase 4) |
| 9.2 | Ollama in separate profile (`profiles: ["ai"]`) | 2 | PASS | |
| 9.3 | Health checks on postgres and minio | 1 | PASS | |
| 9.4 | Service dependencies use `condition: service_healthy` | 1 | PASS | |
| 9.5 | All ports bound to `127.0.0.1` (not `0.0.0.0`) | 1 | PASS | |
| 9.6 | Named volumes for persistent data (pgdata, minio_data, ollama_models) | 1 | PASS | |
| 9.7 | Multi-stage Dockerfile for Go services (builder → alpine/scratch) | 1 | PASS | builder → alpine |
| 9.8 | `docker-compose.dev.yml` with hot-reload setup | 1 | | Not yet created |
| 9.9 | `.env.example` with all required environment variables | 1 | PASS | |
| 9.10 | Docker memory limits on Ollama container | 2 | PASS | 8G limit |

## 10. Security & Privacy

| # | Criterion | Phase | Status | Notes |
|---|-----------|-------|--------|-------|
| 10.1 | All container ports bound to localhost only | 1 | PASS | |
| 10.2 | No hardcoded credentials in source code | 1 | PASS | |
| 10.3 | Credentials loaded from environment variables | 1 | PASS | |
| 10.4 | `.env` in `.gitignore` | 1 | PASS | |
| 10.5 | No external analytics or telemetry in frontend | 4 | | |
| 10.6 | Input validation on all user-facing endpoints (URL format, field lengths, types) | 1+ | | |
| 10.7 | SQL injection prevented (parameterized queries via sqlc/pgx) | 1 | | |
| 10.8 | XSS prevention: HTML content sanitized before storage/display | 2 | | |
| 10.9 | No data sent to external services (except user-initiated scraping) | 1+ | | |
| 10.10 | MinIO presigned URLs use short expiry (1 hour) | 5 | | |

## 11. Documentation

| # | Criterion | Phase | Status | Notes |
|---|-----------|-------|--------|-------|
| 11.1 | README.md exists and is accurate to current state | 1 | PASS | Phase status updated |
| 11.2 | Developer's Guide exists and covers local setup | 1 | PASS | |
| 11.3 | Quick start instructions work end-to-end | 1 | | Needs Docker verification |
| 11.4 | Environment variables documented | 1 | PASS | |
| 11.5 | Makefile targets documented | 1 | PASS | |
| 11.6 | API documentation (OpenAPI/Swagger or equivalent) | 3 | | |
| 11.7 | Current phase status reflected in README | 1+ | | |

## 12. Performance

| # | Criterion | Phase | Status | Notes |
|---|-----------|-------|--------|-------|
| 12.1 | PostgreSQL connection pooling configured (pgx pool) | 1 | | |
| 12.2 | Ristretto cache implemented behind `cache.Store` interface | 3+ | | |
| 12.3 | Graph traversal queries have depth limits (max 5) | 3 | | |
| 12.4 | Pagination on all list endpoints | 3 | | |
| 12.5 | Debounced search on frontend (300ms) | 4 | | |
| 12.6 | Cytoscape.js handles 2-5K nodes without jank | 4 | | |
| 12.7 | Worker backoff prevents CPU spin on empty queue | 2 | | |

---

## Phase Completion Tracker

### Phase 1: Foundation and Ingestion — "The Senses" (Weeks 1-3)

- [x] Go module initialized
- [x] Project directory structure created
- [x] Docker Compose with PostgreSQL 16 + pgvector
- [x] Initial SQL migration with core tables
- [x] golang-migrate configured
- [x] Makefile with build/run/test/migrate/docker targets
- [x] Multi-stage Dockerfile for API
- [x] Git repo initialized with .gitignore
- [ ] HTTP server with chi router
- [ ] `POST /api/v1/ingest/url` endpoint
- [ ] `POST /api/v1/ingest/text` endpoint
- [ ] PostgreSQL repository layer (pgx + sqlc)
- [ ] Request logging middleware
- [ ] CORS middleware
- [ ] Web scraper with colly (timeout, robots.txt, User-Agent rotation)
- [ ] Circuit breaker (sony/gobreaker)
- [ ] Job queue claim logic (FOR UPDATE SKIP LOCKED)
- [ ] Unit tests for handlers (table-driven)
- [ ] Integration tests with testcontainers-go
- [ ] End-to-end verification: curl URL → job → scrape → store

### Phase 2: Processing and Intelligence — "The Brain" (Weeks 4-6)

- [ ] Worker pool (configurable goroutines, graceful shutdown, backoff)
- [ ] HTML stripping pipeline (goquery)
- [ ] Dockerfile.worker + docker-compose worker service
- [ ] Ollama container in docker-compose (profile: ai)
- [ ] Ollama Go client (generate + embeddings)
- [ ] Tag extraction prompt engineering
- [ ] Embedding generation (nomic-embed-text, 384-dim)
- [ ] Tag UPSERT and node_tags association
- [ ] Auto-edge generation (2+ shared tags)
- [ ] Optimistic concurrency control
- [ ] Fallback NLP (jdkato/prose)
- [ ] Integration tests for full pipeline

### Phase 3: Graph Traversal and Query API — "The Memory" (Weeks 7-9)

- [ ] Recursive CTE graph traversal (BFS, depth-limited)
- [ ] Full graph export with pagination
- [ ] Full-text search (pg_trgm)
- [ ] Semantic search (pgvector cosine distance)
- [ ] Hybrid search (Reciprocal Rank Fusion)
- [ ] Node similarity endpoint
- [ ] Node CRUD (GET, PUT, DELETE)
- [ ] Edge management (POST, DELETE)
- [ ] Filtering and pagination on node listing
- [ ] Tags endpoint with counts
- [ ] API documentation (OpenAPI/Swagger)

### Phase 4: Frontend Visualization — "The Eyes" (Weeks 10-14)

- [ ] React + TypeScript + Vite scaffolded
- [ ] Cytoscape.js graph rendering (CoSE layout)
- [ ] Typed API client
- [ ] Global graph view (color-coded, sized by connections)
- [ ] Local view (click node → depth=2 subgraph)
- [ ] Side panel (title, summary, tags, source, connections, similar nodes)
- [ ] Search bar with debounce and mode toggle
- [ ] Filter panel (type, date, tags)
- [ ] Dark mode
- [ ] Loading skeletons and error boundaries
- [ ] Dockerfile.web (Vite build → nginx)
- [ ] Web service in docker-compose

### Phase 5: Multi-Modal and Journaling — "The Human Element" (Weeks 15-18)

- [ ] MinIO bucket initialization
- [ ] Image upload endpoint (multipart, validation, MinIO storage)
- [ ] Image serving with presigned URLs
- [ ] Vision model description (if Ollama available)
- [ ] Journal entry page (rich text editor)
- [ ] Journal auto-processing (tags, embedding, edges)
- [ ] Image gallery view
- [ ] Timeline view

### Phase 6: Anti-Echo Chamber Engine — "Discovery" (Weeks 19-24)

- [ ] Cluster density analysis service
- [ ] Cluster health API
- [ ] Discovery dashboard (treemap/bubble chart)
- [ ] Bridge detection algorithm
- [ ] Adjacent Possible API
- [ ] Discovery tab in frontend
- [ ] Wildcard injector (Wikipedia, HN, arXiv)
- [ ] Catch-up cron logic (advisory locks, missed run detection)
- [ ] External API resilience (circuit breakers per source)
- [ ] Serendipity metrics tracking

### Phase 7: Spaced Repetition and Semantic Depth — "The Slow Burn" (Weeks 25-30)

- [ ] FSRS algorithm ported to Go
- [ ] Auto-enroll nodes into review schedule
- [ ] FSRS unit tests against reference values
- [ ] Review API (today, submit rating, stats)
- [ ] Review card UI
- [ ] Nightly semantic edge builder batch job
- [ ] "Surprisingly similar" section in node detail
- [ ] Distinct edge styling for semantic edges
- [ ] Serendipity metrics dashboard
- [ ] Optional: K3s migration manifests
