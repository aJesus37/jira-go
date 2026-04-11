# jira-go Skill Suite — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extend the jira-go CLI with missing commands (task status, task comment, --status on create, --format json, age column, jira report), then write a 5-file Claude skill suite that guides agents through every supported use case.

**Architecture:** CLI extensions are built first against the existing Cobra/API pattern; the skill suite is written last so it documents a complete, working surface. Each CLI task follows TDD: write failing test → implement → pass → commit.

**Tech Stack:** Go, Cobra, Charmbracelet, Jira REST API v3, `task` (Taskfile), `go test`

---

## Task 1: Add `task status` command

Exposes the existing `GetTransitions` + `TransitionIssue` API methods as a CLI command. Finds the transition by name (case-insensitive), with an optional `--comment` flag to attach a message.

**Files:**
- Modify: `internal/commands/task.go`
- Modify: `internal/api/issues.go` (no change needed — API already exists)
- Test: `internal/commands/task_status_test.go` (new)

**Step 1: Write the failing test**

Create `internal/commands/task_status_test.go`:

```go
package commands_test

import (
	"testing"
)

// Integration-style smoke test: ensure the command is registered and rejects bad args
func TestTaskStatusCommandRegistered(t *testing.T) {
	cmd := rootCmd
	statusCmd, _, err := cmd.Find([]string{"task", "status"})
	if err != nil || statusCmd == nil {
		t.Fatalf("task status command not registered: %v", err)
	}
}

func TestTaskStatusRequiresArgs(t *testing.T) {
	_, err := executeCommand(rootCmd, "task", "status")
	if err == nil {
		t.Fatal("expected error when called with no args")
	}
}
```

Note: `executeCommand` helper should call `cmd.ExecuteC()` with args. Add it to a `helpers_test.go` if it doesn't exist yet.

**Step 2: Run test to verify it fails**

```bash
go test ./internal/commands/... -run TestTaskStatus -v
```
Expected: FAIL — `task status command not registered`

**Step 3: Add the command to `internal/commands/task.go`**

After the existing `taskDeleteCmd` block, add:

```go
var taskStatusCmd = &cobra.Command{
	Use:   "status [key] [status]",
	Short: "Transition issue to a new status",
	Args:  cobra.ExactArgs(2),
	RunE:  runTaskStatus,
}
```

Register in `init()`:
```go
taskCmd.AddCommand(taskStatusCmd)
taskStatusCmd.Flags().String("comment", "", "Optional comment to add with the status change")
```

Add `runTaskStatus`:
```go
func runTaskStatus(cmd *cobra.Command, args []string) error {
	key, targetStatus := args[0], args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	client, err := api.NewClient(cfg, cfg.DefaultProject)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	transitions, err := client.GetTransitions(key)
	if err != nil {
		return fmt.Errorf("getting transitions: %w", err)
	}

	var transitionID string
	for _, t := range transitions {
		if strings.EqualFold(t.Name, targetStatus) {
			transitionID = t.ID
			break
		}
	}
	if transitionID == "" {
		var names []string
		for _, t := range transitions {
			names = append(names, t.Name)
		}
		return fmt.Errorf("status %q not found; available: %s", targetStatus, strings.Join(names, ", "))
	}

	if err := client.TransitionIssue(key, transitionID); err != nil {
		return fmt.Errorf("transitioning issue: %w", err)
	}
	fmt.Printf("✓ %s → %s\n", key, targetStatus)

	if comment, _ := cmd.Flags().GetString("comment"); comment != "" {
		if err := client.AddComment(key, comment); err != nil {
			return fmt.Errorf("adding comment: %w", err)
		}
		fmt.Printf("✓ Comment added\n")
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/commands/... -run TestTaskStatus -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/commands/task.go internal/commands/task_status_test.go
git commit -m "feat(task): add task status command with optional --comment flag"
```

---

## Task 2: Add `task comment` command

**Files:**
- Modify: `internal/commands/task.go`
- Test: `internal/commands/task_comment_test.go` (new)

**Step 1: Write the failing test**

Create `internal/commands/task_comment_test.go`:

