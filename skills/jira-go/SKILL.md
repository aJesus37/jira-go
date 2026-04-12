---
name: jira-go
description: Use when an agent needs to interact with Jira via the jira-go CLI — creating tasks, updating status, managing sprints, epics, or getting a PM overview. Entry point skill; routes to sub-skills for complex operations.
---

# jira-go CLI

## Prerequisites

Config at `~/.config/jira-go/config.yaml` with `default_project`, auth credentials, and `board_id`.
Binary: `jira` (available in PATH).
Always use `--no-interactive` in automation contexts to get plain text output.

## Route to Sub-Skills

| Intent | Sub-skill |
|---|---|
| Create, update, status, comment, list tasks | `jira-go-tasks` |
| Epics and task/subtask hierarchy | `jira-go-epics` |
| Sprint lifecycle, board, moving issues | `jira-go-sprints` |
| PM workload and age-in-status report | `jira-go-reports` |

## Most Common Commands

```bash
# List active tasks (plain text)
jira task list --no-interactive --active

# List as JSON (for agents)
jira task list --no-interactive --format json

# Create a task
jira task create --summary "Fix login bug" --assignee user@example.com

# View a task
jira task view PROJ-123

# Transition status
jira task status PROJ-123 "In Progress"

# Transition + comment in one step
jira task status PROJ-123 "Done" --comment "Deployed to prod"

# Add a comment
jira task comment PROJ-123 "Blocked waiting on PROJ-99"

# PM workload report
jira report --sprint active

# Sprint kanban board
jira sprint board
```

## Key Flags

| Flag | Applies to | Effect |
|---|---|---|
| `--no-interactive` | list commands | Plain text, no TUI |
| `--format json` | task list, report | Machine-readable JSON |
| `--project KEY` | most commands | Override default project |
| `--assignee email` | task list, report | Filter by assignee |
| `--active` | task list | Exclude Done tickets |
| `--age` | task list | Show days-in-status column |
| `--limit N` | task list, report | Max results (default 50/200) |
