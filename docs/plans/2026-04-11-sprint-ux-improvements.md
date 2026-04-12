# Sprint UX Improvements Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix 6 UX gaps in sprint-related commands discovered during real usage: bulk issue moving, JSON output on sprint issues, sprint keyword aliases, FUTURE→CLOSED auto-handling, sprint_field warning, and pagination.

**Architecture:** All changes are in `internal/commands/sprint.go` and `internal/api/sprint.go`. A shared `resolveSprintID` helper centralises keyword→ID resolution. Each task is independently committable and non-breaking.

**Tech Stack:** Go, Cobra, Jira Agile REST API v1 (`/rest/agile/1.0/`)

---

## Task 1: Add `--format json`, `--status`, and `--limit` to `sprint issues`

Currently `sprint issues` outputs a hardcoded text table with a silent 100-issue cap and no filtering. This task adds machine-readable output and status filtering.

**Files:**
- Modify: `internal/commands/sprint.go` — `runSprintIssues` (lines 259–316) and `init()` (lines 81–100)
- Modify: `internal/api/sprint.go` — `GetSprintIssues` (lines 256–266)
- Test: `internal/commands/sprint_issues_flags_test.go` (new)

**Step 1: Write failing tests**

Create `internal/commands/sprint_issues_flags_test.go`:

```go
package commands_test

import (
	"testing"
	"github.com/user/jira-go/internal/commands"
)

func TestSprintIssuesHasFormatFlag(t *testing.T) {
	cmd, _, err := commands.RootCmd.Find([]string{"sprint", "issues"})
	if err != nil || cmd == nil {
		t.Fatalf("sprint issues not found: %v", err)
	}
	if f := cmd.Flags().Lookup("format"); f == nil {
		t.Fatal("sprint issues --format flag not registered")
	}
}

func TestSprintIssuesHasStatusFlag(t *testing.T) {
	cmd, _, _ := commands.RootCmd.Find([]string{"sprint", "issues"})
	if f := cmd.Flags().Lookup("status"); f == nil {
		t.Fatal("sprint issues --status flag not registered")
	}
}

func TestSprintIssuesHasLimitFlag(t *testing.T) {
	cmd, _, _ := commands.RootCmd.Find([]string{"sprint", "issues"})
	if f := cmd.Flags().Lookup("limit"); f == nil {
		t.Fatal("sprint issues --limit flag not registered")
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/commands/... -run TestSprintIssues -v
```
Expected: FAIL — flags not registered

**Step 3: Update `GetSprintIssues` in `internal/api/sprint.go`**

Change the signature to accept `limit int` and `status string`:

```go
// GetSprintIssues retrieves issues in a sprint using JQL search.
// Pass limit=0 to use default (100). Pass status="" for no filter.
func (c *Client) GetSprintIssues(sprintID int, ownerFieldID, sprintFieldID, status string, limit int) ([]models.Issue, int, error) {
	jql := fmt.Sprintf("sprint = %d", sprintID)
	if status != "" {
		jql += fmt.Sprintf(" AND status = %q", status)
	}
	if limit <= 0 {
		limit = 100
	}

	result, err := c.SearchIssues(jql, 0, limit, ownerFieldID, sprintFieldID)
	if err != nil {
		return nil, 0, fmt.Errorf("searching sprint issues: %w", err)
	}

	return result.Issues, result.Total, nil
}
```

Note: `result.Total` is the true Jira total (may exceed limit). Return it so callers can warn about truncation.

**Step 4: Fix all callers of the old `GetSprintIssues` signature**

There are two callers — update both:

In `runSprintIssues` (sprint.go line 278), change:
```go
issues, err := client.GetSprintIssues(sprintID, project.MultiOwnerField, project.SprintField)
```
to (temporary, will be replaced fully in Step 5):
```go
issues, _, err := client.GetSprintIssues(sprintID, project.MultiOwnerField, project.SprintField, "", 100)
```

In `runSprintBoard` (sprint.go line 375), change:
```go
issues, err := client.GetSprintIssues(sprintID, project.MultiOwnerField, project.SprintField)
```
to:
```go
issues, _, err := client.GetSprintIssues(sprintID, project.MultiOwnerField, project.SprintField, "", 100)
```

**Step 5: Register flags in `init()` and update `runSprintIssues`**

