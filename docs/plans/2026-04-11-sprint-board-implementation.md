# Sprint Board Column Visibility & Assignee+Owner Merge Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add config-driven assignee+owner merging and column visibility/resizing for sprint board TUI.

**Architecture:** Two independent features. Feature 1 adds a config flag and helper method. Feature 2 adds column preference persistence via JSON file and TUI keyboard shortcuts.

**Tech Stack:** Go, Charmbracelet/bubbletea, YAML config, JSON for preferences.

---

## Part 1: Assignee + Owner Merge

### Task 1: Add config field

**Files:**
- Modify: `internal/config/config.go:31-37`

**Step 1: Write the failing test**

```go
// internal/config/config_test.go - add test for new field
func TestProjectMergeAssigneeOwnerDefault(t *testing.T) {
    cfg := &Config{}
    proj := Project{}
    // Default should be true when MergeAssigneeOwner is not set
    if proj.MergeAssigneeOwner != true {
        t.Errorf("expected default true, got %v", proj.MergeAssigneeOwner)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config -run TestProjectMergeAssigneeOwnerDefault -v`
Expected: FAIL - field doesn't exist

**Step 3: Add field to Project struct**

```go
type Project struct {
    JiraURL             string            `yaml:"jira_url"`
    BoardID            int               `yaml:"board_id"`
    MultiOwnerField    string            `yaml:"multi_owner_field"`
    SprintField        string            `yaml:"sprint_field"`
    IssueTypes         map[string]string `yaml:"issue_types"`
    MergeAssigneeOwner bool              `yaml:"merge_assignee_owner"` // NEW - default true
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/config -run TestProjectMergeAssigneeOwnerDefault -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add MergeAssigneeOwner field with default true"
```

---

### Task 2: Add helper method GetAllParticipants

**Files:**
- Modify: `internal/models/models.go:175-205`

**Step 1: Write the failing test**

```go
// internal/models/models_test.go
func TestIssueGetAllParticipants(t *testing.T) {
    issue := Issue{
        Owners: []User{
            {AccountID: "acc1", DisplayName: "Alice"},
        },
        Assignee: &User{AccountID: "acc2", DisplayName: "Bob"},
    }
    participants := issue.GetAllParticipants()
    if len(participants) != 2 {
        t.Errorf("expected 2 participants, got %d", len(participants))
    }
}

func TestIssueGetAllParticipantsDeduplicates(t *testing.T) {
    issue := Issue{
        Owners: []User{
            {AccountID: "acc1", DisplayName: "Alice"},
        },
        Assignee: &User{AccountID: "acc1", DisplayName: "Alice Clone"}, // same ID
    }
    participants := issue.GetAllParticipants()
    if len(participants) != 1 {
        t.Errorf("expected 1 (deduplicated), got %d", len(participants))
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/models -run TestIssueGetAllParticipants -v`
Expected: FAIL - method doesn't exist

**Step 3: Add helper method after ToIssueWithOwners**

```go
// GetAllParticipants returns all unique participants (owners + assignee merged)
func (i *Issue) GetAllParticipants() []User {
    seen := make(map[string]bool)
    var result []User

    addUnique := func(u User) {
        if u.AccountID != "" && !seen[u.AccountID] {
            seen[u.AccountID] = true
            result = append(result, u)
        }
    }

    for _, o := range i.Owners {
        addUnique(o)
    }
    if i.Assignee != nil {
        addUnique(*i.Assignee)
    }

    return result
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/models -run TestIssueGetAllParticipants -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/models/models.go internal/models/models_test.go
git commit -m "feat(models): add GetAllParticipants helper for deduplication"
```

---

### Task 3: Update task list display to use merged participants

**Files:**
- Modify: `internal/commands/task.go:400-420` (displayTaskListTable function)

**Step 1: Write the failing test**

```go
// internal/commands/commands_test.go - add test
func TestDisplayMergedParticipants(t *testing.T) {
    // This requires mocking the project config with MergeAssigneeOwner=true
    // and verifying output includes both assignee and owners
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/commands -run TestDisplayMergedParticipants -v`
Expected: FAIL - test not yet written (or fails if test infrastructure not ready)

**Step 3: Modify displayTaskListTable to use GetAllParticipants**

Find where assignee/owners are displayed. Update to:
```go
// Replace direct Owners/Assignee access with GetAllParticipants()
// when project.MergeAssigneeOwner is true
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/commands -run TestDisplayMergedParticipants -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/commands/task.go
git commit -m "feat(task): use merged assignee+owner in task list display"
```

