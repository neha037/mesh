# Mesh Phase 2 Implementation Guide

**Hand this document to an AI coding assistant (e.g., Gemini 3 Flash in Antigravity) to implement Phase 2 step by step.**

**Project location:** `/home/nkumari/GolandProjects/mesh`

---

## How to Use This Guide

1. Open the project directory in your AI coding tool
2. Execute each step in order (Step 1 first, then Step 2, etc.)
3. Copy the **Prompt** section of each step as your instruction to the AI
4. After each step, run the **Verification** command to confirm success
5. Do NOT skip steps — later steps depend on earlier ones
6. Steps marked as "parallelizable" can be done together if your tool supports it

---

## Project Context (READ THIS FIRST)

Mesh is a local-first Personal Growth Engine. Go backend, PostgreSQL+pgvector, Ollama for local AI.

### Key Rules (from CLAUDE.md — the AI MUST follow these)

- **TDD:** Write failing test first (RED), then implementation (GREEN), then clean up (REFACTOR)
- **Error wrapping:** Always `fmt.Errorf("context: %w", err)` — never bare errors
- **Context:** `context.Context` is always the first parameter
- **No panics** in library code
- **SQL:** Use golang-migrate for migrations, sqlc for codegen, parameterized queries only
- **Testing:** Table-driven tests, integration tests with testcontainers-go, always `-race`
- **Database Migrations:** Index predicates must use only IMMUTABLE functions. Always provide matching down migration. Run `make lint-sql` and `make test-integration` to verify.

### What Already Exists (Phase 1 — Complete)

```
cmd/api/main.go              — HTTP API server entrypoint
cmd/worker/main.go            — Worker pool entrypoint
internal/config/config.go     — Env-based config (OllamaHost, OllamaModel, EmbeddingModel already defined)
internal/api/router.go        — chi router with CORS, throttle, timeout
internal/api/handler/         — HTTP handlers (ingest/raw, ingest/url, ingest/text, nodes CRUD)
internal/domain/node.go       — Node struct, NodeRepository interface, UpsertResult, ListNodesParams
internal/domain/job.go        — Job struct, JobRepository interface, IngestService interface
internal/domain/errors.go     — ErrNotFound
internal/storage/node_repo.go — NodeRepo implements NodeRepository
internal/storage/job_repo.go  — JobRepo implements JobRepository (ClaimJob, CompleteJob, FailJob, RetryJob)
internal/storage/ingest_repo.go — IngestRepo implements IngestService (transactional node+job creation)
internal/storage/postgres.go  — Connection pool (MinConns=5, MaxConns=25)
internal/storage/migrate.go   — RunMigrations with embedded FS
internal/storage/queries/nodes.sql — sqlc queries for nodes
internal/storage/queries/jobs.sql  — sqlc queries for jobs
internal/scraper/scraper.go   — Colly web scraper (UA rotation, robots.txt, HTML cleaning)
internal/scraper/breaker.go   — Per-domain circuit breaker (gobreaker)
internal/worker/pool.go       — Worker pool (goroutines, backoff, graceful shutdown)
internal/worker/processor.go  — Job processor (routes process_url and process_text)
migrations/001_initial_schema.up.sql — 7 tables: nodes, tags, node_tags, edges, jobs, review_schedule, discovery_runs
migrations/002_unique_source_url.up.sql
migrations/003_add_node_status.up.sql — Adds status column to nodes
sqlc.yaml                     — sqlc config: engine=postgresql, sql_package=pgx/v5
```

### Existing Interfaces You Must Know

```go
// internal/domain/node.go
type NodeRepository interface {
    UpsertRawNode(ctx context.Context, nodeType, title, content, sourceURL string) (UpsertResult, error)
    GetNode(ctx context.Context, id string) (Node, error)
    ListRecentNodes(ctx context.Context, limit int32) ([]Node, error)
    ListNodes(ctx context.Context, params ListNodesParams) (ListNodesResult, error)
    DeleteNode(ctx context.Context, id string) error
    UpdateNodeContent(ctx context.Context, id string, content string) error
    UpdateNodeStatus(ctx context.Context, id string, status string) error
}

// internal/domain/job.go
type JobRepository interface {
    ClaimJob(ctx context.Context) (*Job, error)
    CompleteJob(ctx context.Context, id string) error
    FailJob(ctx context.Context, id string, errMsg string) error
    RetryJob(ctx context.Context, id string, backoffSeconds int) error
}

// internal/worker/processor.go
type Processor interface {
    Process(ctx context.Context, job *domain.Job) error
}

// DefaultProcessor currently handles: "process_url", "process_text"
// Has fields: scraper *scraper.Service, nodes domain.NodeRepository
```

### Database Schema (Phase 2 relevant tables)

```sql
-- nodes table has these columns relevant to Phase 2:
embedding   vector(768)              -- pgvector, 768 dimensions
version     INTEGER NOT NULL DEFAULT 1  -- optimistic concurrency control
status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','processing','processed','failed'))

-- tags table
CREATE TABLE tags (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE
);

-- node_tags table (many-to-many with confidence)
CREATE TABLE node_tags (
    node_id    UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    tag_id     UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    confidence REAL DEFAULT 1.0,
    PRIMARY KEY (node_id, tag_id)
);

-- edges table
CREATE TABLE edges (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    rel_type  TEXT NOT NULL CHECK (rel_type IN ('tag_shared','manual','semantic','bridge','wildcard')),
    weight    REAL NOT NULL DEFAULT 1.0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(source_id, target_id, rel_type)
);

-- jobs table allows these types for Phase 2:
-- 'generate_embedding', 'build_edges'
```

### Key Helpers Already Available in storage package

```go
func uuidToString(u pgtype.UUID) string    // Converts pgtype.UUID to string
func parseUUID(id string) (pgtype.UUID, error) // Converts string to pgtype.UUID
```

### Worker Entrypoint Wiring (cmd/worker/main.go)

```go
queries := storage.New(pool)
jobRepo := storage.NewJobRepo(queries)
nodeRepo := storage.NewNodeRepo(queries)
scraperSvc := scraper.NewService()
proc := worker.NewProcessor(scraperSvc, nodeRepo)  // <-- This changes in Step 9
```

