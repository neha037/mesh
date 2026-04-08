# Mesh

Local-first Personal Growth Engine. Go backend, React frontend, PostgreSQL+pgvector, Ollama, MinIO.

See `docs/PROJECT_MESH_BLUEPRINT.md` for the full architectural blueprint.

## Living Documents

After completing implementation work, run `/update-docs` to update all living documents.

## Code Conventions

- Go: wrap errors with context, use context.Context as first param, no panics in library code
- SQL: use golang-migrate for migrations, sqlc for codegen, parameterized queries only
- Testing: table-driven tests, integration tests with testcontainers-go, always run with -race
- Docker: all ports bind to 127.0.0.1, health checks on infrastructure services
- Config: all settings via environment variables, never hardcode credentials

## Error Handling Rules

- **Never discard errors with `_, _ :=`**. If an error can occur, either handle it
  or propagate it. If you consciously decide to ignore an error, use an explicit
  `//nolint:errcheck` comment explaining why.
- **Fallback paths must propagate errors too.** If the primary and fallback both
  fail, return an error — don't return a zero-value result that looks like success.
- **Log level must match severity:** use `slog.Error` for failures that lose data
  or skip work, `slog.Warn` for degraded-but-recoverable situations.
- **Config validation:** every config value that will cause a runtime failure if
  wrong/empty must be validated at startup — either fatal error or explicit warning.

## System Boundary Validation

Data crossing a system boundary (external API response, user input, file read,
database result) must be validated before use:

- **External API responses (Ollama, MinIO, etc.):** validate structure, types,
  ranges, and invariants. For numeric vectors: check for NaN, Inf, zero vectors,
  expected dimensions. For strings: normalize case/whitespace if the system
  requires consistency.
- **Data invariants:** if the system assumes a property (e.g., embeddings are
  normalized, tag names are lowercase), enforce it at the point of entry, not at
  every point of use.
- **Library calls that can panic:** wrap with a length/nil check before calling.
  Prefer returning errors over allowing panics (e.g., pgvector.NewVector).
- **SQL edge cases:** always handle NULL, zero-division, and empty-set scenarios
  in queries. Use COALESCE/NULLIF defensively.

## Resilience Patterns

When writing worker/pipeline/async code:

- **Transient vs fatal errors:** classify errors. Transient errors (network,
  timeout, service down) should be retried. Fatal errors (malformed data, invalid
  payload) should fail immediately. Use sentinel errors (e.g., `domain.ErrFatal`).
- **Retry with jitter:** always add random jitter (25% of base delay) to backoff
  timers to prevent thundering herd when multiple workers recover simultaneously.
- **Timeout boundaries:** every job/request must have a timeout context. Default
  to 5 minutes for background jobs, 30 seconds for HTTP calls.
- **Stale state cleanup:** if code sets a status to "processing" or "in-progress",
  there must be a corresponding cleanup mechanism for when the processor crashes.
  Add a startup sweep or periodic cleanup.
- **Skipped work recovery:** if work is skipped because a dependency is down
  (e.g., Ollama unavailable), the job must be retried, not silently completed.
  Return an error to trigger the retry mechanism.
- **Startup dependency checks:** when a service depends on another (Ollama, MinIO,
  PostgreSQL), check connectivity at startup and log clearly if unavailable.

## Pipeline Design Rules

When building multi-stage job pipelines (e.g., Scrape -> Tag -> Embed -> Edge):

- **Never recompute across stages.** If stage N produces a result that stage N+1
  needs, store it (in DB or job payload) — don't regenerate it. Each stage should
  read its inputs from storage, not re-derive them.
- **Trace the full data flow** before implementing a new stage. Read the prior
  stage's code to understand what data is already available.
- **Prefer batch operations** over loops with individual inserts. Use SQL batch
  inserts (unnest, multi-row VALUES) when associating multiple records.
- **Job payloads carry IDs, not data.** Pass node_id/job_id in payloads, read
  full data from the database in each stage. This ensures stages always work with
  the latest state.

## Database Migration Rules

When writing PostgreSQL migrations (`migrations/*.sql`):

- Index predicates (WHERE clauses) must use only IMMUTABLE functions — `now()`,
  `current_timestamp`, `current_date`, and `clock_timestamp()` are NOT immutable
  and will cause `CREATE INDEX` to fail. Use column comparisons instead.
- Do not use `now()` or any non-immutable function in CHECK constraints,
  generated columns, or index predicates. They are only safe in DEFAULT clauses.
- Always provide a matching down migration that reverses the up migration exactly.
- Adding NOT NULL columns: always include a DEFAULT value.
- Run `make lint-sql` before committing any new migration to catch anti-patterns.
- Run `make test-integration` to verify the round-trip (up/down/up) passes.

## sqlc Query Rules

- **Never use `SELECT *`** — always list columns explicitly in every query.
- **Never hand-edit generated files** (`internal/storage/*.sql.go`, `internal/storage/models.go`).
  Always edit `internal/storage/queries/*.sql` and run `make sqlc` to regenerate.
