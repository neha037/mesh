---
name: debug-issue
model: sonnet
description: Systematically debug issues using graph-powered code navigation and 4-phase process
---

## Debug Issue

Use the knowledge graph and systematic process to trace and debug issues.

### Graph-Powered Investigation

1. Use `semantic_search_nodes` to find code related to the issue.
2. Use `query_graph` with `callers_of` and `callees_of` to trace call chains.
3. Use `get_flow` to see full execution paths through suspected areas.
4. Run `detect_changes` to check if recent changes caused the issue.
5. Use `get_impact_radius` on suspected files to see what else is affected.

### 4-Phase Debugging Process

NO fixes before root cause is found.

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

### Stop Signals — return to Phase 1 if you catch yourself:
- Proposing fixes before understanding the issue
- Making multiple changes at once
- Planning a "quick fix now, investigate later"

### Tips

- Check both callers and callees to understand the full context.
- Look at affected flows to find the entry point that triggers the bug.
- Recent changes are the most common source of new issues.