```go
package commands_test

import (
	"testing"
)

func TestTaskCommentCommandRegistered(t *testing.T) {
	cmd := rootCmd
	commentCmd, _, err := cmd.Find([]string{"task", "comment"})
	if err != nil || commentCmd == nil {
		t.Fatalf("task comment command not registered: %v", err)
	}
}

func TestTaskCommentRequiresTwoArgs(t *testing.T) {
	_, err := executeCommand(rootCmd, "task", "comment", "PROJ-1")
	if err == nil {
		t.Fatal("expected error with only one arg (missing message)")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/commands/... -run TestTaskComment -v
```
Expected: FAIL

**Step 3: Add command to `internal/commands/task.go`**

```go
var taskCommentCmd = &cobra.Command{
	Use:   "comment [key] [message]",
	Short: "Add a comment to an issue",
	Args:  cobra.ExactArgs(2),
	RunE:  runTaskComment,
}
```

Register in `init()`:
```go
taskCmd.AddCommand(taskCommentCmd)
```

Add `runTaskComment`:
```go
func runTaskComment(cmd *cobra.Command, args []string) error {
	key, message := args[0], args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	client, err := api.NewClient(cfg, cfg.DefaultProject)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	if err := client.AddComment(key, message); err != nil {
		return fmt.Errorf("adding comment: %w", err)
	}

	fmt.Printf("✓ Comment added to %s\n", key)
	return nil
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/commands/... -run TestTaskComment -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/commands/task.go internal/commands/task_comment_test.go
git commit -m "feat(task): add task comment command"
```

---

## Task 3: Add `--status` flag to `task create`

After creating an issue, immediately transition it to the requested status. This enables retroactive "done" task logging.

**Files:**
- Modify: `internal/commands/task.go` (runTaskCreate + init flags)
- Test: `internal/commands/task_create_status_test.go` (new)

**Step 1: Write the failing test**

```go
package commands_test

import (
	"testing"
)

func TestTaskCreateHasStatusFlag(t *testing.T) {
	flag := taskCreateCmd.Flags().Lookup("status")
	if flag == nil {
		t.Fatal("task create --status flag not registered")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/commands/... -run TestTaskCreateHasStatusFlag -v
```
Expected: FAIL

**Step 3: Add flag and post-create transition to `task.go`**

In `init()`, add to create flags:
```go
taskCreateCmd.Flags().String("status", "", "Set initial status after creation (e.g. 'Done')")
```

At the end of `runTaskCreate`, before `return nil`, add:
```go
if targetStatus, _ := cmd.Flags().GetString("status"); targetStatus != "" {
	transitions, err := client.GetTransitions(issue.Key)
	if err != nil {
		return fmt.Errorf("getting transitions: %w", err)
	}
	var transitionID string
	for _, t := range transitions {
		if strings.EqualFold(t.Name, targetStatus) {
			transitionID = t.ID
			break
		}
	}
	if transitionID == "" {
		return fmt.Errorf("status %q not available after creation", targetStatus)
	}
	if err := client.TransitionIssue(issue.Key, transitionID); err != nil {
		return fmt.Errorf("setting initial status: %w", err)
	}
	fmt.Printf("✓ Status set to %s\n", targetStatus)
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/commands/... -run TestTaskCreateHasStatusFlag -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/commands/task.go internal/commands/task_create_status_test.go
git commit -m "feat(task): add --status flag to task create for done-state registration"
```

---

## Task 4: Add `--format json` and `--age` column to `task list`

**Files:**
- Modify: `internal/commands/task.go` (runTaskList + init)
- Modify: `internal/models/models.go` (add StatusAge helper method)
- Test: `internal/commands/task_list_format_test.go` (new)

**Step 1: Write the failing test**

```go
package commands_test

import (
	"testing"
)

func TestTaskListHasFormatFlag(t *testing.T) {
	flag := taskListCmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatal("task list --format flag not registered")
	}
}

func TestTaskListHasAgeFlag(t *testing.T) {
	flag := taskListCmd.Flags().Lookup("age")
	if flag == nil {
		t.Fatal("task list --age flag not registered")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/commands/... -run TestTaskListHas -v
```
Expected: FAIL

