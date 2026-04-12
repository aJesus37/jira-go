# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Interactive sprint editing with `jira sprint edit` command
- Sprint management: create, start, complete, edit sprints
- Task management with full CRUD operations
- Multi-owner support via custom fields
- Rich TUI with Bubble Tea for task list and kanban board
- Markdown rendering with syntax highlighting
- Daily standup ceremony with timer
- Sprint planning and retrospective ceremonies
- SQLite caching with TTL support
- Configuration via YAML and environment variables
- Sprint board column visibility toggle with `x` key
- Sprint board column resizing with `+/-` keys
- Sprint board column reordering with `[`/`]` keys
- Sprint board focus cycling between visible/hidden columns with `f` key
- Configurable assignee+owner merging via `merge_assignee_owner` setting

### Features
- List, view, create, edit, delete Jira issues
- Assign multiple owners to issues via custom fields
- View sprint boards in interactive kanban view with column management
- Toggle, resize, and reorder sprint board columns
- Interactive task detail view with markdown rendering
- Cache management commands
- Project switching and configuration

## [0.1.0] - 2026-04-03

### Added
- Initial release
- Basic task management (list, view, create, edit, delete)
- Sprint operations (list, create, start, complete)
- Epic management framework
- Cache management
- Configuration wizard via `jira init`
- TUI for interactive task browsing

[Unreleased]: https://github.com/aJesus37/jira-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/aJesus37/jira-go/releases/tag/v0.1.0