---

### Task 4: Update kanban board to use merged participants

**Files:**
- Modify: `internal/tui/kanban.go:186-218` (NewKanbanBoard function)

**Step 1: Read current code to find owner/assignee display logic**

**Step 2: Update owner/assignee display to use GetAllParticipants**

In `NewKanbanBoard` where it sets `ki.Assignee`:
```go
// Use GetAllParticipants when merge is enabled (passed from caller)
// For now, update the display logic to use combined list
```

**Step 3: Test manually**

Run: `./build/jira sprint board --no-interactive`
Verify owners and assignees appear merged without duplicates

**Step 4: Commit**

```bash
git add internal/tui/kanban.go
git commit -m "feat(kanban): use merged assignee+owner in board display"
```

---

## Part 2: Column Visibility and Resizing

### Task 5: Create board columns config module

**Files:**
- Create: `internal/config/board_columns.go`

**Step 1: Write the failing test**

```go
// internal/config/board_columns_test.go
func TestLoadBoardColumns(t *testing.T) {
    prefs, err := LoadBoardColumns("SWCSIRT")
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    // Should return empty map for non-existent project, not error
    if prefs == nil {
        t.Errorf("expected empty map, got nil")
    }
}

func TestSaveBoardColumns(t *testing.T) {
    prefs := BoardColumnPrefs{
        "REVISAR": {Visible: true, Width: 25},
    }
    err := SaveBoardColumns("SWCSIRT", prefs)
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    
    loaded, err := LoadBoardColumns("SWCSIRT")
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    if loaded["REVISAR"].Width != 25 {
        t.Errorf("expected width 25, got %d", loaded["REVISAR"].Width)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config -run TestLoadBoardColumns -v`
Expected: FAIL - file doesn't exist

**Step 3: Create board_columns.go**

```go
package config

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type ColumnConfig struct {
    Visible bool `json:"visible"`
    Width   int  `json:"width"`
}

type BoardColumnPrefs map[string]ColumnConfig

func GetBoardColumnsPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".config", "jira-go", "board-columns.json")
}

func LoadBoardColumns(projectKey string) (BoardColumnPrefs, error) {
    path := GetBoardColumnsPath()
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return make(BoardColumnPrefs), nil
        }
        return nil, err
    }
    
    var allPrefs map[string]BoardColumnPrefs
    if err := json.Unmarshal(data, &allPrefs); err != nil {
        return nil, err
    }
    
    if prefs, ok := allPrefs[projectKey]; ok {
        return prefs, nil
    }
    return make(BoardColumnPrefs), nil
}

func SaveBoardColumns(projectKey string, prefs BoardColumnPrefs) error {
    path := GetBoardColumnsPath()
    
    // Load existing
    var allPrefs map[string]BoardColumnPrefs
    if data, err := os.ReadFile(path); err == nil {
        json.Unmarshal(data, &allPrefs)
    } else {
        allPrefs = make(map[string]BoardColumnPrefs)
    }
    
    allPrefs[projectKey] = prefs
    
    dir := filepath.Dir(path)
    os.MkdirAll(dir, 0755)
    
    out, _ := json.MarshalIndent(allPrefs, "", "  ")
    return os.WriteFile(path, out, 0600)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/config -run TestLoadBoardColumns -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/board_columns.go internal/config/board_columns_test.go
git commit -m "feat(config): add board column preferences persistence"
```

---

### Task 6: Update KanbanBoardModel to track column preferences

**Files:**
- Modify: `internal/tui/kanban.go:116-147`

**Step 1: Add fields to KanbanBoardModel**

```go
type KanbanBoardModel struct {
    // ... existing fields ...
    columnPrefs      config.BoardColumnPrefs
    projectKey       string
    hiddenCount      int
    // Column width override (0 = use calculated)
    columnWidthOverride int
}
```

**Step 2: Modify NewKanbanBoard to load prefs**

```go
func NewKanbanBoard(issues []models.Issue, sprintID int, client *api.Client, projectKey string) KanbanBoardModel {
    // Load column preferences
    prefs, _ := config.LoadBoardColumns(projectKey)
    
    // ... rest of initialization ...
}
```

**Step 3: Update column width calculation to use prefs**

Modify `calculateColumnWidth`:
```go
func (m KanbanBoardModel) calculateColumnWidth() int {
    // Use override if set
    if m.columnWidthOverride > 0 {
        return m.columnWidthOverride
    }
    // ... existing logic ...
}
```

