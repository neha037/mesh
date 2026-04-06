# Mesh Developer's Guide

This guide covers everything needed to develop, test, and extend Mesh. It will be updated as the project evolves through each phase.

---

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.24+ | Backend API and workers |
| Docker | 24+ | Container runtime |
| Docker Compose | v2+ | Service orchestration |
| Node.js | 20+ | Frontend development (Phase 4+) |
| Make | any | Build automation |
| golang-migrate CLI | latest | Database migrations |
| sqlc | latest | Type-safe SQL codegen |
| golangci-lint | latest | Go linting |

### Install golang-migrate

```bash
# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# macOS
brew install golang-migrate
```

### Install sqlc

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

---

## Repository Structure

```
mesh/
├── cmd/                    # Application entrypoints
│   ├── api/main.go         # HTTP API server
│   ├── worker/main.go      # Background job workers
│   └── discovery/main.go   # Discovery engine (Phase 6)
├── extension/              # Chrome browser extension (Manifest V3)
│   ├── manifest.json       # Extension config, permissions
│   ├── popup.html/js/css   # Popup with auto-save on click
│   ├── saved.html/js/css   # Full-page view of all saved pages
│   ├── options.html/js     # Settings (API URL)
│   └── icons/              # Extension icons
├── internal/               # Private application code
│   ├── api/                # HTTP handlers, middleware, router
│   ├── config/             # Environment-based configuration
│   ├── domain/             # Core types and interfaces
│   ├── nlp/                # NLP service and fallback extractor
│   ├── ollama/             # Ollama HTTP client (tags, embeddings)
│   ├── scraper/            # Web scraping and circuit breaker
│   ├── storage/            # PostgreSQL repositories (pgx + sqlc)
│   └── worker/             # Worker pool and job processor
├── migrations/             # SQL migration files (up/down pairs)
├── web/                    # React frontend (Phase 4+)
├── deploy/                 # Docker Compose, Dockerfiles, nginx config
├── scripts/                # Utility scripts and system integration
│   ├── install.sh          # Installer (systemd service, desktop entry)
│   ├── mesh-services.sh    # Docker Compose lifecycle manager
│   ├── mesh-tray.sh/.py    # System tray icon (AppIndicator3/Wayland)
├── .env.example            # Environment variable template
├── go.mod / go.sum         # Go module files
├── sqlc.yaml               # sqlc configuration
└── Makefile                # Build, test, run targets
```

---

## Local Development Workflow

### 1. Start Infrastructure

```bash
# Start PostgreSQL and MinIO only
cd deploy && docker compose up -d postgres minio

# Verify health
docker compose ps
```

### 2. Run Migrations

```bash
# Set DATABASE_URL (or export in your shell profile)
export DATABASE_URL="postgres://mesh:yourpassword@localhost:5432/mesh?sslmode=disable"

# Apply all migrations
make migrate-up

# Rollback last migration
make migrate-down
```

### 3. Run the API (locally, outside Docker)

```bash
# Set required environment variables
export DATABASE_URL="postgres://mesh:yourpassword@localhost:5432/mesh?sslmode=disable"
export MINIO_ENDPOINT="localhost:9000"
export MINIO_ACCESS_KEY="meshadmin"
export MINIO_SECRET_KEY="yourpassword"
export OLLAMA_HOST="http://localhost:11434"

make run-api
```

### 4. Run Workers (locally)

```bash
make run-worker
```

### 5. Full Stack with Hot-Reload (Docker)

```bash
cd deploy
docker compose -f docker-compose.yml -f docker-compose.dev.yml up
```

This mounts source code into containers and uses `go run` for live reloading.

---

## Database Workflow

### Creating a New Migration

```bash
# Create up/down pair
migrate create -ext sql -dir migrations -seq <description>
# Example: migrate create -ext sql -dir migrations -seq add_review_schedule
```

This creates:
- `migrations/00X_<description>.up.sql`
- `migrations/00X_<description>.down.sql`

**Rules:**
- Every `.up.sql` must have a matching `.down.sql` that fully reverses the change
- Test both directions: `make migrate-up && make migrate-down && make migrate-up`
- Never modify an already-applied migration; create a new one instead