**Step 3: Add `StatusAge` to the Issue model**

In `internal/models/models.go`, add to the `Issue` struct:
```go
StatusChanged time.Time `json:"status_changed,omitempty"` // When status last changed
```

Add method below the struct:
```go
// DaysInStatus returns the number of days the issue has been in its current status.
// Falls back to days since last update if StatusChanged is not populated.
func (i Issue) DaysInStatus() int {
	ref := i.StatusChanged
	if ref.IsZero() {
		ref = i.Updated
	}
	if ref.IsZero() {
		return 0
	}
	return int(time.Since(ref).Hours() / 24)
}
```

Note: `StatusChanged` is populated from the Jira changelog API (`/rest/api/3/issue/{key}?expand=changelog`). For now, `DaysInStatus` falls back to `Updated` so the field is useful even without changelog data.

**Step 4: Add flags and output logic to `task.go`**

In `init()`, add to list flags:
```go
taskListCmd.Flags().String("format", "table", "Output format: table or json")
taskListCmd.Flags().Bool("age", false, "Show days-in-current-status column")
```

In `runTaskList`, after the issues are collected, replace the output section with:

```go
format, _ := cmd.Flags().GetString("format")
showAge, _ := cmd.Flags().GetBool("age")

if format == "json" {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(issues)
}

// Table output
if showAge {
	fmt.Printf("%-12s %-10s %-12s %-6s %-20s %s\n", "KEY", "TYPE", "STATUS", "DAYS", "ASSIGNEE", "SUMMARY")
} else {
	fmt.Printf("%-12s %-10s %-12s %-20s %s\n", "KEY", "TYPE", "STATUS", "ASSIGNEE", "SUMMARY")
}
fmt.Println(strings.Repeat("-", 100))

for _, issue := range issues {
	// ... existing truncation logic ...
	if showAge {
		fmt.Printf("%-12s %-10s %-12s %-6d %-20s %s\n", issue.Key, issue.Type, status, issue.DaysInStatus(), displayAssignee, summary)
	} else {
		fmt.Printf("%-12s %-10s %-12s %-20s %s\n", issue.Key, issue.Type, status, displayAssignee, summary)
	}
}
```

Also add `"encoding/json"` and `"os"` to imports if not present.

**Step 5: Run test to verify it passes**

```bash
go test ./internal/commands/... -run TestTaskListHas -v
```
Expected: PASS

**Step 6: Commit**

```bash
git add internal/commands/task.go internal/models/models.go internal/commands/task_list_format_test.go
git commit -m "feat(task): add --format json and --age column to task list"
```

---

## Task 5: Add `jira report` command

A new top-level report command that aggregates tasks per assignee with age-in-status data.

**Files:**
- Create: `internal/commands/report.go`
- Test: `internal/commands/report_test.go` (new)

**Step 1: Write the failing test**

Create `internal/commands/report_test.go`:

```go
package commands_test

import (
	"testing"
)

func TestReportCommandRegistered(t *testing.T) {
	cmd := rootCmd
	reportCmd, _, err := cmd.Find([]string{"report"})
	if err != nil || reportCmd == nil {
		t.Fatalf("report command not registered: %v", err)
	}
}

func TestReportHasFormatFlag(t *testing.T) {
	cmd := rootCmd
	reportCmd, _, _ := cmd.Find([]string{"report"})
	flag := reportCmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatal("report --format flag not registered")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/commands/... -run TestReport -v
```
Expected: FAIL

**Step 3: Create `internal/commands/report.go`**

