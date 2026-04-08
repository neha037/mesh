---
name: review-codebase
model: sonnet
description: Review entire codebase against completed-phase criteria using graph-powered analysis
---

## Review Entire Codebase

Perform a comprehensive audit of the Mesh codebase against `docs/REVIEW_CHECKLIST.md`,
filtering criteria to only completed phases and using graph tools to minimize token usage.

### Token Optimization Strategy

Use a three-layer progressive depth approach:
- **Layer 1**: Graph inventory (3 MCP calls, ~500 tokens)
- **Layer 2**: Targeted queries (12 MCP calls, ~2000 tokens)
- **Layer 3**: Strategic file reads (5-8 reads, ~3000 tokens)

Total: ~9,700 tokens vs ~30,000 for naive file-reading approach.

---

## Step 1: Detect Completed Phases

1. Read `docs/PROJECT_PROGRESS.md` and find the line containing "Current Phase: Phase N"
2. Set MAX_PHASE = N - 1 (e.g., if current is Phase 3, MAX_PHASE = 2)
3. All criteria for phases 1 through MAX_PHASE will be checked
4. Criteria for phases > MAX_PHASE will be marked as "N/A - Phase X"

---

## Step 2: Filter Applicable Criteria

1. Read `docs/REVIEW_CHECKLIST.md`
2. For each criterion in sections 1-12, check the "Phase" column:
   - If phase number <= MAX_PHASE: INCLUDE in review
   - If phase contains "+" (e.g., "1+"): INCLUDE if base <= MAX_PHASE
   - If phase contains range (e.g., "1-4"): INCLUDE if start <= MAX_PHASE
   - If phase > MAX_PHASE: SKIP (mark as N/A)
3. Count total applicable criteria for the summary

---

## Step 3: Layer 1 - Graph Inventory

Run these three MCP calls **in parallel**:

```
list_graph_stats
get_architecture_overview
list_communities(min_size=2)
```

### Criteria Answerable from Layer 1:

- **1.1** Directory layout: Check architecture overview for cmd/, internal/, migrations/, web/, deploy/, scripts/
- **1.3** Go module: Verify from graph stats that Go files exist
- **1.5** cmd/ entrypoints: Check architecture for cmd/api, cmd/worker, cmd/discovery
- **1.6** Single-responsibility packages: Check for circular import warnings
- **2.7** No global state: Look for constructor patterns in communities
- **2.8** Interfaces at consumer: Check interface locations
- **11.1, 11.2** Docs exist: Verify from file counts

---

## Step 4: Layer 2 - Targeted Graph Queries

### Batch A: Test Coverage (run in parallel)
```
query_graph(pattern="tests_for", target="internal/api/handler/handler.go")
query_graph(pattern="tests_for", target="internal/worker/pool.go")
query_graph(pattern="tests_for", target="internal/storage/node_repo.go")
list_flows(kind="Test", limit=50)
```
**Answers**: 8.1, 8.2, 8.5, 8.8, 8.9

### Batch B: Worker & NLP Systems (run in parallel)
```
semantic_search_nodes(query="worker pool goroutine", kind="Function")
semantic_search_nodes(query="circuit breaker gobreaker", kind="Function")
semantic_search_nodes(query="ollama embedding tag extraction", kind="Function")
query_graph(pattern="callees_of", target="Processor.Process")
```
**Answers**: 5.1-5.4, 5.7-5.15, 6.1-6.7

### Batch C: API & Error Handling (run in parallel)
```
query_graph(pattern="children_of", target="internal/api/handler")
semantic_search_nodes(query="graceful shutdown signal", kind="Function")
query_graph(pattern="callees_of", target="main")
```
**Answers**: 2.4, 2.6, 3.1-3.2, 3.19-3.22

### Batch D: Code Quality
```
semantic_search_nodes(query="panic", kind="Function")
refactor_tool(mode="dead_code")
find_large_functions(min_lines=50)
```
**Answers**: 2.1, 2.3, 2.9

---

## Step 5: Layer 3 - Strategic File Reads

Only read specific files for criteria that graph tools cannot answer:

### Database Schema
- Read `migrations/001_initial_schema.up.sql` — Check criteria 4.3-4.12
- Read `migrations/001_initial_schema.down.sql` — Check criterion 4.2

### Docker Configuration
- Read `deploy/docker-compose.yml` — Check criteria 9.1-9.6, 9.10, 10.1

### Configuration Files
- Read `.gitignore` — Check criteria 1.4, 10.4
- Read `.env.example` — Check criteria 9.9, 11.4
- Read `sqlc.yaml` — Check criterion 4.13
- Run `ls` in root directory — Check criterion 1.2
- Grep `Makefile` for `-race` and list of targets — Check criteria 8.3, 11.5

---

## Step 6: Generate Review Report

Output a structured markdown report:

```markdown
# Mesh Codebase Review — Phase 1-[MAX_PHASE]

**Date:** [today]
**Completed Phases:** 1 through [MAX_PHASE]
**Criteria Checked:** X / 86
**Criteria Skipped (future phases):** Y

## Summary
- PASS: X
- PARTIAL: Y
- MISSING: Z
- NEEDS VERIFICATION: W

## Findings by Category

### 1. Project Structure (N applicable)
| # | Criterion | Status | Evidence |
|---|-----------|--------|----------|
| 1.1 | Directory layout | PASS/PARTIAL/MISSING | [Evidence from graph/files] |
[... continue for all categories ...]

## Action Items

High Priority:
1. [MISSING] X.Y — [what needs to be done]

Medium Priority:
2. [PARTIAL] X.Y — [what needs improvement]

Needs Manual Verification:
3. [VERIFY] 4.14 — Run `make docker-up && make migrate-up`
4. [VERIFY] 11.3 — Follow README quick start guide

## Criteria Deferred to Future Phases
- Phase 3: [list]
- Phase 4+: [list]
```

---

## Edge Cases

### If Graph Tools Are Unavailable
If `list_graph_stats` returns an error:
1. Log: "Code graph unavailable, falling back to file reads"
2. Suggest: `code-review-graph build --full`
3. Use Read/Grep/Glob tools instead

### Manual Verification Criteria
These require running commands, mark as NEEDS VERIFICATION:
- 4.14: Migrations apply cleanly — `make docker-up && make migrate-up`
- 4.15: Migrations rollback cleanly — `make migrate-down && make migrate-up`
- 11.3: Quick start works — Follow README steps

### Updating the Checklist
Do NOT update `docs/REVIEW_CHECKLIST.md` unless explicitly requested by the user.