In `init()`, add after the existing sprint flags:
```go
// Issues flags
sprintIssuesCmd.Flags().String("format", "table", "Output format: table or json")
sprintIssuesCmd.Flags().String("status", "", "Filter by status name (e.g. 'Em andamento')")
sprintIssuesCmd.Flags().Int("limit", 100, "Maximum issues to fetch")
```

Replace `runSprintIssues` entirely:

```go
func runSprintIssues(cmd *cobra.Command, args []string) error {
	sprintID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid sprint ID: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	format, _ := cmd.Flags().GetString("format")
	if format != "table" && format != "json" {
		return fmt.Errorf("unknown format %q: must be table or json", format)
	}

	statusFilter, _ := cmd.Flags().GetString("status")
	limit, _ := cmd.Flags().GetInt("limit")

	projectKey := getProjectKey(cmd, cfg)
	project, _ := cfg.GetProject(projectKey)

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	issues, total, err := client.GetSprintIssues(sprintID, project.MultiOwnerField, project.SprintField, statusFilter, limit)
	if err != nil {
		return fmt.Errorf("fetching sprint issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No issues in sprint")
		return nil
	}

	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(issues)
	}

	fmt.Printf("Issues in sprint %d:\n\n", sprintID)
	fmt.Printf("%-12s %-10s %-12s %-20s %s\n", "KEY", "TYPE", "STATUS", "ASSIGNEE", "SUMMARY")
	fmt.Println(strings.Repeat("-", 100))

	for _, issue := range issues {
		status := issue.Status
		if len(status) > 12 {
			status = status[:9] + "..."
		}
		assignee := "Unassigned"
		if issue.Assignee != nil {
			assignee = issue.Assignee.DisplayName
			if len(assignee) > 18 {
				assignee = assignee[:15] + "..."
			}
		}
		summary := issue.Summary
		if len(summary) > 40 {
			summary = summary[:37] + "..."
		}
		fmt.Printf("%-12s %-10s %-12s %-20s %s\n", issue.Key, issue.Type, status, assignee, summary)
	}

	if total > len(issues) {
		fmt.Printf("\nShowing %d of %d issues (use --limit to fetch more)\n", len(issues), total)
	} else {
		fmt.Printf("\nTotal: %d issues\n", len(issues))
	}
	return nil
}
```

Add `"encoding/json"` and `"os"` to the import block if not already present.

**Step 6: Run tests to verify they pass**

```bash
go test ./internal/commands/... -run TestSprintIssues -v
```
Expected: PASS

**Step 7: Run full test suite**

```bash
go test ./...
```
Expected: all pass

**Step 8: Commit**

```bash
git add internal/commands/sprint.go internal/api/sprint.go internal/commands/sprint_issues_flags_test.go
git commit -m "feat(sprint): add --format json, --status, --limit to sprint issues"
```

---

## Task 2: Add `--from-sprint` bulk move to `sprint move`

Moving all incomplete tickets from one sprint to another currently requires extracting keys manually. `--from-sprint <id>` fetches all non-done issues from the source sprint and moves them in one step.

**Files:**
- Modify: `internal/commands/sprint.go` — `runSprintMove` and `init()`
- Test: `internal/commands/sprint_move_test.go` (new)

**Step 1: Write failing test**

Create `internal/commands/sprint_move_test.go`:

```go
package commands_test

import (
	"testing"
	"github.com/user/jira-go/internal/commands"
)

func TestSprintMoveHasFromSprintFlag(t *testing.T) {
	cmd, _, err := commands.RootCmd.Find([]string{"sprint", "move"})
	if err != nil || cmd == nil {
		t.Fatalf("sprint move not found: %v", err)
	}
	if f := cmd.Flags().Lookup("from-sprint"); f == nil {
		t.Fatal("sprint move --from-sprint flag not registered")
	}
}

func TestSprintMoveRequiresAtLeastTargetID(t *testing.T) {
	_, err := executeCommand(commands.RootCmd, "sprint", "move")
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/commands/... -run TestSprintMove -v
```
Expected: FAIL

**Step 3: Update `sprintMoveCmd` and `init()` in `sprint.go`**

Change the command definition to allow 1 arg (target sprint only, when `--from-sprint` is used):

```go
var sprintMoveCmd = &cobra.Command{
	Use:   "move [target-sprint-id] [issue-keys...]",
	Short: "Move issues to sprint",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSprintMove,
}
```