```go
// internal/commands/report.go
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/api"
	"github.com/user/jira-go/internal/config"
	"github.com/user/jira-go/internal/models"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Project status report for project managers",
	Long:  `Summarize tasks per assignee with age-in-status. Use --format json for agent-parseable output.`,
	RunE:  runReport,
}

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.Flags().String("format", "table", "Output format: table or json")
	reportCmd.Flags().String("sprint", "", "Filter by sprint: active, future, closed, or sprint ID")
	reportCmd.Flags().String("assignee", "", "Filter by assignee email")
	reportCmd.Flags().String("project", "", "Project key (defaults to config)")
}

// AssigneeSummary aggregates issue data per person
type AssigneeSummary struct {
	Name       string         `json:"name"`
	Email      string         `json:"email"`
	Total      int            `json:"total"`
	ByStatus   map[string]int `json:"by_status"`
	AvgAgeDays int            `json:"avg_age_days"`
	OldestDays int            `json:"oldest_days"`
	Issues     []models.Issue `json:"issues,omitempty"`
}

func runReport(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := getProjectKey(cmd, cfg)
	project, _ := cfg.GetProject(projectKey)

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	jql := fmt.Sprintf("project = %s AND statusCategory != Done", projectKey)

	if sprint, _ := cmd.Flags().GetString("sprint"); sprint != "" {
		switch sprint {
		case "active":
			jql += " AND sprint in openSprints()"
		case "future":
			jql += " AND sprint in futureSprints()"
		case "closed":
			jql += " AND sprint in closedSprints()"
		default:
			jql += fmt.Sprintf(" AND sprint = %s", sprint)
		}
	}

	if assigneeEmail, _ := cmd.Flags().GetString("assignee"); assigneeEmail != "" {
		user, err := client.ResolveEmail(assigneeEmail)
		if err != nil {
			return fmt.Errorf("resolving assignee: %w", err)
		}
		jql += fmt.Sprintf(" AND assignee = '%s'", user.AccountID)
	}

	resp, err := client.SearchIssues(jql, 0, 200, project.MultiOwnerField, project.SprintField)
	if err != nil {
		return fmt.Errorf("searching issues: %w", err)
	}

	// Aggregate by assignee
	byAssignee := map[string]*AssigneeSummary{}

	for _, issue := range resp.Issues {
		var email, name string
		if issue.Assignee != nil {
			email = issue.Assignee.Email
			name = issue.Assignee.DisplayName
		} else {
			email = "unassigned"
			name = "Unassigned"
		}

		if _, ok := byAssignee[email]; !ok {
			byAssignee[email] = &AssigneeSummary{
				Name:     name,
				Email:    email,
				ByStatus: map[string]int{},
			}
		}

		s := byAssignee[email]
		s.Total++
		s.ByStatus[issue.Status]++
		age := issue.DaysInStatus()
		s.AvgAgeDays += age
		if age > s.OldestDays {
			s.OldestDays = age
		}
		s.Issues = append(s.Issues, issue)
	}

	// Compute averages
	var summaries []AssigneeSummary
	for _, s := range byAssignee {
		if s.Total > 0 {
			s.AvgAgeDays = s.AvgAgeDays / s.Total
		}
		summaries = append(summaries, *s)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Total > summaries[j].Total
	})

	format, _ := cmd.Flags().GetString("format")
	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(summaries)
	}

	// Table output
	fmt.Printf("\nProject: %s — Active Tasks Report\n", projectKey)
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("%-25s %-6s %-10s %-10s\n", "ASSIGNEE", "TOTAL", "AVG DAYS", "OLDEST")
	fmt.Println(strings.Repeat("-", 55))
	for _, s := range summaries {
		name := s.Name
		if len(name) > 23 {
			name = name[:20] + "..."
		}
		fmt.Printf("%-25s %-6d %-10d %-10d\n", name, s.Total, s.AvgAgeDays, s.OldestDays)

		// Print status breakdown
		for status, count := range s.ByStatus {
			fmt.Printf("  %-23s %-6d\n", status, count)
		}
	}
	fmt.Printf("\nTotal active tickets: %d\n", len(resp.Issues))

	return nil
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/commands/... -run TestReport -v
```
Expected: PASS

**Step 5: Run all tests**

```bash
task test
```
Expected: all pass

**Step 6: Commit**

```bash
git add internal/commands/report.go internal/commands/report_test.go
git commit -m "feat: add report command with per-assignee summary and age-in-status"
```

---

## Task 6: Write `jira-go` primary skill (router + cheatsheet)

The top-level skill that orients agents and covers the most common one-liners.