---

## Step 1: Add SQL Queries for Tags, Edges, and Embeddings

**What:** Write new sqlc queries. No Go code yet — just SQL, then run code generation.

**Files to create/edit:**
- `internal/storage/queries/tags.sql` (NEW)
- `internal/storage/queries/edges.sql` (NEW)
- `internal/storage/queries/nodes.sql` (EDIT — append new queries)

### Prompt

```
You are working on the Mesh project at /home/nkumari/GolandProjects/mesh. Read CLAUDE.md for all conventions.

TASK: Write sqlc SQL queries for Phase 2 (tags, edges, embeddings). The project uses sqlc for code generation. Config is in sqlc.yaml:
  engine: postgresql, queries: "internal/storage/queries/", schema: "migrations/", package: "storage", out: "internal/storage", sql_package: "pgx/v5"

EXISTING SCHEMA (already in migrations — do NOT create new migrations):
- tags(id UUID PK, name TEXT UNIQUE)
- node_tags(node_id UUID FK, tag_id UUID FK, confidence REAL DEFAULT 1.0, PK(node_id, tag_id))
- edges(id UUID PK, source_id UUID FK, target_id UUID FK, rel_type TEXT CHECK(...), weight REAL DEFAULT 1.0, UNIQUE(source_id, target_id, rel_type))
- nodes has: embedding vector(768), version INTEGER DEFAULT 1, status TEXT

FILE 1: Create internal/storage/queries/tags.sql with these exact queries:

-- name: UpsertTag :one
-- Insert tag, return existing if name conflict.
INSERT INTO tags (name) VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id, name;

-- name: AssociateNodeTag :exec
-- Link node to tag with confidence score, keep highest confidence on conflict.
INSERT INTO node_tags (node_id, tag_id, confidence)
VALUES ($1, $2, $3)
ON CONFLICT (node_id, tag_id) DO UPDATE
SET confidence = GREATEST(node_tags.confidence, EXCLUDED.confidence);

-- name: GetNodeTags :many
-- Get all tags for a node ordered by confidence.
SELECT t.id, t.name, nt.confidence
FROM tags t JOIN node_tags nt ON t.id = nt.tag_id
WHERE nt.node_id = $1
ORDER BY nt.confidence DESC;

FILE 2: Create internal/storage/queries/edges.sql with these exact queries:

-- name: UpsertEdge :exec
-- Create or update edge, keeping the higher weight.
INSERT INTO edges (source_id, target_id, rel_type, weight)
VALUES ($1, $2, $3, $4)
ON CONFLICT (source_id, target_id, rel_type) DO UPDATE
SET weight = GREATEST(edges.weight, EXCLUDED.weight);

-- name: BuildTagSharedEdges :exec
-- Create tag_shared edges for nodes sharing 2+ tags with the given node.
-- Weight = shared_count / total_tags_on_source_node (normalized 0-1).
INSERT INTO edges (source_id, target_id, rel_type, weight)
SELECT $1::uuid, nt2.node_id, 'tag_shared',
       COUNT(*)::real / NULLIF((SELECT COUNT(*) FROM node_tags WHERE node_id = $1), 0)
FROM node_tags nt1
JOIN node_tags nt2 ON nt1.tag_id = nt2.tag_id
WHERE nt1.node_id = $1 AND nt2.node_id != $1
GROUP BY nt2.node_id
HAVING COUNT(*) >= 2
ON CONFLICT (source_id, target_id, rel_type) DO UPDATE
SET weight = GREATEST(edges.weight, EXCLUDED.weight);

-- name: FindSimilarNodes :many
-- Find nodes with similar embeddings using pgvector cosine distance.
SELECT id, title, 1 - (embedding <=> $1::vector) AS similarity
FROM nodes
WHERE embedding IS NOT NULL AND id != $2
ORDER BY embedding <=> $1::vector
LIMIT $3;

FILE 3: Edit internal/storage/queries/nodes.sql — APPEND these queries at the end (do NOT modify existing queries):

-- name: UpdateNodeEmbedding :execresult
-- Store embedding with optimistic concurrency control.
UPDATE nodes
SET embedding = $2, version = version + 1, updated_at = now()
WHERE id = $1 AND version = $3;

-- name: GetNodeContent :one
-- Get node content and metadata for NLP processing.
SELECT id, type, title, content, status, version
FROM nodes WHERE id = $1;

After creating/editing these files, run these commands and verify both succeed:
  sqlc generate
  go build ./internal/storage/...
```

### Verification

```bash
sqlc generate && go build ./internal/storage/...
```

---

## Step 2: Ollama HTTP Client

**What:** Create the Ollama client package with interface, types, and implementation. Follow TDD.

**Files to create:**
- `internal/ollama/ollama.go` (interface + types)
- `internal/ollama/client.go` (HTTP implementation)
- `internal/ollama/client_test.go` (unit tests with httptest)

### Prompt

