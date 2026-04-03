# Contributing to jira-go

Thank you for your interest in contributing to jira-go! This document provides guidelines and instructions for contributing.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/jira-go.git`
3. Install dependencies: `task deps`
4. Build the project: `task build`
5. Run tests: `task test`

## Development Setup

### Prerequisites

- Go 1.21 or later
- [Task](https://taskfile.dev/) (build runner)
- A Jira Cloud instance for testing (optional but recommended)

### Build Commands

```bash
# Build the binary
task build

# Run tests
task test

# Run in development mode
task dev

# Format code
task fmt

# Run linter
task lint
```

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/user/jira-go/issues)
2. If not, create a new issue with:
   - Clear description of the bug
   - Steps to reproduce
   - Expected vs actual behavior
   - Your environment (OS, Go version, Jira version)
   - Relevant logs or error messages

### Suggesting Features

1. Check if the feature has already been suggested in [Issues](https://github.com/user/jira-go/issues)
2. If not, create a new issue with the `enhancement` label
3. Describe the feature and its use case
4. Explain why it would be useful to most users

### Pull Requests

1. Create a new branch: `git checkout -b feature/my-feature`
2. Make your changes
3. Add or update tests as needed
4. Ensure tests pass: `task test`
5. Format your code: `task fmt`
6. Commit with a clear message (see [Commit Messages](#commit-messages))
7. Push to your fork: `git push origin feature/my-feature`
8. Open a Pull Request

## Code Guidelines

### Go Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Keep functions focused and small
- Document public APIs with Go doc comments
- Handle errors explicitly, don't ignore them

### Example

```go
// Good
func (c *Client) GetIssue(key string) (*Issue, error) {
    resp, err := c.Get(fmt.Sprintf("/issue/%s", key))
    if err != nil {
        return nil, fmt.Errorf("fetching issue: %w", err)
    }
    defer resp.Body.Close()
    
    // ...
}

// Bad - ignores error
func (c *Client) GetIssue(key string) *Issue {
    resp, _ := c.Get(fmt.Sprintf("/issue/%s", key))
    // ...
}
```

### Error Handling

- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Return meaningful error messages to users
- Use validation in `config.Validate()` for config errors

### Testing

- Write tests for new functionality
- Use table-driven tests where appropriate
- Mock external dependencies (API calls)
- Ensure existing tests still pass

## Project Structure

```
cmd/jira/              # Main entry point
internal/
├── commands/          # CLI commands
├── api/               # Jira API client
├── cache/             # SQLite caching
├── config/            # Configuration management
├── models/            # Domain models
└── tui/               # Bubble Tea TUI components
```

## Commit Messages

Use clear, descriptive commit messages:

```
feat: add sprint edit command
fix: resolve sprint completion API error
docs: update README with new commands
test: add tests for sprint operations
refactor: simplify error handling in API client
chore: update dependencies
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code restructuring
- `chore`: Build/tooling changes

## Testing with Real Jira

To test with a real Jira instance:

1. Run `jira init` to configure:
   - Jira URL (e.g., https://your-domain.atlassian.net)
   - Email address
   - API token
   - Project key

2. Create a test sprint and issues
3. Test your changes

## Documentation

- Update README.md if adding user-facing features
- Update AGENTS.md if changing architecture or development workflows
- Add comments to complex code
- Update CHANGELOG.md for notable changes

## Code Review Process

1. All PRs require at least one review
2. Address review comments promptly
3. Keep PRs focused on a single change
4. Ensure CI checks pass

## Community

- Be respectful and constructive
- Help others in issues and discussions
- Share your use cases and feedback

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

Feel free to open an issue for:
- Questions about the codebase
- Help getting started
- Discussion of potential features

Thank you for contributing!
