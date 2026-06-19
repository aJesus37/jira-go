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

# Create a task and link it to an epic (one step)
jira task create --epic EPIC-10 --summary "Implement Stripe integration"

# Create a subtask that also belongs to an epic
jira task create --type Sub-task --parent PROJ-45 --epic EPIC-10 --summary "Subtask in epic"

# Assign to a sprint during creation
jira task create --summary "Sprint task" --sprint 48190

# Rich description from markdown (auto-converted to Atlassian Document Format)
jira task create \
  --summary "API contract" \
  --description "### Context\n\n- Endpoint: POST /api/v2/payments\n- Auth: OAuth2" \
  --description-format markdown

# Description from a markdown file
jira task create --summary "My task" --description-file ./desc.md --description-format markdown

# Description from ADF JSON file
jira task create --summary "My task" --description-file ./desc.adf.json --description-format adf-file
```

**Notes:**
- `--status` triggers a Jira workflow transition immediately after creation.
- `--parent` creates a subtask relationship (use with `--type Sub-task`).
- `--epic` links the issue to an epic during creation (uses `epic_link_field` from config for company-managed projects, or parent field for team-managed).
- `--description-format markdown` converts markdown to ADF so formatting renders correctly in Jira.
- `--sprint` assigns the sprint during creation (requires `sprint_field` in project config).

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

# Filter by reporter or watcher
jira task list --no-interactive --reporter user@example.com
jira task list --no-interactive --watcher user@example.com

# Filter by owner (multi-owner custom field)
jira task list --no-interactive --owner user@example.com

# Backlog only (not in any sprint)
jira task list --no-interactive --backlog

# Specific status
jira task list --no-interactive --status "In Review"

# Cross-project — multiple projects at once
jira task list --no-interactive --project SWCSIRT,SEC --active

# Cross-project — all configured projects
jira task list --no-interactive --all-projects --active

# Custom JQL query (no project scoping, full power)
jira task list --no-interactive --jql "reporter = currentUser() ORDER BY updated DESC"

# With limit
jira task list --no-interactive --active --limit 100
```

## My Tickets

Find all tickets where you are involved (assignee, reporter, or watcher):

```bash
# All tickets related to you (assignee + reporter)
jira task mine --no-interactive

# Active only, JSON for parsing
jira task mine --no-interactive --active --format json

# Only tickets assigned to you
jira task mine --no-interactive --role assignee

# Include watcher role (slower on large instances)
jira task mine --no-interactive --role assignee,reporter,watcher

# Scope to specific projects
jira task mine --no-interactive --project SWCSIRT,SEC

# With days-in-status
jira task mine --no-interactive --active --age
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

**Create epic + tasks + link in one workflow:**
```bash
jira epic create --summary "Payment Gateway v2"                   # → EPIC-10
jira task create --epic EPIC-10 --summary "Design API contract"
jira task create --epic EPIC-10 --summary "Implement Stripe"
jira epic view EPIC-10                                             # verify
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

**Cross-project "all my active work":**
```bash
jira task mine --no-interactive --active --format json
```