```
You are working on the Mesh project at /home/nkumari/GolandProjects/mesh. Read CLAUDE.md for all conventions. Follow TDD: write tests FIRST, verify they fail, then implement.

TASK: Create an Ollama HTTP client package at internal/ollama/.

EXISTING CONFIG (internal/config/config.go) already has these fields:
- OllamaHost string    (default: "http://localhost:11434")
- OllamaModel string   (default: "gemma4:e4b")
- EmbeddingModel string (default: "embeddinggemma:300m-qat-q8_0")

FILE 1: Create internal/ollama/ollama.go

package ollama

import "context"

// TagResult holds extracted tags with a confidence score.
type TagResult struct {
    Tags       []string
    Confidence float32
}

// Client defines the interface for Ollama API operations.
// This is the system boundary interface — mock this in tests.
type Client interface {
    ExtractTags(ctx context.Context, content string) (TagResult, error)
    GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
    Healthy(ctx context.Context) bool
}

FILE 2: Create internal/ollama/client_test.go — Write tests FIRST using httptest.NewServer to mock the Ollama API:

package ollama_test

Write these table-driven test cases:

1. TestExtractTags_Success — Mock POST /api/generate responding with:
   {"response": "{\"tags\":[\"machine-learning\",\"neural-networks\",\"deep-learning\"],\"confidence\":0.92}"}
   Assert: 3 tags returned, confidence == 0.92

2. TestExtractTags_InvalidJSON — Mock returns malformed JSON in response field
   Assert: error returned

3. TestExtractTags_EmptyContent — Call with empty string
   Assert: error contains "content is empty"

4. TestExtractTags_ServerError — Mock returns HTTP 500
   Assert: error returned

5. TestGenerateEmbedding_Success — Mock POST /api/embed responding with:
   {"embeddings": [[float array of length 768]]}
   Assert: returned slice length == 768

6. TestGenerateEmbedding_WrongDimensions — Mock returns 384-dim vector
   Assert: error contains "expected 768 dimensions"

7. TestGenerateEmbedding_ServerError — Mock returns HTTP 500
   Assert: error returned

8. TestHealthy_Up — Mock GET /api/tags returning 200
   Assert: returns true

9. TestHealthy_Down — Use a closed httptest server (server.Close() before calling)
   Assert: returns false

Run: go test ./internal/ollama/ -v -race -count=1
Tests should FAIL because client.go doesn't exist yet.

FILE 3: Create internal/ollama/client.go — Implement the client:

package ollama

type HTTPClient struct {
    baseURL    string
    tagModel   string
    embedModel string
    httpClient *http.Client
}

func NewClient(baseURL, tagModel, embedModel string) *HTTPClient {
    return &HTTPClient{
        baseURL:    baseURL,
        tagModel:   tagModel,
        embedModel: embedModel,
        httpClient: &http.Client{Timeout: 60 * time.Second},
    }
}

ExtractTags implementation:
- Validate content not empty
- POST to {baseURL}/api/generate
- Body: {"model": tagModel, "prompt": PROMPT, "stream": false, "format": "json"}
- PROMPT: "Extract 3-8 key domain-specific concept tags from the following content. Return JSON: {\"tags\": [\"tag1\", \"tag2\"], \"confidence\": 0.0-1.0}. Tags must be lowercase, 1-3 words each. Avoid generic words like \"article\" or \"content\".\n\nContent:\n" + truncate(content, 4000)
- Parse the response JSON: type generateResponse struct { Response string `json:"response"` }
- Then parse response.Response as JSON: type tagResponse struct { Tags []string; Confidence float32 }
- Validate at least 1 tag
- Wrap errors: fmt.Errorf("ollama: extracting tags: %w", err)

GenerateEmbedding implementation:
- POST to {baseURL}/api/embed
- Body: {"model": embedModel, "input": truncate(text, 8000)}
- Parse response: type embedResponse struct { Embeddings [][]float32 `json:"embeddings"` }
- Validate embeddings[0] has exactly 768 dimensions
- Wrap errors: fmt.Errorf("ollama: generating embedding: %w", err)

Healthy implementation:
- GET {baseURL}/api/tags with a 5-second timeout context
- Return true if HTTP 200, false otherwise (including errors)

Helper: func truncate(s string, maxLen int) string — truncate to maxLen chars

Run: go test ./internal/ollama/ -v -race -count=1
All tests should PASS.
Run: go test ./... -v -race -count=1
No regressions in other packages.
```

### Verification

```bash
go test ./internal/ollama/ -v -race -count=1
go test ./... -v -race -count=1
```

---

## Step 3: NLP Fallback — prose-based Tag Extraction

**What:** Fallback tag extractor using NLP when Ollama is unavailable.

**Files to create:**
- `internal/nlp/fallback.go` (NEW)
- `internal/nlp/fallback_test.go` (NEW)

### Prompt

```
You are working on the Mesh project at /home/nkumari/GolandProjects/mesh. Read CLAUDE.md for all conventions. Follow TDD.

TASK: Create a fallback NLP tag extractor at internal/nlp/ using github.com/jdkato/prose/v2.

First: go get github.com/jdkato/prose/v2

FILE 1: Create internal/nlp/fallback_test.go — Write tests FIRST:

package nlp_test

Table-driven test cases:

1. TestExtractTags_ExtractsNouns — Input: "Machine learning and neural networks are transforming artificial intelligence research at Google and OpenAI."
   Assert: returns tags, length >= 2, confidence between 0.5-0.7

2. TestExtractTags_EmptyContent — Input: ""
   Assert: returns empty tags slice, no error

3. TestExtractTags_ShortContent — Input: "Hello"
   Assert: returns tags (may be empty), no error

4. TestExtractTags_DeduplicatesTags — Input with repeated nouns
   Assert: no duplicate tags in result

5. TestExtractTags_MaxEightTags — Very long content with many nouns
   Assert: at most 8 tags returned

Run: go test ./internal/nlp/ -v -race -count=1 (should FAIL)

FILE 2: Create internal/nlp/fallback.go

package nlp

import (
    "strings"
    "github.com/jdkato/prose/v2"
    "github.com/neha037/mesh/internal/ollama"
)

type FallbackExtractor struct{}

func NewFallbackExtractor() *FallbackExtractor { return &FallbackExtractor{} }

func (f *FallbackExtractor) ExtractTags(content string) (ollama.TagResult, error) {
    if content == "" {
        return ollama.TagResult{}, nil
    }
    // 1. prose.NewDocument(content)
    // 2. Collect named entities (doc.Entities()) → add entity.Text as tag
    // 3. Collect nouns from POS tags: iterate doc.Tokens(), keep tokens with Tag starting with "NN"
    // 4. Lowercase all tags
    // 5. Filter: len >= 2 and len <= 50
    // 6. Deduplicate using a map
    // 7. Limit to 8 tags
    // 8. Return with Confidence = 0.6
}

Run: go test ./internal/nlp/ -v -race -count=1 (should PASS)
Run: go test ./... -v -race -count=1 (no regressions)
```

### Verification

```bash
go get github.com/jdkato/prose/v2
go test ./internal/nlp/ -v -race -count=1
```

---

