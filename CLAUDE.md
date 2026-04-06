# Mesh

Local-first Personal Growth Engine. Go backend, React frontend, PostgreSQL+pgvector, Ollama, MinIO.

See `docs/PROJECT_MESH_BLUEPRINT.md` for the full architectural blueprint.

## Living Documents — UPDATE RULES

After completing any implementation work in a session, update the following documents before finishing:

1. **`docs/PROJECT_PROGRESS.md`**
   - Add entry to the Project Timeline table with today's date and milestone
   - Check off completed items in the Phase Progress section
   - Update the "Current State" file inventory if new files/directories were created
   - Update the "Overall Status" if the phase changed

2. **`README.md`**
   - Update the Phase Status table if any phase status changed
   - Update Quick Start instructions if setup steps changed
   - Update the "Current Phase" line

3. **`docs/DEVELOPERS_GUIDE.md`**
   - Update if new conventions, tools, env vars, or Makefile targets were added
   - Update the Repository Structure section if directory layout changed
   - Add new troubleshooting entries for issues encountered

4. **`docs/REVIEW_CHECKLIST.md`**
   - Mark criteria as PASS/PARTIAL after verifying them
   - Check off completed items in the Phase Completion Tracker

5. **`docs/PROJECT_MESH_BLUEPRINT.md`**
   - Check off completed items in the Phase roadmap (Section 6)
   - Update the Project Structure (Section 8) if new files/directories were created
   - Update Docker Compose, Dockerfile, or Makefile sections (Section 9) if they changed
   - Update the header Version and Status if a milestone was reached

6. **`docs/api-reference.md`**
   - Update when API endpoints are added, changed, or removed
   - Keep request/response examples accurate with current field names and types

7. **`docs/index.md`**
   - Update the Feature Status table when features ship (change "Coming Soon" to "Available")
   - Add new features to the table as they are implemented

8. **`docs/roadmap.md`**
   - Update phase status when a phase is completed or started
   - Check off completed items in the phase checklists

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

## Systematic Debugging

When investigating bugs, follow this 4-phase process. NO fixes before root cause is found.

**Phase 1 — Investigate:**
- Read complete error messages and stack traces
- Reproduce the issue reliably with documented steps
- Check `git log` for recent changes that could be related
- Trace data flow backward from the symptom to the source

**Phase 2 — Analyze:**
- Compare broken code against working code (similar endpoints, prior versions)
- List every difference; document all assumptions

**Phase 3 — Hypothesize and Test:**
- State theory explicitly: "X causes this because Y"
- Test with one isolated change at a time
- If 3 separate fix attempts fail, stop — question the architecture, not the symptoms

**Phase 4 — Fix:**
- Write a failing test that reproduces the bug
- Implement the root-cause fix (not a symptom patch)
- Verify the test passes and no other tests break

Stop signals — return to Phase 1 if you catch yourself:
- Proposing fixes before understanding the issue
- Making multiple changes at once
- Planning a "quick fix now, investigate later"

## Verification Before Completion

Before claiming any task is done, follow this gate:

1. **IDENTIFY** the command that proves the claim (`make test`, `make lint`, `curl` endpoint)
2. **RUN** it fresh (not from memory or prior output)
3. **READ** full output including exit codes and failure counts
4. **CONFIRM** the output supports the claim
5. **REPORT** with evidence, not assertions

Never use "should work", "probably fine", or "seems correct". Show the output.
Never commit or push without fresh test evidence.

## Implementation Plans

When creating plans for multi-step work:

- Every task should be completable in 2-5 minutes
- Include exact file paths and complete code — no "add appropriate validation" or "TBD"
- Each task follows the TDD cycle: write test → verify failure → implement → verify pass → commit
- Self-review checklist before execution:
  1. Map each requirement to at least one task
  2. Scan for placeholder language and remove it
  3. Verify type/function names are consistent across tasks
  4. Confirm every instruction is exact and actionable

## Subagent Usage

For parallelizable work (e.g., implementing independent endpoints, writing tests for separate packages):

- Dispatch one agent per independent task with full context (file paths, specs, constraints)
- Each agent follows TDD and commits its own work
- After each agent completes, review for: spec compliance, then code quality
- Never dispatch agents for interdependent tasks — run those sequentially
- Use the cheapest capable model: simple file changes → haiku, integration work → sonnet, architecture → opus

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
