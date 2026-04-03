# AGENTS.md - jira-go

## Project Overview

A comprehensive Go CLI for Jira Software that supports task/sprint/epic management, agile ceremonies (planning, retros, dailies), and multi-owner task assignment via custom fields. Built with Charmbracelet libraries for a rich TUI experience.

## Quick Start for Agents

```bash
# Build the project
task build

# Run tests
task test

# Run in development mode
task dev

# Install locally
task install
```

## Architecture

```
cmd/jira-go/          # Main entry point
internal/
├── commands/         # Cobra CLI commands
│   ├── root.go       # Root command with global flags
│   ├── init.go       # Interactive configuration setup
│   ├── task.go       # Task CRUD operations
│   ├── cache.go      # Cache management
│   ├── project.go    # Project switching/config
│   ├── sprint.go     # Sprint operations
│   ├── ceremony.go   # Agile ceremonies (planning, retro, daily)
│   ├── tui.go        # TUI launcher
│   └── version.go    # Version info
├── config/           # Configuration management
│   ├── config.go     # YAML + env var config with validation
│   └── config_test.go
├── api/              # Jira REST API client
│   ├── client.go     # HTTP client with Basic Auth
│   ├── client_test.go
│   ├── users.go      # Email → AccountID resolution
│   ├── users_test.go
│   └── issues.go     # Issue CRUD operations
├── cache/            # SQLite caching layer
│   ├── cache.go      # TTL-based caching
│   └── cache_test.go
├── models/           # Domain models
│   └── models.go     # Issue, User, Sprint structs
└── tui/              # Bubble Tea TUI components
    ├── tui.go        # TUI runner
    └── issue_list.go # Issue list view
```

## Key Technologies

- **CLI Framework:** Cobra
- **TUI Framework:** Charmbracelet (Bubble Tea, Lipgloss, Bubbles)
- **Configuration:** YAML with environment variable override
- **Caching:** SQLite with TTL support
- **HTTP Client:** Standard library with Basic Auth
- **Build Tool:** Taskfile (not Make)

## Configuration

### Config File Location
`~/.config/jira-go/config.yaml`

### Config Structure
```yaml
default_project: PROJ
auth:
  email: user@example.com
  api_token: token_here
projects:
  PROJ:
    jira_url: https://company.atlassian.net
    board_id: 1
    multi_owner_field: customfield_10001
cache:
  enabled: true
  default_ttl: 30m
  location: ~/.cache/jira-go/cache.db
```

### Environment Variables
- `JIRA_GO_CONFIG` - Path to config file
- `JIRA_GO_EMAIL` - Override email
- `JIRA_GO_API_TOKEN` - Override API token
- `JIRA_GO_DEFAULT_PROJECT` - Override default project
- `JIRA_GO_CACHE` - Override cache path

## Multi-Owner Support

Jira natively only supports a single assignee. Multi-owner support is implemented via a custom multi-user picker field:

1. Create custom field in Jira (Multi User Picker type)
2. Note the field ID (e.g., `customfield_10001`)
3. Configure in `jira-go init`
4. Use `--owners` flag with comma-separated emails

## Testing

### Running Tests
```bash
# All tests
task test

# Specific package
go test ./internal/config -v
go test ./internal/api -v
go test ./internal/cache -v
```

### Test Coverage
- Config loading and validation
- API client authentication
- User email resolution
- Cache operations with TTL

## Jira API Requirements

The CLI requires a Jira Cloud instance with:

1. **API Token** (not password)
   - Generated at: https://id.atlassian.com/manage-profile/security/api-tokens

2. **Permissions:**
   - Browse projects
   - Create issues
   - Edit issues
   - Delete issues (optional)
   - Search for issues (JQL)
   - View user details (for email resolution)

3. **Custom Field (for multi-owner):**
   - Type: Multi User Picker
   - Applied to relevant issue type screens

## Common Development Tasks

### Adding a New Command
1. Create file in `internal/commands/`
2. Define cobra command with `RunE`
3. Register in `init()` function
4. Add to root command via `rootCmd.AddCommand()`

### Adding API Methods
1. Add method to `internal/api/client.go` or appropriate file
2. Follow pattern: return `(*Model, error)` for gets, `error` for mutations
3. Handle HTTP status codes appropriately
4. Add test in corresponding `_test.go` file

### Working with the Cache
```go
import "github.com/user/jira-go/internal/cache"

// Get cache path from config
cachePath := config.GetCachePath()

// Create/open cache
c, err := cache.New(cachePath)
if err != nil {
    return err
}
defer c.Close()

// Store with TTL
c.Set("key", data, 30*time.Minute)

// Retrieve
data, err := c.Get("key")

// User mappings
c.SetUserMapping(email, accountID, 7*24*time.Hour)
accountID, err := c.GetUserMapping(email)
```

## Debugging

### Enable Verbose Output
```bash
jira-go -v task list
```

### Bypass Cache
```bash
jira-go --no-cache task list
```

### Custom Cache TTL
```bash
jira-go --cache-ttl=5m task list
```

### View Cache Contents
```bash
# SQLite CLI
sqlite3 ~/.cache/jira-go/cache.db

# Or use cache command
jira-go cache status
```

## Build & Release

### Build with Version Info
```bash
task build
# Binary: ./build/jira-go
```

### Version Injection
The Taskfile injects version info at build time:
- Version: Git tag or "dev"
- Commit: Git short SHA
- Build Date: ISO 8601 timestamp

## Known Limitations

1. **Sprint API:** Framework exists but full implementation pending
2. **Ceremony TUI:** Placeholder commands only, full TUI coming
3. **Epic Management:** Not yet implemented
4. **Glow Integration:** Markdown rendering planned but not implemented
5. **Webhook Support:** Real-time updates not implemented

## Error Handling Philosophy

- Wrap errors with context: `fmt.Errorf("doing thing: %w", err)`
- Return meaningful errors to users
- Use validation in `config.Validate()`
- Handle API errors with status code checks

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting (`task fmt`)
- Run linter before commits (`task lint`)
- Write tests for new functionality
- Keep functions focused and small
- Document public APIs with comments

## Git Workflow

```bash
# Check status
git status

# Make changes, then commit
git add -A
git commit -m "type: description"

# Commit types:
# - feat: new feature
# - fix: bug fix
# - docs: documentation
# - test: adding tests
# - chore: build/tooling changes
# - refactor: code restructuring
```

## Resources

- [Jira REST API v3 Docs](https://developer.atlassian.com/cloud/jira/platform/rest/v3/)
- [Cobra Docs](https://cobra.dev/)
- [Bubble Tea Docs](https://github.com/charmbracelet/bubbletea)
- [Taskfile Docs](https://taskfile.dev/)

## Troubleshooting

### Build Errors
```bash
# Clean and rebuild
task clean
task deps
task build
```

### SQLite Issues
```bash
# Clear cache if corrupted
jira-go cache clear
```

### Config Issues
```bash
# Check config location
jira-go init  # Re-run setup

# Or edit directly
cat ~/.config/jira-go/config.yaml
```

### API Connection Issues
- Verify Jira URL includes protocol (https://)
- Check API token hasn't expired
- Ensure email matches Atlassian account
- Verify project key exists and you have access
