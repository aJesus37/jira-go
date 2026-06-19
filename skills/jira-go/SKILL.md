---
name: jira-go
description: Use when an agent needs to interact with Jira via the jira-go CLI — creating tasks, updating status, managing sprints, epics, or getting a PM overview. Entry point skill; routes to sub-skills for complex operations.
---

# jira-go CLI

## Prerequisites

Config at `~/.config/jira-go/config.yaml` with `default_project`, auth credentials, and project configs.
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

# All tickets related to you across all projects
jira task mine --no-interactive

# Active tickets assigned to you across all projects
jira task mine --no-interactive --active --role assignee

# Cross-project listing
jira task list --no-interactive --all-projects --active --format json

# Custom JQL (full power, reuses CLI auth/formatting)
jira task list --no-interactive --jql "assignee = currentUser() AND statusCategory != Done ORDER BY updated DESC"

# PM workload report
jira report --sprint active

# Sprint kanban board
jira sprint board
```

## Key Flags

| Flag | Applies to | Effect |
|---|---|---|
| `--no-interactive` | list commands | Plain text, no TUI |
| `--format json` | task list, report, mine | Machine-readable JSON |
| `--project KEY(S)` | most commands | Override or comma-separated cross-project |
| `--assignee email` | task list, report | Filter by assignee |
| `--reporter email` | task list | Filter by reporter |
| `--watcher email` | task list | Filter by watcher |
| `--active` | task list, mine | Exclude Done tickets |
| `--age` | task list, mine | Show days-in-status column |
| `--limit N` | task list, report, mine | Max results (default 50/200) |
| `--all-projects` | task list | Search across all configured projects |
| `--jql` | task list | Custom JQL query (overrides other filters) |
| `--role` | task mine | Roles: assignee, reporter, watcher (comma-separated) |
| `--parent KEY` | task create | Parent for subtask creation |
| `--epic KEY` | task create | Link to epic during creation |
| `--description-format` | task create | plain, markdown, or adf-file |
| `--description-file` | task create | Path to description file (markdown or ADF JSON) |
| `--sprint ID` | task create | Sprint ID to assign to |

## Multi-Project Setup

Projects are configured in `~/.config/jira-go/config.yaml`. Only `jira_url` is required:

```yaml
projects:
    SWCSIRT:
        jira_url: https://ifood.atlassian.net/
        board_id: 16502
        multi_owner_field: "customfield_26030"
        sprint_field: "customfield_10007"
        epic_link_field: "customfield_10014"  # optional, for company-managed projects
    SEC:
        jira_url: https://ifood.atlassian.net/
    SWSENTI:
        jira_url: https://ifood.atlassian.net/
```

- `board_id` is optional — only needed for `sprint` commands
- `multi_owner_field` / `sprint_field` / `epic_link_field` are optional
- `epic_link_field` enables epic linking for company-managed projects; omit for team-managed
- Use `jira project switch KEY` to change the default project
- Use `--project KEY` on any command to override temporarily
- Use `--project KEY1,KEY2` for cross-project queries

## Cross-Project Workflows

When a user asks for "all my tickets" or "everything I'm involved in", use:

```bash
jira task mine                       # assignee + reporter, all projects
jira task mine --active              # same, but only active tickets
jira task mine --role assignee,reporter,watcher  # include watcher (slower)
```

For custom cross-project queries, use `--jql`:

```bash
jira task list --jql "assignee = currentUser() AND project IN (SWCSIRT, SEC)" --format json
```