**Files:**
- Create: `~/.claude/skills/jira-go.md`

**Step 1: Create the skill file**

```markdown
---
name: jira-go
description: Use when an agent needs to interact with Jira via the jira-go CLI — creating tasks, updating status, managing sprints, epics, or getting a PM overview. This is the entry point; it routes to sub-skills for complex operations.
type: skill
---

# jira-go CLI — Quick Reference

## Prerequisites

- Config at `~/.config/jira-go/config.yaml` with `default_project`, auth, and board ID
- Binary: `jira` (built from this repo)
- Always use `--no-interactive` in automation contexts to get plain text output

## Route to Sub-Skills

| Intent | Sub-skill |
|---|---|
| Create, update, comment, list tasks | `jira-go:tasks` |
| Epics and task/subtask hierarchy | `jira-go:epics` |
| Sprint lifecycle, board, moving issues | `jira-go:sprints` |
| PM workload and age-in-status report | `jira-go:reports` |

## Most Common Commands

\`\`\`bash
# List active tasks (plain text)
jira task list --no-interactive --active

# List tasks as JSON (for agents)
jira task list --no-interactive --format json

# Create a task
jira task create --summary "Fix login bug" --assignee user@example.com

# View a task
jira task view PROJ-123

# Transition status
jira task status PROJ-123 "In Progress"

# Add comment
jira task comment PROJ-123 "Reviewed — looks good"

# PM report
jira report --sprint active

# Sprint board
jira sprint board
\`\`\`

## Key Flags

| Flag | Applies to | Effect |
|---|---|---|
| `--no-interactive` | list commands | Plain text, no TUI |
| `--format json` | list, report | Machine-readable JSON |
| `--project KEY` | most commands | Override default project |
| `--assignee email` | task list, report | Filter by assignee |
| `--active` | task list | Exclude Done tickets |
| `--age` | task list | Show days-in-status column |
```

**Step 2: Verify skill is reachable**

```bash
ls ~/.claude/skills/jira-go.md
```
Expected: file exists

**Step 3: Commit**

```bash
git add ~/.claude/skills/jira-go.md 2>/dev/null || true
git -C ~/.claude commit -m "feat: add jira-go primary skill" 2>/dev/null || echo "Committed in skills repo or standalone"
```

---

## Task 7: Write `jira-go:tasks` sub-skill

**Files:**
- Create: `~/.claude/skills/jira-go:tasks.md`

**Step 1: Create the skill file**

```markdown
---
name: jira-go:tasks
description: Full task CRUD for the jira-go CLI — create, view, edit, delete, transition status, add comments, list with filters. Includes creating tasks directly in Done state for retroactive logging.
type: skill
---

# jira-go — Tasks

## Create a Task

\`\`\`bash
# Basic
jira task create --summary "Task title"

# With all options
jira task create \
  --summary "Implement login flow" \
  --description "Add OAuth2 + session management" \
  --type Task \
  --assignee dev@example.com \
  --owners "dev@example.com,qa@example.com"

# Create directly in Done (retroactive logging)
jira task create --summary "Deployed hotfix" --status "Done"

# Create a subtask (parent must exist)
jira task create --type Sub-task --parent PROJ-123 --summary "Write unit tests"
\`\`\`

## List Tasks

\`\`\`bash
# All active tasks
jira task list --no-interactive --active

# With age column (days in current status)
jira task list --no-interactive --active --age

# JSON output (for agents)
jira task list --no-interactive --format json

# Filter by assignee
jira task list --no-interactive --assignee user@example.com

# Backlog only (not in any sprint)
jira task list --no-interactive --backlog
\`\`\`

## View / Edit / Delete

\`\`\`bash
jira task view PROJ-123
jira task edit PROJ-123 --summary "New title" --assignee other@example.com
jira task delete PROJ-123
\`\`\`

## Transition Status

\`\`\`bash
# Simple transition
jira task status PROJ-123 "In Progress"

# Transition + comment in one step
jira task status PROJ-123 "Done" --comment "Deployed to production"
\`\`\`

Status names are case-insensitive and must match available transitions for that issue.
If the status is not found, the CLI will list available options.

## Add a Comment

\`\`\`bash
jira task comment PROJ-123 "Blocked by PROJ-99 — waiting on API changes"
\`\`\`

## Common Patterns

**Register completed work retroactively:**
\`\`\`bash
jira task create \
  --summary "Migrated prod DB to v2 schema" \
  --description "Ran migration script, verified row counts, updated runbook" \
  --status "Done" \
  --assignee engineer@example.com
\`\`\`

**Bulk status update with context:**
\`\`\`bash
jira task status PROJ-45 "In Review" --comment "PR #234 open for review"
jira task status PROJ-46 "In Review" --comment "PR #235 open for review"
\`\`\`
```

