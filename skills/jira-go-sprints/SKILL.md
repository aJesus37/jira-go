---
name: jira-go-sprints
description: Sprint lifecycle management for jira-go — create sprints with auto-calculated 2-week dates, start, complete, move issues, view board.
---

# jira-go — Sprints

## Create a Sprint (2-week default)

Calculate dates before running: **start = nearest upcoming Monday, end = start + 14 days**.

```bash
# Example: today is 2026-04-11 (Saturday)
# Next Monday: 2026-04-13, end: 2026-04-27
jira sprint create \
  --name "Sprint 12" \
  --goal "Ship auth overhaul and payment v2 foundation" \
  --start 2026-04-13 \
  --end 2026-04-27
```

**Date calculation rule:**
- Start: the nearest upcoming Monday (if today is Monday, use today)
- End: start + 14 days (inclusive)

## List Sprints

```bash
jira sprint list                      # All sprints
jira sprint list --state active       # Currently running
jira sprint list --state future       # Upcoming
jira sprint list --state closed       # Past sprints
```

## Start a Sprint

```bash
# Get the sprint ID from sprint list output
jira sprint start 42

# Or use keyword aliases
jira sprint start future    # Start the first future sprint
```

## Complete a Sprint

```bash
jira sprint complete 42

# With auto-start if sprint is in FUTURE state
jira sprint complete future
```

## Move Issues to a Sprint

```bash
# Move specific issues
jira sprint move 42 PROJ-1 PROJ-2 PROJ-3

# Bulk move: move all non-done issues from another sprint
jira sprint move 42 --from-sprint 41

# Using keyword aliases
jira sprint move future --from-sprint active
```

## Bulk Move Non-Done Issues Between Sprints

```bash
# Move all non-done issues from sprint 41 to sprint 42
jira sprint move 42 --from-sprint 41

# Move from active to future sprint
jira sprint move future --from-sprint active
```

Note: Only issues with statusCategory != "Done" are moved.

## Complete a Sprint and Move Incomplete Tickets

```bash
# Complete sprint 42, move non-done tickets to sprint 43
jira sprint complete 42 --move-to 43

# Complete active sprint, move non-done tickets to future sprint
jira sprint complete active --move-to future
```

Note: FUTURE sprints are automatically started then completed (Jira requires ACTIVE→CLOSED).

## View Sprint Board (Kanban)

```bash
jira sprint board              # Active sprint (interactive TUI)
jira sprint board 42           # Specific sprint (interactive TUI)
jira sprint board --no-interactive  # Plain text output
```

## List Issues in a Sprint

```bash
jira sprint issues 42                  # All issues in sprint 42
jira sprint issues active               # Issues in active sprint
jira sprint issues 42 --status "In Progress"   # Filter by status
jira sprint issues 42 --format json     # Machine-readable output
jira sprint issues 42 --limit 200       # Increase beyond default 100
```

## Sprint ID Keyword Aliases

Most sprint commands accept `active` or `future` instead of a numeric sprint ID:

```bash
jira sprint issues active
jira sprint issues future
jira sprint complete active --move-to future
jira sprint start future
jira sprint move future --from-sprint active
```

## Edit Sprint

```bash
jira sprint edit 42   # Interactive: name, goal, start/end dates
```

## Full Sprint Kick-off Workflow

```bash
# 1. Create sprint with 2-week window
jira sprint create --name "Sprint 12" --goal "..." --start 2026-04-13 --end 2026-04-27
# → Created sprint ID: 42

# 2. Move planned issues into the sprint
jira sprint move 42 PROJ-10 PROJ-11 PROJ-12 PROJ-13

# 3. Start the sprint
jira sprint start 42

# 4. View the board
jira sprint board 42
```

## Sprint Close Workflow

```bash
# 1. Check what's still open
jira task list --no-interactive --active --sprint 42

# 2. Move unfinished tickets to next sprint (if needed)
jira sprint move 43 --from-sprint 42

# 3. Complete the sprint (auto-handles FUTURE→CLOSED)
jira sprint complete 42 --move-to 43
```