- When a migration **adds a column**: update every query in `queries/*.sql` that
  touches that table — either add the column to the SELECT list or add a comment
  explaining why it is excluded (e.g., `-- excluding: embedding (large, fetched separately)`).
- When a migration **removes or renames a column**: update queries and run `make sqlc`
  before compiling.
- After running `make sqlc`, verify the generated structs in `nodes.sql.go` match
  what `node_repo.go` expects — check field names, types, and scan order.
- Run `make check-sqlc` before committing any migration or query change.

## Test-Driven Development

All new code follows RED-GREEN-REFACTOR:

1. **RED** — Write one minimal failing test. Run it. Confirm it fails for the right reason.
2. **GREEN** — Write the simplest code that makes it pass. No extras.
3. **REFACTOR** — Clean up only after green. Keep tests passing throughout.

Rules:
- Never write production code without a failing test first
- If code exists before tests, delete it and restart with TDD
- Each test covers exactly one behavior with a descriptive name
- Prefer real dependencies over mocks; mock only at system boundaries (external APIs, Ollama)
- Integration tests use testcontainers-go with real PostgreSQL, not mocks
- Run `make test` after every green step to catch regressions

Anti-patterns to avoid:
- Testing mock behavior instead of real component behavior
- Test-only methods in production code (move helpers to `_test.go`)
- Mock setup exceeding test logic in length — use integration tests instead
- Partial mock structures that don't mirror real API responses

## Debugging

For systematic debugging, use `/debug-issue`. No fixes before root cause is found.

## Verification Before Completion

Before claiming any task is done, follow this gate:

1. **IDENTIFY** the command that proves the claim (`make test`, `make lint`, `curl` endpoint)
2. **RUN** it fresh (not from memory or prior output)
3. **READ** full output including exit codes and failure counts
4. **CONFIRM** the output supports the claim
5. **REPORT** with evidence, not assertions

Never use "should work", "probably fine", or "seems correct". Show the output.
Never commit or push without fresh test evidence.

## CI Completeness

When adding new infrastructure or deployment artifacts, verify CI covers them:

- New Dockerfile -> add `docker build` step to CI
- New SQL migration -> verify `make lint-sql` and `make test-integration` pass
- New binary -> verify `make build` compiles it
- New service dependency -> add startup health check in entrypoint

## Subagent Usage

- Dispatch one agent per independent task with full context
- Each agent follows TDD and commits its own work
- Never dispatch agents for interdependent tasks — run those sequentially
- Model defaults by task type:

  | Task | Model | Reason |
  |------|-------|--------|
  | Codebase exploration, file search, graph queries | **haiku** | Fast + cheap; MCP tools do the heavy lifting |
  | Code implementation, TDD, testing, code review | **sonnet** | Speed + quality balance for iterative Go work |
  | Phase kickoff planning, API/schema design, architecture | **opus** | Worth the cost for decisions expensive to reverse |

## Before Starting Work

1. Read `docs/PROJECT_PROGRESS.md` to understand current state
2. Check which phase we're in and what items remain
3. Continue from where the last session left off

<!-- code-review-graph MCP tools -->
## MCP Tools: code-review-graph

**IMPORTANT: This project has a knowledge graph. ALWAYS use the
code-review-graph MCP tools BEFORE using Grep/Glob/Read to explore
the codebase.** The graph is faster, cheaper (fewer tokens), and gives
you structural context (callers, dependents, test coverage) that file
scanning cannot.

### When to use graph tools FIRST

- **Exploring code**: `semantic_search_nodes` or `query_graph` instead of Grep
- **Understanding impact**: `get_impact_radius` instead of manually tracing imports
- **Code review**: `detect_changes` + `get_review_context` instead of reading entire files
- **Finding relationships**: `query_graph` with callers_of/callees_of/imports_of/tests_for
- **Architecture questions**: `get_architecture_overview` + `list_communities`

Fall back to Grep/Glob/Read **only** when the graph doesn't cover what you need.

### Key Tools

| Tool | Use when |
|------|----------|
| `detect_changes` | Reviewing code changes — gives risk-scored analysis |
| `get_review_context` | Need source snippets for review — token-efficient |
| `get_impact_radius` | Understanding blast radius of a change |
| `get_affected_flows` | Finding which execution paths are impacted |
| `query_graph` | Tracing callers, callees, imports, tests, dependencies |
| `semantic_search_nodes` | Finding functions/classes by name or keyword |
| `get_architecture_overview` | Understanding high-level codebase structure |
| `refactor_tool` | Planning renames, finding dead code |

### Workflow

1. The graph auto-updates on file changes (via hooks).
2. Use `detect_changes` for code review.
3. Use `get_affected_flows` to understand impact.
4. Use `query_graph` pattern="tests_for" to check coverage.

## Compact Instructions

When compacting, preserve:
- Code changes made and their file paths
- Test results (pass/fail counts, specific failures)
- Architectural decisions and their rationale
- Current task progress and next steps
Discard: file exploration results, full file contents, verbose command output.