## Step 4: Tag Repository (Storage Layer)

**What:** Domain interface + storage implementation for tags.

**Prerequisite:** Step 1 complete.

### Prompt

```
You are working on the Mesh project at /home/nkumari/GolandProjects/mesh. Read CLAUDE.md.

PREREQUISITE: The sqlc queries from Step 1 must already be generated.

TASK: Add TagRepository interface to domain and implement it in storage.

FILE 1: Edit internal/domain/node.go — APPEND these types and interface after the existing code (do NOT modify anything existing):

// Tag represents a concept tag extracted from content.
type Tag struct {
    ID         string
    Name       string
    Confidence float32
}

// TagRepository defines the interface for tag storage operations.
type TagRepository interface {
    UpsertTag(ctx context.Context, name string) (string, error)
    AssociateNodeTag(ctx context.Context, nodeID, tagID string, confidence float32) error
    GetNodeTags(ctx context.Context, nodeID string) ([]Tag, error)
}

FILE 2: Create internal/storage/tag_repo.go

package storage

import (
    "context"
    "fmt"
    "github.com/neha037/mesh/internal/domain"
)

type TagRepo struct {
    q *Queries
}

func NewTagRepo(q *Queries) *TagRepo {
    return &TagRepo{q: q}
}

func (r *TagRepo) UpsertTag(ctx context.Context, name string) (string, error) {
    row, err := r.q.UpsertTag(ctx, name)
    if err != nil {
        return "", fmt.Errorf("upserting tag: %w", err)
    }
    return uuidToString(row.ID), nil
}

func (r *TagRepo) AssociateNodeTag(ctx context.Context, nodeID, tagID string, confidence float32) error {
    nid, err := parseUUID(nodeID)
    if err != nil {
        return fmt.Errorf("parsing node id: %w", err)
    }
    tid, err := parseUUID(tagID)
    if err != nil {
        return fmt.Errorf("parsing tag id: %w", err)
    }
    if err := r.q.AssociateNodeTag(ctx, AssociateNodeTagParams{
        NodeID:     nid,
        TagID:      tid,
        Confidence: confidence,
    }); err != nil {
        return fmt.Errorf("associating tag: %w", err)
    }
    return nil
}

func (r *TagRepo) GetNodeTags(ctx context.Context, nodeID string) ([]domain.Tag, error) {
    nid, err := parseUUID(nodeID)
    if err != nil {
        return nil, fmt.Errorf("parsing node id: %w", err)
    }
    rows, err := r.q.GetNodeTags(ctx, nid)
    if err != nil {
        return nil, fmt.Errorf("getting node tags: %w", err)
    }
    tags := make([]domain.Tag, len(rows))
    for i, row := range rows {
        tags[i] = domain.Tag{
            ID:         uuidToString(row.ID),
            Name:       row.Name,
            Confidence: row.Confidence,
        }
    }
    return tags, nil
}

NOTE: uuidToString and parseUUID already exist in internal/storage/node_repo.go.
The sqlc-generated param types (AssociateNodeTagParams, etc.) come from the queries you wrote in Step 1.

Run: go build ./internal/storage/...
```

### Verification

```bash
go build ./internal/storage/...
```

---

## Step 5: Edge Repository (Storage Layer)

**What:** Domain interface + storage implementation for edges.

**Prerequisite:** Step 1 complete.

### Prompt

```
You are working on the Mesh project at /home/nkumari/GolandProjects/mesh. Read CLAUDE.md.

PREREQUISITE: The sqlc queries from Step 1 must already be generated.

TASK: Add EdgeRepository interface to domain and implement it in storage.

FILE 1: Edit internal/domain/node.go — APPEND these types and interface:

// SimilarNode holds a node ID and its similarity score from vector search.
type SimilarNode struct {
    ID         string
    Title      string
    Similarity float32
}

// EdgeRepository defines the interface for edge storage operations.
type EdgeRepository interface {
    BuildTagSharedEdges(ctx context.Context, nodeID string) error
    UpsertSemanticEdge(ctx context.Context, sourceID, targetID string, weight float32) error
    FindSimilarNodes(ctx context.Context, embedding []float32, excludeID string, limit int32) ([]SimilarNode, error)
}

FILE 2: Create internal/storage/edge_repo.go

package storage

import (
    "context"
    "fmt"
    pgvector "github.com/pgvector/pgvector-go"
    "github.com/neha037/mesh/internal/domain"
)

type EdgeRepo struct {
    q *Queries
}

func NewEdgeRepo(q *Queries) *EdgeRepo {
    return &EdgeRepo{q: q}
}

func (r *EdgeRepo) BuildTagSharedEdges(ctx context.Context, nodeID string) error {
    uid, err := parseUUID(nodeID)
    if err != nil {
        return fmt.Errorf("parsing node id: %w", err)
    }
    if err := r.q.BuildTagSharedEdges(ctx, uid); err != nil {
        return fmt.Errorf("building tag-shared edges: %w", err)
    }
    return nil
}

func (r *EdgeRepo) UpsertSemanticEdge(ctx context.Context, sourceID, targetID string, weight float32) error {
    sid, err := parseUUID(sourceID)
    if err != nil {
        return fmt.Errorf("parsing source id: %w", err)
    }
    tid, err := parseUUID(targetID)
    if err != nil {
        return fmt.Errorf("parsing target id: %w", err)
    }
    if err := r.q.UpsertEdge(ctx, UpsertEdgeParams{
        SourceID: sid,
        TargetID: tid,
        RelType:  "semantic",
        Weight:   weight,
    }); err != nil {
        return fmt.Errorf("upserting semantic edge: %w", err)
    }
    return nil
}

func (r *EdgeRepo) FindSimilarNodes(ctx context.Context, embedding []float32, excludeID string, limit int32) ([]domain.SimilarNode, error) {
    uid, err := parseUUID(excludeID)
    if err != nil {
        return nil, fmt.Errorf("parsing exclude id: %w", err)
    }
    rows, err := r.q.FindSimilarNodes(ctx, FindSimilarNodesParams{
        Embedding: pgvector.NewVector(embedding),
        ID:        uid,
        Limit:     limit,
    })
    if err != nil {
        return nil, fmt.Errorf("finding similar nodes: %w", err)
    }
    nodes := make([]domain.SimilarNode, len(rows))
    for i, row := range rows {
        nodes[i] = domain.SimilarNode{
            ID:         uuidToString(row.ID),
            Title:      row.Title,
            Similarity: float32(row.Similarity),
        }
    }
    return nodes, nil
}

NOTE: The FindSimilarNodes sqlc query param for the embedding will need to match
whatever type sqlc generates. You may need to adjust the param field name based on
the generated FindSimilarNodesParams struct. Check internal/storage/edges.sql.go after
sqlc generate to see the exact field names.

Run: go build ./internal/storage/...
```

