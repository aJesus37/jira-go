# jira

A comprehensive CLI tool for managing Jira Software projects with support for task/sprint/epic management, agile ceremonies, and a beautiful TUI powered by Charmbracelet.

## Features

- 🎯 **Task Management**: Create, edit, delete, and list Jira issues
- 🏃 **Sprint Operations**: Manage sprints and view sprint boards
- 📊 **Epic Management**: Organize work with epics
- 👥 **Multi-Owner Support**: Assign multiple owners to tasks via custom fields
- 📧 **Email-Based Interaction**: Use emails instead of account IDs
- 🖥️ **Rich TUI**: Beautiful terminal interface using Bubble Tea
- ⚡ **Local Caching**: SQLite-based cache for offline access
- 🎪 **Agile Ceremonies**: Interactive planning, retrospectives, and dailies

## Installation

### From Source

```bash
git clone https://github.com/user/jira
cd jira
task build
```

Then install to your system:

```bash
# Install to user-local directory (no admin required)
./build/jira install

# Or install system-wide (may require sudo/admin)
./build/jira install --global
```

### Download Binary

Download the latest release from the releases page, then run:

```bash
jira install
```

The installer will:
- Detect your operating system
- Install to a user-local directory by default (no admin privileges needed)
- Automatically elevate permissions only if required (for `--global` installs)
- Add installation path instructions if not in your PATH

## Quick Start

### 1. Initialize Configuration

```bash
jira init
# Or install to system: jira install
```

This interactive wizard will guide you through:
- Jira URL (e.g., https://your-domain.atlassian.net)
- Email address
- API Token ([create one here](https://id.atlassian.com/manage-profile/security/api-tokens))
- Default project key
- Multi-owner custom field ID (optional)

### 2. Verify Connection

```bash
jira task list
```

## Usage

### Task Commands

```bash
# List tasks
jira task list
jira task list --assignee user@example.com
jira task list --status "In Progress"

# Create a task
jira task create --summary "Fix login bug" --type Bug --assignee dev@example.com

# Create with multiple owners
jira task create --summary "Refactor API" --owners "dev1@example.com,dev2@example.com"

# View task details
jira task view PROJ-123

# Edit a task
jira task edit PROJ-123 --summary "Updated summary"

# Delete a task
jira task delete PROJ-123
```

### TUI Mode

```bash
# Interactive issue browser
jira task list
```

### Project Management

```bash
# List configured projects
jira project list

# Switch default project
jira project switch OTHER

# View project config
jira project config
```

### Cache Management

```bash
# View cache status
jira cache status

# Clear cache
jira cache clear

# Show cache path
jira cache path
```

### Agile Ceremonies

```bash
# Sprint planning
jira ceremony planning

# Retrospective
jira ceremony retro

# Daily standup
jira ceremony daily
```

## Configuration

Configuration is stored in `~/.config/jira-go/config.yaml`:

```yaml
default_project: PROJ
auth:
  email: user@example.com
  api_token: ${JIRA_GO_API_TOKEN}  # Can reference env vars

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
- `JIRA_GO_API_TOKEN` - API token (overrides config)
- `JIRA_GO_EMAIL` - Email (overrides config)
- `JIRA_GO_DEFAULT_PROJECT` - Default project (overrides config)
- `JIRA_GO_CACHE` - Cache file path

## Multi-Owner Configuration

Jira only supports a single assignee per issue. To enable multiple owners:

1. Go to Jira Project Settings → Issue Types → Fields
2. Create a custom field: "Additional Owners" (Multi User Picker type)
3. Note the field ID (e.g., `customfield_10001`)
4. Run `jira init` again and enter the field ID
5. Use `--owners` flag when creating/editing tasks

## Global Flags

All commands support these global flags:

- `-p, --project` - Override default project
- `--no-cache` - Disable cache for this command
- `--cache-ttl` - Custom cache TTL (e.g., `5m`, `1h`)
- `-v, --verbose` - Enable verbose output

## Development

```bash
# Run tests
task test

# Build for development
task dev

# Lint
task lint
```

## License

MIT License
