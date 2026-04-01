# Mesh

A localized, private **Personal Growth Engine** that maps structured technical knowledge and creative pursuits into a unified, interactive knowledge graph — with algorithmic serendipity to combat intellectual stagnation.

## Why Mesh?

Every existing PKM tool is passive. They rely on you to manually create connections and review old material. Mesh is an **active cognitive partner** that:

- Visualizes your knowledge as an interactive graph
- Automatically extracts concepts and builds connections
- Detects knowledge gaps and over-saturated clusters
- Injects serendipitous cross-domain discoveries
- Uses spaced repetition to fight knowledge decay

All running **locally** on your machine — zero cloud costs, total data sovereignty.

## Architecture

```
┌──────────────────┐
│  React Frontend  │ ← Cytoscape.js graph, search, filters, review cards
│    (Port 3000)   │
└────────┬─────────┘
         │ REST/JSON
┌────────┴─────────┐
│   Go API Server  │ ← chi router, ingestion, graph queries, search
│    (Port 8080)   │
└────────┬─────────┘
         │
┌────────┴─────────┐
│  Go Workers      │ ← scraping, NLP tagging, embedding, edge building
└──┬─────┬─────┬───┘
   │     │     │
   ▼     ▼     ▼
 PG16  MinIO  Ollama
```

## Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Backend | Go + chi | REST API and background workers |
| Database | PostgreSQL 16 + pgvector | Graph storage, vector search, job queue |
| Object Store | MinIO | Image and file storage (S3-compatible) |
| AI/NLP | Ollama (local) | Tag extraction, embeddings, summaries |
| Frontend | React + TypeScript + Cytoscape.js | Interactive graph visualization |
| Orchestration | Docker Compose | Service topology and lifecycle |

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Git

### Run

```bash
# Clone the repository
git clone https://github.com/neha037/mesh.git
cd mesh

# Configure environment
cp .env.example .env
# Edit .env and set PG_PASSWORD and MINIO_PASSWORD

# Start core services (PostgreSQL, MinIO, API, Worker, Web)
cd deploy && docker compose up -d

# Run database migrations
make migrate-up

# (Optional) Start Ollama for AI features
docker compose --profile ai up -d ollama
make pull-models
```

The web UI will be available at `http://localhost:3000` and the API at `http://localhost:8080`.

### Verify

```bash
# Ingest a test URL
curl -X POST http://localhost:8080/api/v1/ingest/url \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "type": "article"}'
```

## Development

See the [Developer's Guide](docs/DEVELOPERS_GUIDE.md) for detailed setup, workflows, and conventions.

```bash
# Run with hot-reload (development mode)
cd deploy && docker compose -f docker-compose.yml -f docker-compose.dev.yml up

# Run tests
make test

# Run linter
make lint
```

## Project Status

**Current Phase:** Phase 1 — Foundation & Ingestion ("The Senses")

| Phase | Name | Status |
|-------|------|--------|
| 1 | Foundation & Ingestion — "The Senses" | In progress (Week 1 scaffolding complete) |
| 2 | Processing & Intelligence — "The Brain" | Not started |
| 3 | Graph Traversal & Query API — "The Memory" | Not started |
| 4 | Frontend Visualization — "The Eyes" | Not started |
| 5 | Multi-Modal & Journaling — "The Human Element" | Not started |
| 6 | Anti-Echo Chamber Engine — "Discovery" | Not started |
| 7 | Spaced Repetition & Semantic Depth — "The Slow Burn" | Not started |

## Documentation

- [Project Blueprint](docs/PROJECT_MESH_BLUEPRINT.md) — full architectural design, data model, and roadmap
- [Developer's Guide](docs/DEVELOPERS_GUIDE.md) — setup, workflows, and conventions
- [Review Checklist](docs/REVIEW_CHECKLIST.md) — codebase audit framework

## Design Constraints

- **Zero cloud costs** — all compute and storage runs locally
- **Single developer** — 6-8 hours/week sustainable pace
- **Absolute data sovereignty** — no data leaves the machine
- **Standard hardware** — runs on 16 GB RAM workstation
- **Ephemeral compute** — safe to shut down at any time
