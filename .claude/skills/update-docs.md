---
name: Update Docs
model: haiku
description: Update all living documents after completing implementation work
user_invocable: true
---

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
