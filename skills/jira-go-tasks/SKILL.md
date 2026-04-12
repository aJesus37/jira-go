---
name: jira-go-tasks
description: Full task CRUD for the jira-go CLI — create, view, edit, delete, transition status, add comments, list with filters. Includes creating tasks directly in Done state for retroactive logging.
---

# jira-go — Tasks

## Create a Task

```bash
# Basic
jira task create --summary "Task title"

# Full options
jira task create \
  --summary "Implement login flow" \
  --description "Add OAuth2 + session management" \
  --type Task \
  --assignee dev@example.com \
  --owners "dev@example.com,qa@example.com"

# Create directly in Done (retroactive logging)
jira task create --summary "Deployed hotfix v1.2.3" --status "Done"

# Create a subtask under an existing task
jira task create --type Sub-task --parent PROJ-123 --summary "Write unit tests"
```

**Note:** `--status` triggers a Jira workflow transition immediately after creation. If the status is not reachable from the initial state, the CLI will list available options.

## List Tasks

```bash
# All active tasks (plain text, no TUI)
jira task list --no-interactive --active

# With days-in-status column
jira task list --no-interactive --active --age

# JSON output (for agent parsing)
jira task list --no-interactive --format json

# Filter by assignee
jira task list --no-interactive --assignee user@example.com

# Backlog only (not in any sprint)
jira task list --no-interactive --backlog

# Specific status
jira task list --no-interactive --status "In Review"
```

## View / Edit / Delete

```bash
jira task view PROJ-123
jira task edit PROJ-123 --summary "Updated title" --assignee other@example.com
jira task delete PROJ-123
```

## Transition Status

```bash
# Simple transition
jira task status PROJ-123 "In Progress"

# Transition + comment in one step
jira task status PROJ-123 "Done" --comment "Deployed to production at 14:30 UTC"
```

Status names are case-insensitive. If the target status is not available, the error lists valid options.

## Add a Comment

```bash
jira task comment PROJ-123 "Blocked by PROJ-99 — waiting on API contract"
```

## Common Patterns

**Register completed work retroactively:**
```bash
jira task create \
  --summary "Migrated prod DB to v2 schema" \
  --description "Ran migration, verified row counts, updated runbook" \
  --status "Done" \
  --assignee engineer@example.com
```

**Bulk status update with context:**
```bash
jira task status PROJ-45 "In Review" --comment "PR #234 open"
jira task status PROJ-46 "In Review" --comment "PR #235 open"
```

**Find stuck tasks:**
```bash
jira task list --no-interactive --active --age --format json \
  | jq '[.[] | select(.days_in_status > 5)]'
```
