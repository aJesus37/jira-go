# Jira CLI Design Document

**Date:** 2025-01-22
**Project:** jira-go
**Status:** Approved

## Overview

A comprehensive CLI tool for managing Jira Software projects, supporting the full lifecycle of tasks, sprints, and epics. Includes interactive ceremonies (planning, retrospectives, dailies) using rich TUI experiences powered by Charmbracelet libraries.

## Goals

- Provide a fast, intuitive CLI for daily Jira operations
- Support email-based user interactions (transparently resolved to Jira accountIds)
- Enable multi-owner task assignment via custom multi-select fields
- Facilitate agile ceremonies (planning, retros, dailies) through beautiful TUI
- Offer flexible caching for offline access and performance
- Deliver a single, portable binary

## Architecture

### High-Level Structure

```
┌─────────────────────────────────────────────────────────────┐
│                     CLI Entry (cobra)                        │
├─────────────────────────────────────────────────────────────┤
│  Commands Layer                                              │
│  ├── task (create, edit, delete, list, view)                │
│  ├── sprint (create, start, complete, list)                 │
│  ├── epic (create, edit, list)                              │
│  ├── project (switch, list, config)                         │
│  └── ceremony (planning, retro, daily)                      │
├─────────────────────────────────────────────────────────────┤
│  Core Services                                               │
│  ├── Jira API Client (wrapper with retries)                 │
│  ├── Cache Manager (SQLite with TTL)                        │
│  ├── Config Manager (YAML + env vars)                       │
│  └── User Resolver (email ↔ accountId)                      │
├─────────────────────────────────────────────────────────────┤
│  TUI Components (charmbracelet)                              │
│  ├── IssueList (filterable, sortable)                       │
│  ├── IssueDetail (with glow markdown renderer)              │
│  ├── SprintBoard (kanban view)                              │
│  ├── CeremonyViews (planning, retro, daily)                 │
│  └── Forms (create/edit with validation)                    │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack

- **Language:** Go 1.21+
- **CLI Framework:** Cobra
- **TUI Framework:** Charmbracelet (Bubble Tea, Lipgloss, Bubbles)
- **Markdown Rendering:** Glow
- **API Client:** Custom Jira REST API v3 wrapper
- **Cache:** SQLite with TTL support
- **Config:** YAML with environment variable override

## Core Features

### 1. Task Management

**Commands:**
- `jira-go task create` - Create new issues
- `jira-go task list` - List and filter issues
- `jira-go task view <key>` - View issue details with markdown rendering
- `jira-go task edit <key>` - Edit existing issues
- `jira-go task delete <key>` - Delete issues
- `jira-go task assign <key>` - Assign to user

**Features:**
- Email-based user references (auto-resolved to accountIds)
- Multi-owner support via custom field
- Rich markdown description viewing via Glow
- JQL query support for filtering

### 2. Sprint Management

**Commands:**
- `jira-go sprint list` - List sprints
- `jira-go sprint create` - Create new sprint
- `jira-go sprint start <id>` - Start sprint
- `jira-go sprint complete <id>` - Complete sprint
- `jira-go sprint board` - Interactive kanban board

### 3. Epic Management

**Commands:**
- `jira-go epic list` - List epics
- `jira-go epic create` - Create new epic
- `jira-go epic view <key>` - View epic with child issues
- `jira-go epic add <epic-key> <issue-key>` - Add issue to epic

### 4. Agile Ceremonies (TUI)

#### Sprint Planning
- Kanban-style backlog view
- Drag-and-drop issue assignment to sprint
- Story point estimation with team voting
- Timer for time-boxed discussions
- Export planning notes to Markdown

#### Retrospective
- Three-column board (Went Well, Improve, Action Items)
- Anonymous card submission
- Card voting and grouping
- Timer for each phase
- Export action items as Jira issues

#### Daily Standup
- Team member checklist
- Quick update entry per person
- Blocker highlighting with priority
- Timer to keep standups brief
- Export daily summary

### 5. Project Context

**Commands:**
- `jira-go init` - First-time setup wizard
- `jira-go project list` - List accessible projects
- `jira-go project switch` - Switch default project
- `jira-go project config` - View/edit project settings

## Multi-Owner Implementation

### Strategy

Jira natively supports only a single assignee per issue. To enable multiple owners:

1. **Custom Multi-User Picker Field**
   - Configure a custom field in Jira (e.g., `customfield_10001`)
   - Store as array of user accountIds
   - Visible and queryable in Jira web UI

2. **Configuration**
   ```yaml
   projects:
     PROJ:
       multi_owner_field: "customfield_10001"
   ```

3. **Email Resolution**
   - CLI accepts: `--owners "alice@example.com,bob@example.com"`
   - Service layer resolves emails → accountIds
   - Cached mapping for performance

### User Resolution Flow

```
User Input (email)
    ↓
Check Cache (SQLite)
    ↓
Cache Hit → Return accountId
Cache Miss → API Lookup → Update Cache → Return accountId
```

## Caching Strategy

### Cache Types

1. **API Response Cache**
   - Raw Jira API responses
   - Configurable TTL (default: 30 minutes)
   - Redis-like key structure

2. **User Mapping Cache**
   - Email → accountId mappings
   - Long TTL (default: 7 days)
   - Updated on lookup miss

3. **Ceremony State Cache**
   - Persistent SQLite storage
   - Session recovery
   - Export functionality

### Cache Control

```bash
# Global commands
jira-go cache clear              # Clear all caches
jira-go cache status             # Show statistics