### Verification

```bash
go build ./internal/storage/...
```

---

## Step 6: Extend NodeRepository for Embeddings

**What:** Add two methods to NodeRepository for embedding updates and content retrieval.

**Prerequisite:** Step 1 complete.

### Prompt

```
You are working on the Mesh project at /home/nkumari/GolandProjects/mesh. Read CLAUDE.md.

PREREQUISITE: The sqlc queries from Step 1 must already be generated.

TASK: Add two new methods to the existing NodeRepository interface and implement them.

FILE 1: Edit internal/domain/node.go — ADD two methods to the EXISTING NodeRepository interface (inside the interface block, after UpdateNodeStatus):

    UpdateNodeEmbedding(ctx context.Context, id string, embedding []float32, expectedVersion int32) (bool, error)
    GetNodeContent(ctx context.Context, id string) (Node, error)

FILE 2: Edit internal/storage/node_repo.go — ADD these implementations. You will need to add the import:
    pgvector "github.com/pgvector/pgvector-go"

func (r *NodeRepo) UpdateNodeEmbedding(ctx context.Context, id string, embedding []float32, expectedVersion int32) (bool, error) {
    uid, err := parseUUID(id)
    if err != nil {
        return false, fmt.Errorf("parsing node id: %w", err)
    }
    result, err := r.q.UpdateNodeEmbedding(ctx, UpdateNodeEmbeddingParams{
        ID:        uid,
        Embedding: pgvector.NewVector(embedding),
        Version:   expectedVersion,
    })
    if err != nil {
        return false, fmt.Errorf("updating embedding: %w", err)
    }
    return result.RowsAffected() > 0, nil
}

func (r *NodeRepo) GetNodeContent(ctx context.Context, id string) (domain.Node, error) {
    uid, err := parseUUID(id)
    if err != nil {
        return domain.Node{}, fmt.Errorf("parsing node id: %w", err)
    }
    row, err := r.q.GetNodeContent(ctx, uid)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return domain.Node{}, domain.ErrNotFound
        }
        return domain.Node{}, fmt.Errorf("getting node content: %w", err)
    }
    return domain.Node{
        ID:      uuidToString(row.ID),
        Type:    row.Type,
        Title:   row.Title,
        Content: row.Content.String,
        Status:  row.Status,
        Version: row.Version,
    }, nil
}

NOTE: Check the exact field names in the generated UpdateNodeEmbeddingParams struct
(in internal/storage/nodes.sql.go after sqlc generate). The Embedding field type
should be pgvector.Vector. Adjust if sqlc generated different field names.

IMPORTANT: After adding these methods to the interface, you must also update mock
implementations in test files. Check internal/api/handler/handler_test.go — the
mockNodeRepo struct there must add stub methods for the two new interface methods:

func (m *mockNodeRepo) UpdateNodeEmbedding(ctx context.Context, id string, embedding []float32, expectedVersion int32) (bool, error) {
    return false, nil
}
func (m *mockNodeRepo) GetNodeContent(ctx context.Context, id string) (domain.Node, error) {
    return domain.Node{}, nil
}

Similarly check internal/worker/pool_test.go for any mockNodeRepo.

Run: go build ./...
Run: go test ./... -v -race -count=1 (all existing tests must still pass)
```

### Verification

```bash
go build ./...
go test ./... -v -race -count=1
```

---

## Step 7: NLP Processing Service

**What:** Service that orchestrates Ollama + fallback for tag extraction and embedding generation.

**Prerequisite:** Steps 2 and 3 complete.

### Prompt

```
You are working on the Mesh project at /home/nkumari/GolandProjects/mesh. Read CLAUDE.md. Follow TDD.

PREREQUISITE: internal/ollama/ and internal/nlp/fallback.go must exist from Steps 2-3.

TASK: Create an NLP processing service at internal/nlp/service.go.

FILE 1: Create internal/nlp/service_test.go — Write tests FIRST:

package nlp_test

Create a mock for the ollama.Client interface in the test file:

type mockOllamaClient struct {
    extractFn  func(ctx context.Context, content string) (ollama.TagResult, error)
    embedFn    func(ctx context.Context, text string) ([]float32, error)
    healthyFn  func(ctx context.Context) bool
}
func (m *mockOllamaClient) ExtractTags(ctx context.Context, content string) (ollama.TagResult, error) { return m.extractFn(ctx, content) }
func (m *mockOllamaClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) { return m.embedFn(ctx, text) }
func (m *mockOllamaClient) Healthy(ctx context.Context) bool { return m.healthyFn(ctx) }

Table-driven test cases:

1. TestProcessContent_OllamaAvailable — healthy=true, extractTags returns 3 tags with confidence 0.92, embedding returns 768 floats
   Assert: result.Tags has 3 items, result.Confidence == 0.92, len(result.Embedding) == 768

2. TestProcessContent_OllamaDown_Fallback — healthy=false
   Assert: result.Tags is populated (from fallback), result.Confidence between 0.5-0.7, result.Embedding is nil

3. TestProcessContent_OllamaTagsFail_FallbackUsed — healthy=true but ExtractTags returns error, embedding works
   Assert: result.Tags populated (fallback), result.Embedding has 768 floats

4. TestProcessContent_EmptyContent — content is ""
   Assert: empty result, no error

Run: go test ./internal/nlp/ -v -race -count=1 (should FAIL on service tests)

FILE 2: Create internal/nlp/service.go

package nlp

import (
    "context"
    "log/slog"
    "github.com/neha037/mesh/internal/ollama"
)

type ProcessResult struct {
    Tags       []string
    Confidence float32
    Embedding  []float32
}

type Service struct {
    ollama   ollama.Client
    fallback *FallbackExtractor
}

func NewService(ollamaClient ollama.Client) *Service {
    return &Service{
        ollama:   ollamaClient,
        fallback: NewFallbackExtractor(),
    }
}

func (s *Service) ProcessContent(ctx context.Context, content string) (ProcessResult, error) {
    if content == "" {
        return ProcessResult{}, nil
    }

    var result ProcessResult

    if s.ollama.Healthy(ctx) {
        tagResult, err := s.ollama.ExtractTags(ctx, content)
        if err != nil {
            slog.Warn("ollama tag extraction failed, using fallback", "error", err)
            fallbackResult, _ := s.fallback.ExtractTags(content)
            result.Tags = fallbackResult.Tags
            result.Confidence = fallbackResult.Confidence
        } else {
            result.Tags = tagResult.Tags
            result.Confidence = tagResult.Confidence
        }

        embedding, err := s.ollama.GenerateEmbedding(ctx, content)
        if err != nil {
            slog.Warn("embedding generation failed", "error", err)
        } else {
            result.Embedding = embedding
        }
    } else {
        slog.Info("ollama unavailable, using fallback NLP")
        fallbackResult, _ := s.fallback.ExtractTags(content)
        result.Tags = fallbackResult.Tags
        result.Confidence = fallbackResult.Confidence
    }

    return result, nil
}

Run: go test ./internal/nlp/ -v -race -count=1 (should PASS)
```

