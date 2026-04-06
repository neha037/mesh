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

## Subagent Usage

- Dispatch one agent per independent task with full context
- Model selection: exploration/search → haiku, coding/testing → sonnet, architecture → opus
- Each agent follows TDD and commits its own work
- Never dispatch agents for interdependent tasks — run those sequentially

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