In `init()`, add flag:
```go
sprintMoveCmd.Flags().String("from-sprint", "", "Move all non-done issues from this sprint ID")
```

Replace `runSprintMove`:

```go
func runSprintMove(cmd *cobra.Command, args []string) error {
	targetSprintID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid target sprint ID: %w", err)
	}

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

	var issueKeys []string

	fromSprint, _ := cmd.Flags().GetString("from-sprint")
	if fromSprint != "" {
		sourceID, err := strconv.Atoi(fromSprint)
		if err != nil {
			return fmt.Errorf("invalid --from-sprint ID: %w", err)
		}
		// Fetch all non-done issues from source sprint
		issues, _, err := client.GetSprintIssues(sourceID, project.MultiOwnerField, project.SprintField, "", 500)
		if err != nil {
			return fmt.Errorf("fetching source sprint issues: %w", err)
		}
		for _, issue := range issues {
			// Skip done/rejected statuses — use statusCategory instead via JQL
			issueKeys = append(issueKeys, issue.Key)
		}
		if len(issueKeys) == 0 {
			fmt.Println("No issues to move")
			return nil
		}
	} else {
		if len(args) < 2 {
			return fmt.Errorf("provide issue keys as arguments, or use --from-sprint to bulk move")
		}
		issueKeys = args[1:]
	}

	if err := client.MoveIssuesToSprint(targetSprintID, issueKeys); err != nil {
		return fmt.Errorf("moving issues: %w", err)
	}

	fmt.Printf("✓ Moved %d issue(s) to sprint %d\n", len(issueKeys), targetSprintID)
	return nil
}
```

Note on `--from-sprint` behaviour: `GetSprintIssues` with `status=""` returns all issues. To get only non-done issues, update the JQL in the API: in `GetSprintIssues`, when `status == ""`, use `sprint = X AND statusCategory != Done` as the default instead of plain `sprint = X`. This avoids moving already-done tickets.

**Update `GetSprintIssues` default JQL in `internal/api/sprint.go`:**

```go
func (c *Client) GetSprintIssues(sprintID int, ownerFieldID, sprintFieldID, status string, limit int) ([]models.Issue, int, error) {
	jql := fmt.Sprintf("sprint = %d", sprintID)
	if status != "" {
		jql += fmt.Sprintf(" AND status = %q", status)
	}
	// Note: returns ALL issues including done. Callers filter as needed.
	// ...
}
```

Actually keep it returning all issues. In `runSprintMove` with `--from-sprint`, add a JQL status filter by calling `SearchIssues` directly:

```go
if fromSprint != "" {
	sourceID, err := strconv.Atoi(fromSprint)
	if err != nil {
		return fmt.Errorf("invalid --from-sprint ID: %w", err)
	}
	jql := fmt.Sprintf("sprint = %d AND statusCategory != Done", sourceID)
	resp, err := client.SearchIssues(jql, 0, 500, project.MultiOwnerField, project.SprintField)
	if err != nil {
		return fmt.Errorf("fetching source sprint issues: %w", err)
	}
	for _, issue := range resp.Issues {
		issueKeys = append(issueKeys, issue.Key)
	}
	if len(issueKeys) == 0 {
		fmt.Println("No non-done issues to move")
		return nil
	}
}
```

**Step 4: Run tests**

```bash
go test ./internal/commands/... -run TestSprintMove -v
```
Expected: PASS

**Step 5: Run full suite**

```bash
go test ./...
```

**Step 6: Commit**

```bash
git add internal/commands/sprint.go internal/commands/sprint_move_test.go
git commit -m "feat(sprint): add --from-sprint flag to bulk move non-done issues"
```

---

## Task 3: Sprint keyword aliases (`active`, `future`)

Commands that take a sprint ID should also accept `active` and `future` as keywords, resolved automatically against the board's open sprints. Affected commands: `sprint issues`, `sprint complete`, `sprint start`, `sprint move`.

**Files:**
- Modify: `internal/commands/sprint.go` — add `resolveSprintID` helper; update `runSprintIssues`, `runSprintComplete`, `runSprintStart`, `runSprintMove`
- Test: `internal/commands/sprint_resolve_test.go` (new)

**Step 1: Write failing test**

Create `internal/commands/sprint_resolve_test.go`:

