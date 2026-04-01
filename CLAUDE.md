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

## Code Conventions

- Go: wrap errors with context, use context.Context as first param, no panics in library code
- SQL: use golang-migrate for migrations, sqlc for codegen, parameterized queries only
- Testing: table-driven tests, integration tests with testcontainers-go, always run with -race
- Docker: all ports bind to 127.0.0.1, health checks on infrastructure services
- Config: all settings via environment variables, never hardcode credentials

## Before Starting Work

1. Read `docs/PROJECT_PROGRESS.md` to understand current state
2. Check which phase we're in and what items remain
3. Continue from where the last session left off