**Step 2: Verify file exists, commit** (same pattern as Task 6 Step 2–3)

---

## Task 8: Write `jira-go:epics` sub-skill

**Files:**
- Create: `~/.claude/skills/jira-go:epics.md`

**Step 1: Create the skill file**

```markdown
---
name: jira-go:epics
description: Epic management for jira-go — create epics, add tasks, create subtasks, view full hierarchy. Supports both flat (Epic→Tasks) and 3-level (Epic→Tasks→Subtasks) models.
type: skill
---

# jira-go — Epics

## Create an Epic

\`\`\`bash
jira epic create --summary "Q3 Auth Overhaul" --description "Replace legacy session system with OAuth2"
\`\`\`

## List Epics

\`\`\`bash
jira epic list
\`\`\`

## View Epic with Child Issues

\`\`\`bash
jira epic view PROJ-1
\`\`\`

## Add Existing Tasks to an Epic

\`\`\`bash
jira epic add PROJ-1 PROJ-45 PROJ-46 PROJ-47
\`\`\`

## Remove Tasks from an Epic

\`\`\`bash
jira epic remove PROJ-45 PROJ-46
\`\`\`

## Create Tasks Directly Under an Epic

\`\`\`bash
# Create task and immediately add to epic
jira task create --summary "Design OAuth2 flow" --assignee dev@example.com
jira epic add PROJ-1 PROJ-<new-key>
\`\`\`

## Create Subtasks (3-level hierarchy)

\`\`\`bash
# Subtask under a task (parent = task key)
jira task create \
  --type Sub-task \
  --parent PROJ-45 \
  --summary "Write integration tests for OAuth callback"
\`\`\`

## Typical Epic Setup Workflow

\`\`\`bash
# 1. Create the epic
jira epic create --summary "Payment Gateway v2"

# 2. Create tasks (note returned keys)
jira task create --summary "Design API contract" --assignee backend@example.com
jira task create --summary "Implement Stripe integration" --assignee backend@example.com
jira task create --summary "QA regression suite" --assignee qa@example.com

# 3. Link all tasks to epic
jira epic add PROJ-<epic-key> PROJ-<task1> PROJ-<task2> PROJ-<task3>
\`\`\`
```

**Step 2: Commit** (same pattern)

---

## Task 9: Write `jira-go:sprints` sub-skill

**Files:**
- Create: `~/.claude/skills/jira-go:sprints.md`

**Step 1: Create the skill file**