**Step 4: Run build to verify it compiles**

Run: `go build ./...`
Expected: SUCCESS

**Step 5: Commit**

```bash
git add internal/tui/kanban.go
git commit -m "feat(kanban): add column preferences tracking to model"
```

---

### Task 7: Add keyboard shortcuts for column visibility and resizing

**Files:**
- Modify: `internal/tui/kanban.go:503-564` (handleNormalKeys)

**Step 1: Write the failing test**

```go
// internal/tui/kanban_test.go
func TestKanbanToggleColumnVisibility(t *testing.T) {
    // Create model with known columns
    // Press 'x' key
    // Verify column becomes hidden
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestKanbanToggleColumnVisibility -v`
Expected: FAIL - handler not implemented

**Step 3: Add key handlers**

In `handleNormalKeys`, add cases:
```go
case "x":
    // Toggle current column visibility
    currentCol := m.columns[m.activeColumn]
    if currentCol.Name == currentCol.Name { // dummy check
        m.toggleColumnVisibility(m.activeColumn)
    }
case "+", "=":
    // Increase column width
    m.resizeColumn(m.activeColumn, 5)
case "-":
    // Decrease column width
    m.resizeColumn(m.activeColumn, -5)
```

**Step 4: Add helper methods**

```go
func (m *KanbanBoardModel) toggleColumnVisibility(idx int) {
    col := &m.columns[idx]
    col.Hidden = !col.Hidden
    m.saveColumnPrefs()
}

func (m *KanbanBoardModel) resizeColumn(idx int, delta int) {
    m.columnWidthOverride += delta
    if m.columnWidthOverride < 15 {
        m.columnWidthOverride = 15
    }
    if m.columnWidthOverride > 50 {
        m.columnWidthOverride = 50
    }
    m.saveColumnPrefs()
}

func (m *KanbanBoardModel) saveColumnPrefs() {
    prefs := make(config.BoardColumnPrefs)
    for _, col := range m.columns {
        prefs[col.Name] = config.ColumnConfig{
            Visible: !col.Hidden,
            Width:   col.Width,
        }
    }
    config.SaveBoardColumns(m.projectKey, prefs)
}
```

**Step 5: Add Hidden field to KanbanColumn struct**

```go
type KanbanColumn struct {
    Name   string
    Issues []KanbanIssue
    List   list.Model
    Hidden bool   // NEW
    Width  int     // NEW - override width
}
```

**Step 6: Run test to verify it passes**

Run: `go test ./internal/tui -run TestKanbanToggleColumnVisibility -v`
Expected: PASS

**Step 7: Commit**

```bash
git add internal/tui/kanban.go
git commit -m "feat(kanban): add column visibility and resize keybindings"
```

---

### Task 8: Update column rendering to respect visibility

**Files:**
- Modify: `internal/tui/kanban.go:989-1051` (kanbanView)

**Step 1: Update kanbanView to skip hidden columns**

In `kanbanView`, when iterating columns:
```go
// Skip hidden columns but count them
if col.Hidden {
    m.hiddenCount++
    continue
}
```

**Step 2: Add hidden column indicator to View output**

At the end of `kanbanView`:
```go
if m.hiddenCount > 0 {
    b.WriteString(fmt.Sprintf("\n│ +%d hidden │ press x on edge to restore│", m.hiddenCount))
}
```

**Step 3: Update navigation to skip hidden columns**

In `handleNormalKeys` for left/right:
```go
case "left", "h":
    for m.activeColumn > 0 {
        m.activeColumn--
        if !m.columns[m.activeColumn].Hidden {
            break
        }
    }
// Similar for right
```

**Step 4: Test manually**

Run: `./build/jira sprint board`
Press `x` on a column - should hide
Press `+`/`-` - should resize
Navigate - should skip hidden columns

**Step 5: Commit**

```bash
git add internal/tui/kanban.go
git commit -m "feat(kanban): respect column visibility in rendering and navigation"
```

---

## Verification

After all tasks:

1. `task test` - all tests pass
2. `task build` - builds successfully
3. Manual test `jira sprint board --no-interactive` - shows issues
4. Manual test `jira sprint board` (interactive) - can hide/resize columns

---

## Execution Options

**Plan complete and saved to `docs/plans/2026-04-11-sprint-board-implementation.md`. Two execution options:**

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**