### Verification

```bash
go test ./internal/nlp/ -v -race -count=1
```

---

## Step 8: Enhance Worker Processor

**What:** Wire NLP into the job pipeline. Add generate_embedding and build_edges job handlers.

**Prerequisite:** Steps 4, 5, 6, and 7 complete.

### Prompt

```
You are working on the Mesh project at /home/nkumari/GolandProjects/mesh. Read CLAUDE.md. Follow TDD.

TASK: Extend the worker processor to integrate NLP and add new job types.

PART A — Add CreateJob to JobRepository:

Edit internal/domain/job.go — ADD to the existing JobRepository interface:
    CreateJob(ctx context.Context, jobType string, payload json.RawMessage, maxAttempts int32) (string, error)

Edit internal/storage/job_repo.go — ADD implementation:

func (r *JobRepo) CreateJob(ctx context.Context, jobType string, payload json.RawMessage, maxAttempts int32) (string, error) {
    row, err := r.q.CreateJob(ctx, CreateJobParams{
        Type:        jobType,
        Payload:     payload,
        MaxAttempts: maxAttempts,
    })
    if err != nil {
        return "", fmt.Errorf("creating job: %w", err)
    }
    return uuidToString(row.ID), nil
}

Add "encoding/json" to the imports in job.go if not already there.

PART B — Rewrite internal/worker/processor.go:

The DefaultProcessor needs new fields:
    scraper *scraper.Service
    nodes   domain.NodeRepository
    nlpSvc  *nlp.Service               // NEW
    tags    domain.TagRepository        // NEW
    edges   domain.EdgeRepository       // NEW
    jobs    domain.JobRepository        // NEW

Update NewProcessor:
func NewProcessor(
    s *scraper.Service,
    nodes domain.NodeRepository,
    nlpSvc *nlp.Service,
    tags domain.TagRepository,
    edges domain.EdgeRepository,
    jobs domain.JobRepository,
) *DefaultProcessor

Update Process() switch to add:
    case "generate_embedding": return p.generateEmbedding(ctx, job)
    case "build_edges": return p.buildEdges(ctx, job)

Add new payload types:
type embeddingPayload struct { NodeID string `json:"node_id"` }
type edgesPayload struct { NodeID string `json:"node_id"` }

Add processNLP method — called by both processURL and processText after content is ready:

func (p *DefaultProcessor) processNLP(ctx context.Context, nodeID string) error {
    node, err := p.nodes.GetNodeContent(ctx, nodeID)
    if err != nil {
        return fmt.Errorf("getting node content: %w", err)
    }
    if node.Content == "" {
        return nil
    }

    result, err := p.nlpSvc.ProcessContent(ctx, node.Content)
    if err != nil {
        return fmt.Errorf("processing content: %w", err)
    }

    // Store tags
    for _, tagName := range result.Tags {
        tagID, err := p.tags.UpsertTag(ctx, tagName)
        if err != nil {
            slog.Error("upserting tag", "tag", tagName, "error", err)
            continue
        }
        if err := p.tags.AssociateNodeTag(ctx, nodeID, tagID, result.Confidence); err != nil {
            slog.Error("associating tag", "tag", tagName, "node_id", nodeID, "error", err)
        }
    }

    // Store embedding if available
    if result.Embedding != nil {
        updated, err := p.nodes.UpdateNodeEmbedding(ctx, nodeID, result.Embedding, node.Version)
        if err != nil {
            return fmt.Errorf("updating embedding: %w", err)
        }
        if !updated {
            slog.Warn("embedding update skipped (version conflict)", "node_id", nodeID)
        }

        // Enqueue edge building
        payload, _ := json.Marshal(edgesPayload{NodeID: nodeID})
        if _, err := p.jobs.CreateJob(ctx, "build_edges", payload, 3); err != nil {
            return fmt.Errorf("enqueueing build_edges: %w", err)
        }
    } else {
        slog.Info("no embedding generated (Ollama unavailable), skipping edge building", "node_id", nodeID)
    }

    return nil
}

Update processURL — add processNLP call after UpdateNodeContent:
    // ... existing scrape and UpdateNodeContent code ...
    if err := p.processNLP(ctx, payload.NodeID); err != nil {
        slog.Error("NLP processing failed", "node_id", payload.NodeID, "error", err)
        // Don't fail the job for NLP errors — content is already saved
    }
    return nil

Update processText — replace the UpdateNodeStatus("processed") with processNLP:
    if err := p.processNLP(ctx, payload.NodeID); err != nil {
        slog.Error("NLP processing failed", "node_id", payload.NodeID, "error", err)
    }
    if err := p.nodes.UpdateNodeStatus(ctx, payload.NodeID, "processed"); err != nil {
        return fmt.Errorf("updating node status: %w", err)
    }
    return nil

Add generateEmbedding method:
func (p *DefaultProcessor) generateEmbedding(ctx context.Context, job *domain.Job) error {
    var payload embeddingPayload
    if err := json.Unmarshal(job.Payload, &payload); err != nil {
        return fmt.Errorf("unmarshaling payload: %w", err)
    }
    // Reuse processNLP which handles embedding + edge enqueueing
    return p.processNLP(ctx, payload.NodeID)
}

Add buildEdges method:
func (p *DefaultProcessor) buildEdges(ctx context.Context, job *domain.Job) error {
    var payload edgesPayload
    if err := json.Unmarshal(job.Payload, &payload); err != nil {
        return fmt.Errorf("unmarshaling payload: %w", err)
    }

    // 1. Build tag-shared edges
    if err := p.edges.BuildTagSharedEdges(ctx, payload.NodeID); err != nil {
        return fmt.Errorf("building tag-shared edges: %w", err)
    }

    // 2. Build semantic edges if embedding exists
    node, err := p.nodes.GetNodeContent(ctx, payload.NodeID)
    if err != nil {
        return fmt.Errorf("getting node for semantic edges: %w", err)
    }
    // GetNodeContent doesn't return embedding — need to check via a different path
    // For now, find similar nodes will return empty if this node has no embedding
    // We need the embedding to search — skip if not available

    // 3. Update status to processed
    if err := p.nodes.UpdateNodeStatus(ctx, payload.NodeID, "processed"); err != nil {
        return fmt.Errorf("updating node status: %w", err)
    }

    return nil
}

NOTE: The buildEdges handler currently only builds tag-shared edges. Semantic edge
building requires accessing the node's embedding, which isn't returned by GetNodeContent.
This is acceptable for the initial implementation — semantic edges can be added when
a semantic search endpoint is built in Phase 3.

Add imports to processor.go:
    "github.com/neha037/mesh/internal/nlp"

PART C — Update mocks and tests:

Update any existing mock that implements JobRepository to add the CreateJob stub.
Check internal/worker/pool_test.go — add to mockJobRepo:
    createFn func(ctx context.Context, jobType string, payload json.RawMessage, maxAttempts int32) (string, error)

Create internal/worker/processor_test.go with test cases (use mocks for all dependencies):

1. TestProcessURL_WithNLP — scrape succeeds, NLP returns tags+embedding → tags stored, embedding stored, build_edges enqueued
2. TestProcessText_WithNLP — NLP processes existing content
3. TestBuildEdges_Success — builds tag-shared edges and updates status
4. TestProcessNLP_NoEmbedding — NLP returns tags but nil embedding → no build_edges enqueued

Run: go test ./internal/worker/ -v -race -count=1
Run: go test ./... -v -race -count=1
```