```markdown
---
name: jira-go:sprints
description: Sprint lifecycle management for jira-go — create sprints with auto-calculated 2-week dates, start, complete, move issues, view board.
type: skill
---

# jira-go — Sprints

## Create a Sprint (2-week default)

Calculate: start = next Monday from today, end = start + 14 days.

\`\`\`bash
# Auto-calculated example (you must compute the dates):
# Today: 2026-04-11 → next Monday: 2026-04-13 → end: 2026-04-27
jira sprint create \
  --name "Sprint 12" \
  --goal "Ship auth overhaul" \
  --start 2026-04-13 \
  --end 2026-04-27
\`\`\`

**Date calculation rule:**
- Start: nearest upcoming Monday (or today if today is Monday)
- End: start + 14 days

## List Sprints

\`\`\`bash
jira sprint list                    # All sprints
jira sprint list --state active     # Active only
jira sprint list --state future     # Upcoming
jira sprint list --state closed     # Past sprints
\`\`\`

## Start a Sprint

\`\`\`bash
jira sprint start 42   # 42 = sprint ID from sprint list
\`\`\`

## Complete a Sprint

\`\`\`bash
jira sprint complete 42
\`\`\`

## Move Issues to a Sprint

\`\`\`bash
jira sprint move 42 PROJ-1 PROJ-2 PROJ-3
\`\`\`

## View Sprint Board (Kanban)

\`\`\`bash
jira sprint board          # Active sprint board (interactive TUI)
jira sprint board 42       # Specific sprint board
jira sprint board --no-interactive  # Plain text output
\`\`\`

## List Issues in a Sprint

\`\`\`bash
jira sprint issues 42
\`\`\`

## Edit Sprint

\`\`\`bash
jira sprint edit 42   # Interactive: name, goal, dates
\`\`\`

## Typical Sprint Kick-off

\`\`\`bash
# 1. Create sprint with 2-week window
jira sprint create --name "Sprint 12" --goal "..." --start 2026-04-13 --end 2026-04-27

# 2. Note the sprint ID from output, move planned issues
jira sprint move <id> PROJ-10 PROJ-11 PROJ-12

# 3. Start the sprint
jira sprint start <id>
\`\`\`
```

**Step 2: Commit** (same pattern)

---

## Task 10: Write `jira-go:reports` sub-skill

**Files:**
- Create: `~/.claude/skills/jira-go:reports.md`

**Step 1: Create the skill file**

```markdown
---
name: jira-go:reports
description: PM-facing reports for jira-go — per-assignee workload, age-in-status, sprint health. Supports JSON output for agent parsing.
type: skill
---

# jira-go — Reports

## Full Active Workload Report

\`\`\`bash
jira report
\`\`\`

Shows: each assignee, total active tickets, average days in current status, oldest ticket age, breakdown by status.

## Filter by Sprint

\`\`\`bash
jira report --sprint active    # Current sprint only
jira report --sprint future    # Upcoming sprint
jira report --sprint 42        # Specific sprint ID
\`\`\`

## Filter by Assignee

\`\`\`bash
jira report --assignee user@example.com
\`\`\`

## JSON Output (for agents)

\`\`\`bash
jira report --format json
\`\`\`

Output schema:
\`\`\`json
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
\`\`\`

## Task List with Age Column

\`\`\`bash
# See how long each task has been in its current status
jira task list --no-interactive --active --age

# JSON with age data
jira task list --no-interactive --format json
\`\`\`

## Typical PM Review Workflow

\`\`\`bash
# 1. Get overview of current sprint
jira report --sprint active

# 2. Investigate a team member with high oldest_days
jira task list --no-interactive --assignee dev@example.com --active --age

# 3. View the oldest ticket
jira task view PROJ-99

# 4. Add a follow-up comment
jira task comment PROJ-99 "Checking in — any blockers?"
\`\`\`
```

**Step 2: Commit** (same pattern)

---

## Task 11: Update AGENTS.md with new commands

Add the new commands to the AGENTS.md reference so future agents have a complete picture.

**Files:**
- Modify: `AGENTS.md`

**Step 1: Add to the Commands section in AGENTS.md**

Under the existing task/sprint/epic tables, add entries for:
- `jira task status [key] [status] [--comment]`
- `jira task comment [key] "message"`
- `jira task create --status "Done"` (retroactive)
- `jira task list --format json --age`
- `jira report [--sprint] [--assignee] [--format json]`

**Step 2: Commit**

```bash
git add AGENTS.md
git commit -m "docs: update AGENTS.md with new task status, comment, and report commands"
```

---

## Task 12: Run full test suite and verify build

**Step 1: Run all tests**

```bash
task test
```
Expected: all pass

**Step 2: Build**

```bash
task build
```
Expected: `./build/jira` produced with no errors

**Step 3: Smoke test new commands**

```bash
./build/jira task --help          # Should show status and comment subcommands
./build/jira task create --help   # Should show --status flag
./build/jira task list --help     # Should show --format and --age flags
./build/jira report --help        # Should show report command
```

**Step 4: Final commit if any fixes needed, otherwise done**