### sqlc Workflow

1. Write SQL queries in `internal/storage/queries/*.sql`
2. Run `make sqlc` to generate Go code
3. Generated code appears alongside the query files

**sqlc query annotations:**

```sql
-- name: GetNode :one
SELECT * FROM nodes WHERE id = $1;

-- name: ListNodes :many
SELECT id, title, source_url, created_at FROM nodes
WHERE (sqlc.narg('cursor')::TIMESTAMPTZ IS NULL OR created_at < sqlc.narg('cursor'))
ORDER BY created_at DESC LIMIT $1;

-- name: UpsertRawNode :one
INSERT INTO nodes (type, title, content, source_url) VALUES ($1, $2, $3, $4)
ON CONFLICT (source_url) WHERE source_url IS NOT NULL
DO UPDATE SET title = EXCLUDED.title, content = EXCLUDED.content, updated_at = now()
RETURNING *;
```

### Connecting to the Database

```bash
# Via docker exec
docker exec -it mesh-postgres psql -U mesh -d mesh

# Via local psql
psql "postgres://mesh:yourpassword@localhost:5432/mesh?sslmode=disable"
```

---

## Adding a New API Endpoint

1. **Define the domain type** (if new) in `internal/domain/`
2. **Write the SQL query** in `internal/storage/queries/` and run `make sqlc`
3. **Create the handler** in `internal/api/handler/`:

```go
func (h *Handler) HandleNewEndpoint(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    // Parse input, validate, call storage, return response
}
```

4. **Register the route** in `internal/api/router.go`:

```go
r.Get("/api/v1/new-endpoint", h.HandleNewEndpoint)
```

5. **Write tests** — table-driven unit tests for the handler

---

## Adding a New Worker Job Type

1. **Add the job type** to the `jobs.type` CHECK constraint (new migration)
2. **Add the type constant** in `internal/domain/job.go`
3. **Create the processor** function in `internal/queue/processor.go`
4. **Register it** in the job type → handler dispatch map
5. **Write an integration test** that enqueues the job and verifies processing

---

## Testing

### Running Tests

```bash
# All tests (with race detector)
make test

# Integration tests only (requires Docker for testcontainers)
make test-integration

# Specific package
go test ./internal/storage/... -v -race

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Testing Strategy

| Type | Location | Tools | What it covers |
|------|----------|-------|---------------|
| Unit | `*_test.go` alongside code | `testing` stdlib | Handlers, domain logic, algorithms |
| Integration | `*_test.go` with `//go:build integration` | testcontainers-go | Real PostgreSQL, full pipelines |
| E2E | manual / scripts | curl, browser | Full stack verification |

### Conventions

- Use **table-driven tests** for handlers and functions with multiple input cases
- Integration tests use the `integration` build tag
- Tests run with `-race` to detect data races
- Never mock the database in integration tests — use testcontainers

---

## Makefile Reference

| Target | Description |
|--------|------------|
| `make build` | Build all Go binaries to `bin/` |
| `make run-api` | Run API server locally |
| `make run-worker` | Run worker locally |
| `make test` | Run all unit tests with race detector |
| `make test-integration` | Run integration tests |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Rollback last migration |
| `make docker-up` | Start all services via Docker Compose |
| `make docker-up-ai` | Start all services including Ollama |
| `make docker-down` | Stop all services |
| `make docker-logs` | Tail service logs |
| `make lint` | Run golangci-lint |
| `make sqlc` | Regenerate sqlc Go code |
| `make seed` | Seed database with sample data |
| `make backup` | Run PostgreSQL + MinIO backup |
| `make pull-models` | Download Ollama models |

