# Mesh — Project Progress

**Last Updated:** April 2, 2026

This is a living document tracking what has been completed, what's in progress, and what's next. It will be updated as the project evolves.

---

## Project Timeline

| Date | Milestone |
|------|-----------|
| March 31, 2026 | Project conception — architectural blueprint written |
| April 1, 2026 | Documentation framework created (README, Developer's Guide, Review Checklist, this document) |
| April 1, 2026 | Phase 1 Week 1 — Project scaffolding complete (Git, Go module, Docker Compose, migrations, Dockerfiles, Makefile) |
| April 2, 2026 | Phase 1 Week 2 — HTTP server, chi router, CORS, ingest/raw endpoint, recent nodes endpoint, browser extension (auto-save, view all, delete), systemd service + tray icon (Wayland AppIndicator), URL dedup (upsert), keyset pagination, connection pool tuning, GitHub Pages documentation site |

---

## Overall Status

**Current Phase:** Phase 1 — Foundation & Ingestion ("The Senses")

**Week 1 scaffolding is complete.** Git repo initialized, Go module created, Docker Compose configured, initial SQL migration written, Dockerfiles and Makefile ready.

---

## Completed Work

### Planning & Architecture

- [x] **Architectural Blueprint** (`docs/PROJECT_MESH_BLUEPRINT.md`) — 1,784 lines covering:
  - Executive summary and mission statement
  - Competitive landscape analysis (Obsidian, Logseq, Heptabase, Tana, Mem.ai)
  - Core differentiator defined: algorithmic serendipity via Adjacent Possible
  - Feature stratification: MVP (Phases 1-4) vs. Delighters (Phases 5-7)
  - Full technical stack selection with justifications
  - System component diagram and service interaction flows
  - Complete API endpoint specification (25+ endpoints across 5 categories)
  - Full SQL schema with 7 tables, indexes, and constraints
  - Key query patterns (BFS traversal, semantic search, job queue, cluster analysis, bridge detection)
  - 7-phase implementation roadmap spanning 30 weeks
  - Risk assessment with 7 risk categories and mitigations
  - Complete project directory structure
  - Docker Compose topology with 6 services
  - Multi-stage Dockerfile examples
  - Makefile target definitions
  - FSRS algorithm pseudocode
  - Kotkov serendipity metric definition
  - Cluster density scoring algorithm
  - References and research sources

### Key Architectural Decisions (Locked In)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Backend language | **Go** | Compiled binary, goroutine concurrency, developer expertise |
| Database | **PostgreSQL 16 + pgvector** | Recursive CTEs for graph traversal, vector search, single process — rejected Neo4j (heavy, overkill) and Dgraph (unpredictable CPU, extra complexity) |
| Caching | **ristretto (in-process)** | No Redis needed for single-user; ~10ns reads vs ~100us for Redis over loopback |
| Object storage | **MinIO** | S3-compatible, single container, Go-native client |
| AI/NLP | **Ollama (local)** | Zero data leaves machine; optional via Docker profiles |
| Embedding model | **nomic-embed-text (384-dim)** | Standardized from Phase 2 onward |
| LLM | **mistral:7b-instruct-q4_0** | Fits in 16 GB RAM with all other services |
| Frontend | **React + TypeScript + Cytoscape.js** | Larger ecosystem than Vue; Cytoscape has built-in graph algorithms |
| Graph viz library | **Cytoscape.js** | Built-in BFS/DFS/PageRank, multiple layouts, canvas rendering for 2-5K nodes — rejected D3.js (lower-level) and React Flow (no graph algorithms) |
| Orchestration | **Docker Compose** (MVP) | Simple lifecycle; K3s optional for Phase 7+ |
| Job queue | **PostgreSQL-based** (FOR UPDATE SKIP LOCKED) | No external message broker needed at personal scale |
| Spaced repetition | **FSRS v5** | State-of-the-art algorithm, open source reference implementation available |

### Documentation Framework

- [x] **README.md** — Project overview, architecture diagram, tech stack, quick start, phase status table
- [x] **Developer's Guide** (`docs/DEVELOPERS_GUIDE.md`) — Prerequisites, repo structure, local dev workflow, database/sqlc workflow, how to add endpoints and workers, testing strategy, Makefile reference, env vars, Ollama setup, conventions, troubleshooting
- [x] **Review Checklist** (`docs/REVIEW_CHECKLIST.md`) — 80+ audit criteria across 12 categories, plus per-phase completion tracker for all 7 phases
- [x] **Progress Tracker** (`docs/PROJECT_PROGRESS.md`) — This document

---

## Current Repository State

### Files That Exist

```
mesh/
├── .gitignore                          # VCS exclusions
├── .env.example                        # Environment variable template
├── go.mod                              # Go module definition
├── Makefile                            # Build, test, migrate, docker targets
├── sqlc.yaml                           # sqlc code generation config
├── README.md                           # Project overview and quick start
├── CLAUDE.md                           # Claude integration instructions
├── cmd/
│   ├── api/main.go                     # Stub API server entrypoint
│   └── worker/main.go                  # Stub worker entrypoint
├── internal/
│   ├── config/config.go                # Environment-based config loading
│   ├── api/
│   │   ├── router.go                   # chi router with middleware
│   │   └── handler/
│   │       ├── handler.go              # Handler struct with dependencies
│   │       └── ingest.go               # Ingest, list, delete handlers
│   ├── storage/
│   │   ├── db.go                       # Database interface (sqlc generated)
│   │   ├── models.go                   # Data models (sqlc generated)
│   │   ├── nodes.sql.go                # Node queries (sqlc generated)
│   │   ├── postgres.go                 # PostgreSQL connection pool setup
│   │   └── queries/nodes.sql           # SQL query definitions
│   └── domain/                         # Core types (empty, Week 2)
├── migrations/
│   ├── embed.go                        # Embed migrations for binary
│   ├── 001_initial_schema.up.sql       # Full schema: 7 tables + indexes
│   └── 001_initial_schema.down.sql     # Reverse migration
├── deploy/
│   ├── docker-compose.yml              # PostgreSQL, MinIO, Ollama, API, Worker
│   ├── Dockerfile.api                  # Multi-stage Go build for API
│   └── Dockerfile.worker               # Multi-stage Go build for Worker
├── extension/
│   ├── manifest.json                   # Chrome extension manifest v3
│   ├── background.js                   # Badge updates service worker
│   ├── popup.html/js/css               # Extension popup (auto-save on click)
│   ├── saved.html/js/css               # Full-page saved pages view
│   ├── options.html/js                 # API URL settings
│   └── icons/                          # Extension icons (16-128px)
├── scripts/
│   ├── install.sh                      # System installer (systemd, desktop entry)
│   ├── mesh-services.sh                # Docker Compose lifecycle manager
│   ├── mesh-tray.sh                    # Tray icon launcher
│   ├── mesh-tray.py                    # AppIndicator3 tray (Wayland-compatible)
│   ├── mesh.service                    # systemd user service unit
│   └── mesh.desktop                    # Desktop entry for autostart
├── docs/
│   ├── PROJECT_MESH_BLUEPRINT.md       # Full architectural blueprint
│   ├── DEVELOPERS_GUIDE.md             # Developer setup and conventions
│   ├── REVIEW_CHECKLIST.md             # Codebase audit framework
│   └── PROJECT_PROGRESS.md            # This file
└── .idea/                              # GoLand IDE configuration
```

### What Does NOT Exist Yet

| Category | Items Missing |
|----------|--------------|
| **Workers** | No job queue claim logic, no processor (Week 3) |
| **Ingestion** | No scraper, no circuit breaker (Week 3) |
| **Tests** | No test files (Week 3) |
| **Frontend** | No React project (Phase 4) |

---

## Phase Progress

### Phase 0: Planning & Documentation — COMPLETE

| Item | Status |
|------|--------|
| Architectural blueprint | Done |
| Tech stack selection with justifications | Done |
| API endpoint specification | Done |
| SQL schema design | Done |
| Key query patterns documented | Done |
| Docker Compose topology designed | Done |
| Risk assessment | Done |
| Implementation roadmap (7 phases, 30 weeks) | Done |
| README | Done |
| Developer's Guide | Done |
| Review Checklist | Done |
| Progress Tracker | Done |

### Phase 1: Foundation & Ingestion — "The Senses" — IN PROGRESS

**Progress: 15/21 items**

- [x] Initialize Go module (`github.com/neha037/mesh`)
- [x] Create project directory structure
- [x] Initialize Git repository, add `.gitignore`
- [x] Set up Docker Compose with PostgreSQL 16 + pgvector
- [x] Write initial SQL migration (all 7 tables + indexes)
- [x] Configure `golang-migrate`
- [x] Create `Makefile`
- [x] Create `Dockerfile.api` (multi-stage build)
- [x] Create `.env.example`
- [x] Implement HTTP server with chi router
- [ ] `POST /api/v1/ingest/url` endpoint
- [x] `POST /api/v1/ingest/raw` endpoint (URL + title + content)
- [x] `GET /api/v1/nodes/recent` endpoint
- [x] `GET /api/v1/nodes` endpoint (paginated)
- [x] `DELETE /api/v1/nodes/{id}` endpoint
- [ ] `POST /api/v1/ingest/text` endpoint
- [x] PostgreSQL repository layer (pgx + sqlc)
- [x] Request logging middleware
- [x] CORS middleware
- [x] URL deduplication (upsert on conflict)
- [x] Keyset (cursor) pagination for scalability
- [x] Connection pool configuration (MinConns=5, MaxConns=25)
- [ ] Web scraper (colly, timeouts, robots.txt)
- [ ] Circuit breaker (sony/gobreaker)
- [ ] Job queue claim logic (FOR UPDATE SKIP LOCKED)
- [ ] Unit tests for handlers (table-driven)
- [ ] Integration tests with testcontainers-go

### Phases 2-7 — NOT STARTED

See [Review Checklist](REVIEW_CHECKLIST.md) for detailed per-phase checklists.

---

## What's Next

Week 1 scaffolding is complete. Next is **Week 2 — Ingestion API**:

1. **Implement HTTP server** with `chi` router and signal handling
2. **`POST /api/v1/ingest/url`** — validate URL, create node, enqueue `process_url` job, return `202 Accepted`
3. **`POST /api/v1/ingest/text`** — validate title/content, create node, enqueue `process_text` job, return `201 Created`
4. **PostgreSQL repository layer** using `pgx` + `sqlc`
5. **Request logging middleware**
6. **CORS middleware**

---

*This document is updated after each work session. For the full architectural design, see [PROJECT_MESH_BLUEPRINT.md](PROJECT_MESH_BLUEPRINT.md).*