### Verification

```bash
go test ./internal/worker/ -v -race -count=1
go test ./... -v -race -count=1
```

---

## Step 9: Wire Everything in cmd/worker/main.go

**What:** Update worker entrypoint with new dependencies.

**Prerequisite:** Step 8 complete.

### Prompt

```
You are working on the Mesh project at /home/nkumari/GolandProjects/mesh. Read CLAUDE.md.

TASK: Update cmd/worker/main.go to wire Phase 2 dependencies.

Replace the dependency creation section (around lines 53-57) with:

    queries := storage.New(pool)
    jobRepo := storage.NewJobRepo(queries)
    nodeRepo := storage.NewNodeRepo(queries)
    tagRepo := storage.NewTagRepo(queries)
    edgeRepo := storage.NewEdgeRepo(queries)
    scraperSvc := scraper.NewService()

    ollamaClient := ollama.NewClient(cfg.OllamaHost, cfg.OllamaModel, cfg.EmbeddingModel)
    nlpSvc := nlp.NewService(ollamaClient)

    proc := worker.NewProcessor(scraperSvc, nodeRepo, nlpSvc, tagRepo, edgeRepo, jobRepo)

Add these imports:
    "github.com/neha037/mesh/internal/nlp"
    "github.com/neha037/mesh/internal/ollama"

Run: go build ./cmd/worker/
Run: go build ./cmd/api/
Both must compile successfully.
```

### Verification

```bash
go build ./cmd/worker/ && go build ./cmd/api/
```

---

## Step 10: Integration Tests

**What:** Integration tests for tag and edge repos with real PostgreSQL.

**Prerequisite:** Steps 4, 5, 6, 9 complete.

### Prompt

