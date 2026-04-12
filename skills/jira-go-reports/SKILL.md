---
name: jira-go-reports
description: PM-facing reports for jira-go — per-assignee workload, days-in-status, sprint health. Supports JSON output for agent parsing.
---

# jira-go — Reports

## Full Active Workload Report

```bash
jira report
```

When no `--sprint` flag is given, output is **grouped by sprint** — each sprint appears as a section header with its own assignee breakdown. Named sprints sort alphabetically; unassigned tickets appear under **Backlog** at the end.

Within each sprint section, columns are: **ASSIGNEE**, **TOTAL** (active tickets), **AVG DAYS** (average days in current status), **MAX DAYS** (highest days-in-status for that person). Followed by an indented per-status breakdown.

## Filter by Sprint

Passing `--sprint` switches to a **flat assignee view** (no sprint grouping) filtered to that sprint only.

```bash
jira report --sprint active     # Current sprint only (flat assignee view)
jira report --sprint future     # Upcoming sprint
jira report --sprint closed     # Most recent closed sprint
jira report --sprint 42         # Specific sprint by ID
```

## Filter by Assignee

```bash
jira report --assignee user@example.com
```

## JSON Output (for agents)

```bash
jira report --format json
```

Output schema when no `--sprint` filter (grouped by sprint):
```json
[
  {
    "sprint": "Sprint 12",
    "total": 5,
    "assignees": [
      {
        "name": "Alice Smith",
        "email": "alice@example.com",
        "total": 3,
        "by_status": { "In Progress": 2, "In Review": 1 },
        "avg_age_days": 4,
        "oldest_days": 9
      }
    ]
  },
  {
    "sprint": "Backlog",
    "total": 2,
    "assignees": [...]
  }
]
```

Output schema when `--sprint` is specified (flat assignee list):
```json
[
  {
    "name": "Alice Smith",
    "email": "alice@example.com",
    "total": 5,
    "by_status": { "In Progress": 3, "In Review": 2 },
    "avg_age_days": 4,
    "oldest_days": 9
  }
]
```

**Note:** `avg_age_days` and `oldest_days` are based on days since last status update (proxy for time in current status), not creation date.

## Task List with Age Column

```bash
# See days in current status per ticket
jira task list --no-interactive --active --age

# JSON with full data for agent processing
jira task list --no-interactive --format json
```

## Increase Fetch Limit

```bash
# Default is 200 issues; increase for large projects
jira report --limit 500
```

## Typical PM Review Workflow

```bash
# 1. Get current sprint overview
jira report --sprint active

# 2. Drill into someone with high MAX DAYS
jira task list --no-interactive --assignee dev@example.com --active --age

# 3. View the oldest stuck ticket
jira task view PROJ-99

# 4. Add a follow-up nudge
jira task comment PROJ-99 "Checking in — any blockers on this?"

# 5. If unblocked, update status
jira task status PROJ-99 "In Review" --comment "Moved to review after standup"
```

## Agent Parsing Example

```bash
# Find all assignees with tickets older than 5 days
jira report --format json | jq '[.[] | select(.oldest_days > 5) | {name, oldest_days, total}]'
```