# Per-command flags
jira-go task list --no-cache     # Bypass cache
jira-go task list --cache-ttl=5m # Custom TTL
```

### Configuration

```yaml
cache:
  enabled: true
  default_ttl: 30m
  max_size_mb: 100
  location: "~/.cache/jira-go/"
```

## Configuration

### File Location

- **Primary:** `~/.config/jira-go/config.yaml`
- **Override:** Environment variables with `JIRA_GO_` prefix

### Structure

```yaml
# ~/.config/jira-go/config.yaml
default_project: "PROJ"
auth:
  email: "user@example.com"
  api_token: "${JIRA_GO_API_TOKEN}"  # Env var reference

projects:
  PROJ:
    jira_url: "https://company.atlassian.net"
    board_id: 1
    multi_owner_field: "customfield_10001"
    issue_types:
      story: "10001"
      bug: "10002"
      task: "10003"
    
cache:
  enabled: true
  default_ttl: 30m
  location: "~/.cache/jira-go/"
```

### Environment Variables

```bash
JIRA_GO_API_TOKEN=your-token-here
JIRA_GO_DEFAULT_PROJECT=PROJ
JIRA_GO_CACHE_ENABLED=true
```

## User Experience Flow

### First-Time Setup

```bash
$ jira-go init
? Jira URL: https://company.atlassian.net
? Email: user@company.com
? API Token: [hidden]
? Default Project: PROJ
✓ Configuration saved to ~/.config/jira-go/config.yaml
✓ Verified connection to Jira
```

### Daily Usage

```bash
# Default project from config
jira-go task list
jira-go task create --title "Fix login bug" --type Bug --owners "dev1@company.com,dev2@company.com"

# Override for different project
jira-go --project OTHER task list

# Interactive project switch
jira-go project switch

# Rich markdown viewing
jira-go task view PROJ-123

# Run ceremonies
jira-go ceremony planning
jira-go ceremony retro
jira-go ceremony daily
```

## TUI Design (Charmbracelet)

### Components

1. **IssueList**
   - Table view with sortable columns
   - Filter input with real-time search
   - Status indicators with colors
   - Keyboard navigation (vim-style)

2. **IssueDetail**
   - Header with key metadata
   - Glow-rendered markdown description
   - Comments section
   - Related issues

3. **SprintBoard**
   - Kanban columns (To Do, In Progress, Done)
   - Issue cards with assignee avatars
   - Drag-and-drop support
   - Story point totals

4. **CeremonyViews**
   - Planning: Split view (backlog / sprint)
   - Retro: Three-column card board
   - Daily: Checklist with timers

### Styling

- Consistent color scheme via Lipgloss
- Status colors: TODO (gray), IN_PROGRESS (blue), DONE (green), BLOCKED (red)
- Priority indicators: Low, Medium, High, Critical
- Responsive layouts for different terminal sizes

## Error Handling

### Strategies

1. **Graceful Degradation**
   - Cache miss → API call → Cache update
   - API unavailable → Serve from cache (stale data warning)

2. **User-Friendly Errors**
   - Clear error messages without stack traces
   - Suggestions for resolution
   - Connection troubleshooting hints

3. **Validation**
   - Pre-flight checks for required fields
   - Email format validation
   - Project/field existence verification

### Exit Codes

- `0` - Success
- `1` - General error
- `2` - Invalid arguments
- `3` - Authentication failure
- `4` - API error (rate limit, not found)
- `5` - Configuration error

## Security Considerations

1. **API Token Storage**
   - Store in config file with 0600 permissions
   - Support environment variable override
   - Never log or display tokens

2. **Cache Security**
   - SQLite database with restricted permissions
   - No sensitive data in logs
   - Clear cache on logout

3. **Email Privacy**
   - Cache email mappings locally
   - No email transmission outside Jira API

## Future Enhancements

- Webhook integration for real-time updates
- Plugin system for custom commands
- Confluence integration for documentation
- Slack/Teams notification support
- Import/export functionality
- CI/CD integration commands

## Dependencies

### Production

```go
require (
    github.com/spf13/cobra v1.8.0
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.0
    github.com/charmbracelet/bubbles v0.18.0
    github.com/charmbracelet/glow v1.5.0
    github.com/mattn/go-sqlite3 v1.14.0
    gopkg.in/yaml.v3 v3.0.1
)
```

### Development

```go
require (
    github.com/stretchr/testify v1.8.4
    github.com/golang/mock v1.6.0
)
```

## Appendix

### A. Jira API Endpoints Used

- `GET /rest/api/3/search` - Issue search (JQL)
- `POST /rest/api/3/issue` - Create issue
- `PUT /rest/api/3/issue/{id}` - Update issue
- `DELETE /rest/api/3/issue/{id}` - Delete issue
- `GET /rest/api/3/user/search` - User lookup by email
- `GET /rest/api/3/project` - List projects
- `GET /rest/agile/1.0/board/{boardId}/sprint` - Sprint operations
- `GET /rest/agile/1.0/sprint/{sprintId}/issue` - Sprint issues

### B. Multi-Owner Field Configuration

1. Go to Jira Project Settings → Issue Types
2. Find/Create custom field: "Additional Owners" (Multi User Picker)
3. Note the custom field ID (e.g., `customfield_10001`)
4. Add to relevant issue type screens
5. Configure in CLI: `multi_owner_field: "customfield_10001"`

---

**Next Step:** Create implementation plan using `writing-plans` skill.