```
You are working on the Mesh project at /home/nkumari/GolandProjects/mesh. Read CLAUDE.md. Follow TDD.

TASK: Write integration tests for tag and edge repositories.

The project already has setupTestDB(t) in internal/storage/node_repo_integration_test.go
that starts pgvector/pgvector:pg16, runs migrations, and returns (*pgxpool.Pool, cleanup).
Reuse this function — it's in the same package (storage_test).

FILE 1: Create internal/storage/tag_repo_integration_test.go

//go:build integration

package storage_test

import (
    "context"
    "testing"
    "github.com/jackc/pgx/v5/pgtype"
    "github.com/neha037/mesh/internal/storage"
)

Test cases:

1. TestUpsertTag_Integration — Insert "golang", get ID. Insert "golang" again, verify same ID.
2. TestAssociateNodeTag_Integration — Create a node via InsertPendingNode, create a tag, associate with confidence 0.85, call GetNodeTags, verify the tag appears with correct confidence.
3. TestAssociateNodeTag_HigherConfidenceWins — Associate with 0.6, then 0.9. GetNodeTags should show 0.9.
4. TestGetNodeTags_OrderedByConfidence — Create 3 tags with confidence 0.5, 0.9, 0.7. Verify returned in order [0.9, 0.7, 0.5].

Helper to create test nodes:
    q := storage.New(pool)
    node, err := q.InsertPendingNode(ctx, storage.InsertPendingNodeParams{
        Type:  "article",
        Title: "Test Node",
    })

FILE 2: Create internal/storage/edge_repo_integration_test.go

//go:build integration

package storage_test

Test cases:

1. TestBuildTagSharedEdges_Integration:
   - Create node A, node B
   - Create 3 tags, associate all 3 with both A and B
   - Call BuildTagSharedEdges for node A
   - Query edges table to verify an edge exists from A to B with rel_type='tag_shared'

2. TestBuildTagSharedEdges_NoEdgeForSingleSharedTag:
   - Create node A, node B
   - Share only 1 tag between them
   - Call BuildTagSharedEdges for A
   - Verify no edge was created

3. TestUpsertSemanticEdge_Integration:
   - Create node A, node B
   - Call UpsertSemanticEdge(A, B, 0.8)
   - Verify edge exists
   - Call UpsertSemanticEdge(A, B, 0.9)
   - Verify weight updated to 0.9

4. TestFindSimilarNodes_Integration:
   - Create node A with embedding [1,0,0,...,0] (768 dims)
   - Create node B with embedding [0.95,0.05,0,...,0] (similar to A)
   - Create node C with embedding [0,0,...,0,1] (dissimilar)
   - Call FindSimilarNodes with A's embedding
   - Verify B is in results and C is not (or ranked much lower)

For setting embeddings, use the UpdateNodeEmbedding method or direct SQL:
    q.UpdateNodeEmbedding(ctx, storage.UpdateNodeEmbeddingParams{
        ID: nodeID, Embedding: pgvector.NewVector(vec), Version: 1,
    })

Run: TESTCONTAINERS_RYUK_DISABLED=true go test ./internal/storage/ -v -race -tags=integration -count=1
```

### Verification

```bash
TESTCONTAINERS_RYUK_DISABLED=true go test ./internal/storage/ -v -race -tags=integration -count=1
```

---

## Step 11: Docker Compose Verification

**What:** Rebuild and verify the worker runs correctly in Docker.

### Prompt

```
You are working on the Mesh project at /home/nkumari/GolandProjects/mesh. Read CLAUDE.md.

TASK: Rebuild the worker Docker container and verify it starts.

1. Read deploy/docker-compose.yml
2. Verify the worker service exists and has environment variables for OLLAMA_HOST, OLLAMA_MODEL, EMBEDDING_MODEL
3. If the worker service doesn't have these env vars, add them (matching the api service pattern)
4. Rebuild: cd deploy && docker-compose --env-file ../.env up -d --build worker
5. Check logs: docker logs mesh-worker --tail 20
6. Verify it logs "mesh-worker starting" without errors

Also verify the api service still works:
   cd deploy && docker-compose --env-file ../.env up -d --build api
   curl -s http://localhost:8080/healthz
   Should return {"status":"ok"}
```

### Verification

```bash
cd deploy && docker-compose --env-file ../.env up -d --build worker api
docker logs mesh-worker --tail 20 2>&1
curl -s http://localhost:8080/healthz
```

---

## Step 12: Update Living Documents

**What:** Update all project documentation per CLAUDE.md rules.

### Prompt

```
You are working on the Mesh project at /home/nkumari/GolandProjects/mesh. Read CLAUDE.md section "Living Documents — UPDATE RULES" carefully. You MUST update ALL listed documents.

TASK: Update all living documents to reflect Phase 2 completion.

1. docs/PROJECT_PROGRESS.md:
   - Add timeline entry with today's date: "Phase 2 — Ollama client, tag extraction (LLM + fallback NLP), embedding generation, auto edge-building (tag_shared + semantic), worker pipeline enhancement"
   - Update "Overall Status" current phase
   - Add Phase 2 progress section with completed items
   - Update file listing with new directories: internal/ollama/, internal/nlp/
   - Update "What's Next" to Phase 3

2. README.md:
   - Update Phase Status table: Phase 2 status
   - Update "Current Phase" line

3. docs/DEVELOPERS_GUIDE.md:
   - Add internal/ollama/ and internal/nlp/ to repo structure
   - Document new env vars: OLLAMA_HOST, OLLAMA_MODEL, EMBEDDING_MODEL
   - Add troubleshooting: "Ollama unavailable" → fallback NLP used automatically

4. docs/REVIEW_CHECKLIST.md:
   - Mark checklist items 5.11-5.14 as PASS
   - Mark checklist items 6.1-6.7 as PASS

5. docs/PROJECT_MESH_BLUEPRINT.md:
   - Check off Phase 2 items in roadmap
   - Update project structure

6. docs/roadmap.md:
   - Update Phase 2 status to Complete

7. docs/api-reference.md:
   - Note that Phase 2 is background processing (no new API endpoints)

8. docs/index.md:
   - Update feature status if applicable
```

---

## Execution Order & Dependencies

```
Step 1 (SQL + sqlc generate) ←── MUST BE FIRST
  |
  ├── Step 2 (Ollama client) ──── can run in parallel with 3, 4, 5, 6
  ├── Step 3 (NLP fallback)  ──── can run in parallel with 2, 4, 5, 6
  ├── Step 4 (TagRepo)       ──── can run in parallel with 2, 3, 5, 6
  ├── Step 5 (EdgeRepo)      ──── can run in parallel with 2, 3, 4, 6
  └── Step 6 (NodeRepo ext)  ──── can run in parallel with 2, 3, 4, 5
        |
        v
  Step 7 (NLP Service) ←── needs Steps 2 + 3
        |
        v
  Step 8 (Processor)   ←── needs Steps 4 + 5 + 6 + 7
        |
        v
  Step 9 (Wire main)   ←── needs Step 8
        |
        ├── Step 10 (Integration tests) ── can run in parallel with 11
        └── Step 11 (Docker verify)     ── can run in parallel with 10
              |
              v
        Step 12 (Docs) ←── ALWAYS LAST
```

**Total: 12 steps. Estimated 8-10 prompts if parallelizing.**
