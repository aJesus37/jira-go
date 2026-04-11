# jira-go Skill Suite — Design Doc

**Date:** 2026-04-11  
**Status:** Approved

## Overview

Design for a modular Claude skill suite that enables AI agents to interact with the jira-go CLI fluently. The skill suite ships alongside CLI extensions that close gaps in the current command set. Skill and CLI are designed together so the skill documents a complete, working surface with no caveats.

## Goals

- AI agents (autonomous and interactive) can create, update, and report on Jira work via the CLI
- Product managers can get a workload/status summary through a dedicated report command
- Sprint lifecycle (create, start, complete, 2-week default) is fully scriptable
- Epic → Task → Subtask hierarchy is supported at all levels

## Skill Architecture

Five files under `~/.claude/skills/`:

```
jira-go.md          # Router: overview, cheatsheet, config prereqs
jira-go:tasks.md    # Task CRUD, status transitions, comments, done-state creation
jira-go:epics.md    # Epic creation, task/subtask linking, hierarchy navigation
jira-go:sprints.md  # Sprint lifecycle, date auto-calc, board view
jira-go:reports.md  # PM view: per-person workload, age-in-status
```

### Invocation pattern
Agents invoke `jira-go` first. It describes which sub-skill to load for each intent and includes a compact cheatsheet covering the most common operations (so short tasks don't need a sub-skill at all).

## CLI Extensions Required

The following commands/flags must be implemented before the skill is written:

| Command | Flags | Purpose |
|---|---|---|
| `jira task status [key] [status]` | `--comment "msg"` | Transition issue status, optionally add comment |
| `jira task comment [key] "message"` | — | Add a comment to an issue |
| `jira task create` | `--status "Done"` | Create directly in a target status (for retroactive logging) |
| `jira task list` | `--format json` | Machine-readable output for agent parsing |
| `jira task list` | `--age` column | Days in current status shown in output |
| `jira report` | `--assignee`, `--sprint`, `--format json` | PM summary: tasks per person, age-in-status |

## Sub-skill Content Outline

### `jira-go` (router + cheatsheet)
- Decision tree: which sub-skill to invoke for which intent
- Top ~10 most common commands with real examples
- Config prerequisites: `default_project`, `multi_owner_field`, board ID
- Reminder: always use `--no-interactive` in agent/automation contexts

### `jira-go:tasks`
- Create with all flags: `--summary`, `--description`, `--assignee`, `--owners`, `--type`, `--status`
- Creating in Done state: `jira task create --summary "..." --status "Done"`
- Transition status: `jira task status PROJ-123 "In Progress" --comment "starting now"`
- Add comment: `jira task comment PROJ-123 "reviewed and approved"`
- List with JSON: `jira task list --no-interactive --format json`
- View / edit / delete patterns

### `jira-go:epics`
- Create epic: `jira epic create --summary "Epic name"`
- Add existing tasks: `jira epic add PROJ-1 PROJ-2 PROJ-3`
- Create subtask: `jira task create --type Sub-task --parent PROJ-123 --summary "..."`
- View hierarchy: `jira epic view PROJ-1`
- List all epics: `jira epic list`

### `jira-go:sprints`
- Create with 2-week auto-dates: start = next Monday, end = start + 14 days
- Explicit override: `jira sprint create --start 2026-04-14 --end 2026-04-28`
- Start sprint: `jira sprint start [id]`
- Complete sprint: `jira sprint complete [id]`
- Move issues: `jira sprint move [sprint-id] PROJ-1 PROJ-2`
- Board view: `jira sprint board`

### `jira-go:reports`
- Full PM summary: `jira report`
- Filter by sprint: `jira report --sprint active`
- Filter by assignee: `jira report --assignee user@example.com`
- JSON for agent parsing: `jira report --format json`
- Output includes: ticket count per person, average days-in-status, oldest tickets

## Sequence

1. Implement CLI extensions (task status, task comment, --status flag on create, --format json on list, --age column, jira report command)
2. Write the five skill files against the completed CLI
3. Test skill invocations with real agent prompts

## Non-Goals

- No webhook or real-time update support
- No Glow/markdown rendering integration (separate future initiative)
- No ceremony TUI (separate future initiative)