```go
package commands_test

import (
	"testing"
	"github.com/user/jira-go/internal/commands"
)

func TestSprintIssuesAcceptsKeywordArg(t *testing.T) {
	// Verify the command no longer uses ExactArgs(1) with numeric-only validation
	// by checking it is still registered (keyword validation happens at runtime)
	cmd, _, err := commands.RootCmd.Find([]string{"sprint", "issues"})
	if err != nil || cmd == nil {
		t.Fatalf("sprint issues not found: %v", err)
	}
	if cmd.Args == nil {
		t.Fatal("expected args validator on sprint issues")
	}
}
```

**Step 2: Run test to verify it passes already (structural check)**

```bash
go test ./internal/commands/... -run TestSprintIssuesAcceptsKeywordArg -v
```

**Step 3: Add `resolveSprintID` helper to `sprint.go`**

Add this function above `runSprintList`:

```go
// resolveSprintID converts a string argument to a sprint ID integer.
// Accepts numeric IDs directly, or keywords "active"/"future" which
// are resolved against the configured board's open sprints.
func resolveSprintID(arg string, client *api.Client, boardID int) (int, error) {
	switch arg {
	case "active":
		sprints, err := client.GetSprints(boardID, "active")
		if err != nil {
			return 0, fmt.Errorf("fetching active sprints: %w", err)
		}
		if len(sprints) == 0 {
			return 0, fmt.Errorf("no active sprint found")
		}
		return sprints[0].ID, nil
	case "future":
		sprints, err := client.GetSprints(boardID, "future")
		if err != nil {
			return 0, fmt.Errorf("fetching future sprints: %w", err)
		}
		if len(sprints) == 0 {
			return 0, fmt.Errorf("no future sprint found")
		}
		return sprints[0].ID, nil
	default:
		id, err := strconv.Atoi(arg)
		if err != nil {
			return 0, fmt.Errorf("invalid sprint ID %q: must be a number, 'active', or 'future'", arg)
		}
		return id, nil
	}
}
```

**Step 4: Use `resolveSprintID` in affected commands**

In `runSprintIssues`, replace:
```go
sprintID, err := strconv.Atoi(args[0])
if err != nil {
    return fmt.Errorf("invalid sprint ID: %w", err)
}
```
with:
```go
project, _ := cfg.GetProject(projectKey)  // move this up before resolveSprintID
// ...
sprintID, err := resolveSprintID(args[0], client, project.BoardID)
if err != nil {
    return fmt.Errorf("resolving sprint: %w", err)
}
```

Important: `resolveSprintID` needs `client` and `project.BoardID`, so the config/client setup must happen before the call. Reorder `runSprintIssues` so config load and client creation happen first, then `resolveSprintID`, then flag reads.

Apply the same pattern to:
- `runSprintComplete` — replace `strconv.Atoi(args[0])` with `resolveSprintID`
- `runSprintStart` — replace `strconv.Atoi(args[0])` with `resolveSprintID`
- `runSprintMove` — replace `strconv.Atoi(args[0])` for target ID with `resolveSprintID`

For the `--from-sprint` flag in `runSprintMove`, also resolve it:
```go
fromSprint, _ := cmd.Flags().GetString("from-sprint")
if fromSprint != "" {
    sourceID, err := resolveSprintID(fromSprint, client, project.BoardID)
    // ...
}
```

**Step 5: Run full test suite**

```bash
go test ./...
```
Expected: all pass

**Step 6: Commit**

```bash
git add internal/commands/sprint.go internal/commands/sprint_resolve_test.go
git commit -m "feat(sprint): add active/future keyword aliases for sprint IDs"
```

---

## Task 4: `sprint complete --move-to` and auto FUTURE→CLOSED handling

`sprint complete` currently fails with a 400 error when called on a FUTURE sprint (Jira requires ACTIVE→CLOSED). This task adds auto-handling (start then complete) and a `--move-to` flag that bulk-moves non-done issues before completing.

**Files:**
- Modify: `internal/commands/sprint.go` — `runSprintComplete` and `init()`
- Modify: `internal/api/sprint.go` — `CompleteSprint` to expose FUTURE→CLOSED helper OR handle it in command layer
- Test: `internal/commands/sprint_complete_test.go` (new)

**Step 1: Write failing tests**

Create `internal/commands/sprint_complete_test.go`:

