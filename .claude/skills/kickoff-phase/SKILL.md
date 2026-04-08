---
name: kickoff-phase
model: opus
description: Plan a new phase with weekly sprints, dependency ordering, and implementation roadmap
---

## Phase Kickoff Planning

Generate a comprehensive implementation plan for the next phase of the Mesh project,
organized into weekly sprints sized for 6-8 hours/week.

---

### Step 1: Read Current State

1. Read `docs/PROJECT_PROGRESS.md` — find the line `**Current Phase:** Phase N`
2. Set TARGET_PHASE = N (the phase we are planning)
3. Read `docs/REVIEW_CHECKLIST.md` — extract all criteria for TARGET_PHASE
4. Read `docs/PROJECT_MESH_BLUEPRINT.md` — find the architectural design for TARGET_PHASE
   (API endpoints, SQL schemas, algorithms, data flows)

---

### Step 2: Analyze Existing Foundation

Use graph tools (prefer over file reads):

```
get_architecture_overview
list_communities(min_size=2)
query_graph(pattern="children_of", target="internal/storage")
query_graph(pattern="children_of", target="internal/api/handler")
query_graph(pattern="children_of", target="internal/worker")
```

Identify:
- What interfaces already exist that this phase will extend
- What SQL tables/queries already exist that this phase will use
- What test patterns are established (mocks, testcontainers, table-driven)
- What packages need new files vs extending existing ones

Fall back to Read/Grep/Glob if graph tools are unavailable.

---

### Step 3: Extract Phase Scope

From the Phase Completion Tracker in REVIEW_CHECKLIST.md, list every unchecked item:

```
### Phase N: [Name] — [Subtitle] (Weeks X-Y)

- [ ] Item 1
- [ ] Item 2
...
```

Count total items. This is the full scope.

---

### Step 4: Dependency Analysis

For each item, determine:
1. **What it depends on** (which other items must be done first)
2. **What depends on it** (which items are blocked until this is done)
3. **Technical layer** (migration, storage, domain, handler, worker, test)

Build a dependency graph and determine the critical path.

---

### Step 5: Group into Weekly Sprints

Rules:
- **3-5 items per sprint** (sized for 6-8 hours/week)
- **Dependencies flow forward** — no sprint depends on a later sprint
- **Each sprint is independently testable** — can run `make test` after each
- **Migrations go first** — schema changes in Week 1
- **Infrastructure before features** — repos before handlers
- **Tests accompany their feature** — not deferred to a later sprint

---

### Step 6: Generate the Plan

Output this exact structure:

```markdown
# Phase N Kickoff: [Phase Name] — [Subtitle]

**Date:** [today]
**Estimated Duration:** N weeks (6-8 hrs/week)
**Total Items:** X
**Sprints:** Y weeks

## Phase Goal

[2-3 sentences from the blueprint describing what this phase achieves]

## Architecture Decisions Needed

[List any decisions that need user input before implementation begins.
Examples: "Which search algorithm for hybrid mode?", "Should graph depth be configurable or fixed at 5?"]

## SQL Migrations Required

| Migration | Tables/Indexes Affected | Notes |
|-----------|------------------------|-------|
| 00N_name.up.sql | ... | ... |

## New API Endpoints

| Method | Path | Description | Sprint |
|--------|------|-------------|--------|
| GET | /api/v1/... | ... | Week N |

## New/Modified Packages

| Package | Files | Change Type | Sprint |
|---------|-------|-------------|--------|
| internal/... | ... | New/Extend | Week N |

## Sprint Breakdown

### Week 1: [Theme] — Foundation
**Items:**
1. [ ] Item from checklist
2. [ ] Item from checklist
3. [ ] Item from checklist

**What you'll build:**
[Brief description of what gets built this week]

**Tests to write:**
- [specific test descriptions]

**Prompt:**
> /plan
> Implement Week 1 of Phase N: [list items]. Follow TDD. Use existing [specific interfaces/patterns].

---

### Week 2: [Theme] — Core Features
[Same structure as Week 1]

---

### Week 3: [Theme] — Polish & Integration
[Same structure as Week 1]

---

## Testing Strategy

- **Unit tests:** [what gets unit tested, patterns to follow]
- **Integration tests:** [what needs testcontainers, new test helpers]
- **Manual verification:** [curl commands, Docker steps]

## Risks

| Risk | Mitigation |
|------|------------|
| ... | ... |

## Verification Checklist

After all sprints complete, these must pass:
- [ ] `make test` — all unit tests green with -race
- [ ] `make test-integration` — all integration tests green
- [ ] `make lint` — no new lint warnings
- [ ] `/review-codebase` — all Phase N criteria PASS
- [ ] `/update-docs` — all living documents updated
```

---

### Edge Cases

**If the phase has > 15 items:**
Split into 4+ weeks. No sprint should have more than 5 items.

**If architecture decisions are needed:**
List them prominently at the top. Use AskUserQuestion to resolve before generating sprints.

**If a phase depends on external tools not yet set up:**
Flag as a prerequisite in the Risk section (e.g., "MinIO bucket initialization required before image upload endpoints").

**If previous phase has incomplete items:**
Flag them as blockers. Recommend completing them before starting the new phase.
