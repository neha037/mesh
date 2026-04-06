---
name: Refactor Safely
model: sonnet
description: Plan and execute safe refactoring using dependency analysis and structured plans
---

## Refactor Safely

Use the knowledge graph to plan and execute refactoring with confidence.

### Graph-Powered Refactoring

1. Use `refactor_tool` with mode="suggest" for community-driven refactoring suggestions.
2. Use `refactor_tool` with mode="dead_code" to find unreferenced code.
3. For renames, use `refactor_tool` with mode="rename" to preview all affected locations.
4. Use `apply_refactor_tool` with the refactor_id to apply renames.
5. After changes, run `detect_changes` to verify the refactoring impact.

### Safety Checks

- Always preview before applying (rename mode gives you an edit list).
- Check `get_impact_radius` before major refactors.
- Use `get_affected_flows` to ensure no critical paths are broken.
- Run `find_large_functions` to identify decomposition targets.

### Implementation Plans

When creating plans for multi-step work:

- Every task should be completable in 2-5 minutes
- Include exact file paths and complete code — no "add appropriate validation" or "TBD"
- Each task follows the TDD cycle: write test → verify failure → implement → verify pass → commit
- Self-review checklist before execution:
  1. Map each requirement to at least one task
  2. Scan for placeholder language and remove it
  3. Verify type/function names are consistent across tasks
  4. Confirm every instruction is exact and actionable