```go
package commands_test

import (
	"testing"
	"github.com/user/jira-go/internal/commands"
)

func TestSprintCompleteHasMoveToFlag(t *testing.T) {
	cmd, _, err := commands.RootCmd.Find([]string{"sprint", "complete"})
	if err != nil || cmd == nil {
		t.Fatalf("sprint complete not found: %v", err)
	}
	if f := cmd.Flags().Lookup("move-to"); f == nil {
		t.Fatal("sprint complete --move-to flag not registered")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/commands/... -run TestSprintComplete -v
```
Expected: FAIL

**Step 3: Add `CompleteSprintSafe` to `internal/api/sprint.go`**

The Jira API only allows ACTIVE→CLOSED. If a sprint is FUTURE, we must first transition it to ACTIVE. Add a helper:

```go
// CompleteSprintSafe completes a sprint, automatically starting it first
// if it is currently in FUTURE state (Jira does not allow FUTURE→CLOSED directly).
func (c *Client) CompleteSprintSafe(sprintID int) error {
	sprint, err := c.GetSprint(sprintID)
	if err != nil {
		return fmt.Errorf("fetching sprint: %w", err)
	}
	if sprint.State == "future" {
		if err := c.StartSprint(sprintID, ""); err != nil {
			return fmt.Errorf("starting future sprint before close: %w", err)
		}
	}
	return c.CompleteSprint(sprintID)
}
```

**Step 4: Register `--move-to` flag and update `runSprintComplete`**

In `init()`, add:
```go
sprintCompleteCmd.Flags().String("move-to", "", "Move non-done issues to this sprint ID (or 'future') before completing")
```

Replace `runSprintComplete`:

```go
func runSprintComplete(cmd *cobra.Command, args []string) error {
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

	sprintID, err := resolveSprintID(args[0], client, project.BoardID)
	if err != nil {
		return fmt.Errorf("resolving sprint: %w", err)
	}

	// Handle --move-to: bulk move non-done issues before completing
	if moveTo, _ := cmd.Flags().GetString("move-to"); moveTo != "" {
		targetID, err := resolveSprintID(moveTo, client, project.BoardID)
		if err != nil {
			return fmt.Errorf("resolving --move-to sprint: %w", err)
		}

		jql := fmt.Sprintf("sprint = %d AND statusCategory != Done", sprintID)
		resp, err := client.SearchIssues(jql, 0, 500, project.MultiOwnerField, project.SprintField)
		if err != nil {
			return fmt.Errorf("fetching non-done issues: %w", err)
		}

		if len(resp.Issues) > 0 {
			keys := make([]string, len(resp.Issues))
			for i, issue := range resp.Issues {
				keys[i] = issue.Key
			}
			if err := client.MoveIssuesToSprint(targetID, keys); err != nil {
				return fmt.Errorf("moving issues: %w", err)
			}
			fmt.Printf("✓ Moved %d non-done issue(s) to sprint %d\n", len(keys), targetID)
		}
	}

	if err := client.CompleteSprintSafe(sprintID); err != nil {
		return fmt.Errorf("completing sprint: %w", err)
	}

	fmt.Printf("✓ Completed sprint %d\n", sprintID)
	return nil
}
```

**Step 5: Run tests**

```bash
go test ./internal/commands/... -run TestSprintComplete -v
go test ./...
```
Expected: all pass

**Step 6: Commit**

```bash
git add internal/commands/sprint.go internal/api/sprint.go internal/commands/sprint_complete_test.go
git commit -m "feat(sprint): add --move-to flag to sprint complete; auto-handle FUTURE→CLOSED"
```

---

## Task 5: Warn when `sprint_field` not configured in `report`

When `sprint_field` is not configured, `jira report` silently groups everything flat with no explanation. A one-line warning to stderr tells the user why sprint grouping isn't working.

**Files:**
- Modify: `internal/commands/report.go` — `runReport` (around line 84)
- Test: `internal/commands/report_sprint_warn_test.go` (new)

**Step 1: Write failing test**

Create `internal/commands/report_sprint_warn_test.go`:

```go
package commands_test

import (
	"testing"
	"github.com/user/jira-go/internal/commands"
)

func TestReportCommandExists(t *testing.T) {
	cmd, _, err := commands.RootCmd.Find([]string{"report"})
	if err != nil || cmd == nil {
		t.Fatalf("report command not found: %v", err)
	}
	// Structural test — warning logic requires integration test with mock config
	_ = cmd
}
```