---

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `PG_PASSWORD` | Yes (Docker) | — | PostgreSQL password (used in docker-compose) |
| `MINIO_ENDPOINT` | Yes | — | MinIO host:port |
| `MINIO_ACCESS_KEY` | Yes | `meshadmin` | MinIO access key |
| `MINIO_SECRET_KEY` | Yes | — | MinIO secret key |
| `MINIO_BUCKET` | No | `mesh-images` | MinIO bucket name |
| `OLLAMA_HOST` | No | `http://ollama:11434` | Ollama API URL |
| `OLLAMA_MODEL` | No | `gemma4:e4b` | LLM model for tag extraction, summarization, image/audio understanding |
| `EMBEDDING_MODEL` | No | `embeddinggemma:300m-qat-q8_0` | Embedding model (768-dim, Matryoshka) |
| `WORKER_COUNT` | No | `4` | Number of worker goroutines |
| `LOG_LEVEL` | No | `info` | Log level (debug, info, warn, error) |

---

## Ollama Setup

Ollama is optional. The system falls back to `jdkato/prose` for basic NLP when Ollama is unavailable.

### Starting Ollama

```bash
# Start with AI profile
cd deploy && docker compose --profile ai up -d ollama

# Download required models
make pull-models
```

### Models Used

| Model | Purpose | Size | RAM |
|-------|---------|------|-----|
| `embeddinggemma:300m-qat-q8_0` | 768-dim embeddings for semantic search (Matryoshka: 768/512/256/128) | ~338 MB | ~200 MB |
| `gemma4:e4b` | Tag extraction, summaries, image understanding, audio transcription | ~4.5 GB | ~6 GB |

### Stopping Ollama (to free RAM)

```bash
docker compose --profile ai stop ollama
```

### Circuit Breaker Protection

The Ollama client has built-in circuit breaker protection:
- **Open** after 3 consecutive failures
- **Half-open** automatically after 60 seconds to retry
- **Fast-fail**: `Healthy()` returns false immediately when breaker is open (no network call)
- **Shared** across all Ollama calls (tag extraction, embeddings)

When the circuit breaker is open, the system automatically falls back to `jdkato/prose` NLP without making HTTP requests to Ollama.

The system automatically detects Ollama availability and switches between AI and fallback NLP paths.

---

## Conventions

### Error Handling

```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("fetching node %s: %w", id, err)
}
```

### Logging

- Use structured logging (key-value pairs)
- Log levels: `debug` for development details, `info` for normal operations, `warn` for recoverable issues, `error` for failures
- Include request IDs in API logs for tracing

### Naming

- Go packages: short, lowercase, no underscores (`storage`, not `data_access`)
- HTTP handlers: `Handle<Action>` (e.g., `HandleCreateNode`)
- Repository methods: `Get`, `List`, `Create`, `Update`, `Delete` prefixes
- SQL migration files: sequential numbered with description

### Configuration

- All config via environment variables (12-factor)
- Loaded once at startup in `internal/config/config.go`
- No config files checked into the repo (except `.env.example`)

---

## Common Issues

| Problem | Solution |
|---------|----------|
| Port 5432 already in use | Stop local PostgreSQL: `sudo systemctl stop postgresql` |
| Migration fails | Check `DATABASE_URL` is set; verify PostgreSQL is running and healthy |
| Ollama circuit breaker open (tags/embeddings fail) | Wait 60s for recovery, or restart Ollama: `docker compose --profile ai restart ollama`. Check logs: `docker logs mesh-ollama` |
| Ollama OOM | Use quantized models (Q4_0); increase Docker memory limit |
| Tagging/embeddings slow when Ollama down | Circuit breaker opened after 3 failures; worker uses fallback NLP (slower, lower confidence) until Ollama recovers |
| sqlc errors | Ensure `sqlc.yaml` paths are correct; run `make migrate-up` first |
| testcontainers fail | Ensure Docker daemon is running; check Docker socket permissions. If Ryuk reaper crashes, `make test-integration` already sets `TESTCONTAINERS_RYUK_DISABLED=true` |
| Docker group not effective | If user is SSSD/FreeIPA-managed, `newgrp docker` only affects current shell. Use `sg docker -c "<command>"` or log out/in. The systemd service wraps with `sg docker` automatically |
| Tray icon not showing | `yad` doesn't work on Wayland. The tray uses AppIndicator3 (Python). Install `gnome-shell-extension-appindicator` and enable it |

---

*This guide is a living document. It will be updated as new phases are implemented.*