**Step 2: Run test (should pass — just structural)**

```bash
go test ./internal/commands/... -run TestReportCommandExists -v
```

**Step 3: Add warning to `runReport` in `internal/commands/report.go`**

Find the block that sets `ownerField` and `sprintField` (around line 84):

```go
var ownerField, sprintField string
if project != nil {
    ownerField = project.MultiOwnerField
    sprintField = project.SprintField
}
```

Add a warning after it:

```go
var ownerField, sprintField string
if project != nil {
    ownerField = project.MultiOwnerField
    sprintField = project.SprintField
}
// Warn when sprint grouping won't work
if sprintFilter == "" && sprintField == "" {
    fmt.Fprintln(os.Stderr, "warn: sprint_field not configured for this project — showing flat assignee view instead of sprint grouping")
    fmt.Fprintln(os.Stderr, "      Run 'jira init' to configure the sprint custom field ID")
}
```

Note: `sprintFilter` is already read before this point in `runReport`. Ensure the warning is only emitted when no `--sprint` flag is set AND `sprintField` is empty — i.e. the user expected sprint grouping but won't get it.

**Step 4: Run full suite**

```bash
go test ./...
```
Expected: all pass

**Step 5: Commit**

```bash
git add internal/commands/report.go internal/commands/report_sprint_warn_test.go
git commit -m "fix(report): warn to stderr when sprint_field is not configured"
```

---

## Task 6: Update `jira-go:sprints` skill with new command capabilities

The skill file at `~/.claude/skills/jira-go:sprints.md` needs to document all improvements from Tasks 1–5.

**Files:**
- Modify: `~/.claude/skills/jira-go:sprints.md`

**Step 1: Read current skill file**

```bash
cat ~/.claude/skills/jira-go\:sprints.md
```

**Step 2: Add new sections**

Add after the existing "Move Issues to a Sprint" section:

```markdown
## Bulk Move Non-Done Issues Between Sprints

\`\`\`bash
# Move all non-done issues from sprint 45074 to sprint 46722
jira sprint move 46722 --from-sprint 45074

# Using keyword aliases
jira sprint move future --from-sprint active
\`\`\`

## Complete a Sprint and Move Incomplete Tickets

\`\`\`bash
# Complete active sprint, move non-done tickets to future sprint
jira sprint complete active --move-to future

# Complete a specific sprint, move to a specific sprint
jira sprint complete 45074 --move-to 46722
\`\`\`

Note: FUTURE sprints are automatically started then completed (Jira requires ACTIVE→CLOSED).

## Sprint ID Keyword Aliases

Most sprint commands accept `active` or `future` instead of a numeric sprint ID:

\`\`\`bash
jira sprint issues active
jira sprint issues future
jira sprint complete active --move-to future
jira sprint start future
\`\`\`
```

Update the "List Issues in a Sprint" section to add flags:

```markdown
## List Issues in a Sprint

\`\`\`bash
jira sprint issues active               # Active sprint, all issues
jira sprint issues 42 --status "Em andamento"  # Filter by status
jira sprint issues 42 --format json     # Machine-readable output
jira sprint issues 42 --limit 200       # Increase beyond default 100
\`\`\`
```

**Step 3: Verify file saved correctly**

```bash
head -5 ~/.claude/skills/jira-go\:sprints.md
```

**Step 4: Commit the skill update in the project repo**

The skill file lives outside the repo, so just confirm it's saved. No git commit needed for the skill file itself — but update AGENTS.md in the repo to reflect the new flags:

```bash
git add AGENTS.md  # if you updated it
git commit -m "docs: update AGENTS.md and sprint skill with new UX improvements"
```

If AGENTS.md doesn't need updating (the new commands are self-explanatory), commit can be skipped.

---

## Final Verification

**Step 1: Run full test suite**

```bash
go test ./...
```
Expected: all pass, no regressions

**Step 2: Build**

```bash
task build
```
Expected: `./build/jira` produced cleanly

**Step 3: Smoke test new flags**

```bash
./build/jira sprint issues --help    # should show --format, --status, --limit
./build/jira sprint move --help      # should show --from-sprint
./build/jira sprint complete --help  # should show --move-to
```

**Step 4: Final commit if any loose ends**

```bash
git log --oneline main..HEAD
```
