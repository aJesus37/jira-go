# Sprint Board Column Visibility & Assignee+Owner Merge Design

## Overview

Two related improvements to the Jira CLI:
1. Assignee and Owner fields should be mergeable via config
2. Sprint board columns should be hideable and resizable

---

## Feature 1: Assignee + Owner Merge

### Config Change

Add to `Project` struct in `internal/config/config.go`:

```yaml
projects:
  PROJ:
    merge_assignee_owner: true  # default true
```

**Default is `true`** (as requested by user).

### Implementation

1. **Model change** - Add `MergeAssigneeOwner bool` to `Project` struct

2. **Helper method** on `Issue` in `internal/models/models.go`:
```go
func (i *Issue) GetAllParticipants() []User {
    seen := make(map[string]bool)
    var result []User
    
    addUnique := func(u User) {
        if u.AccountID != "" && !seen[u.AccountID] {
            seen[u.AccountID] = true
            result = append(result, u)
        }
    }
    
    // Add from Owners
    for _, o := range i.Owners {
        addUnique(o)
    }
    // Add from Assignee
    if i.Assignee != nil {
        addUnique(*i.Assignee)
    }
    
    return result
}
```

3. **Update all UI code** that displays owners/assignees to use this helper instead of checking `Owners` alone:
- `internal/tui/kanban.go`
- `internal/commands/task.go` (displayTaskListTable)
- Any other place showing assignee/owners

4. **Conditional merge** - Only merge if `project.MergeAssigneeOwner == true`

### Affected Files

- `internal/config/config.go` - add field
- `internal/models/models.go` - add helper method
- `internal/tui/kanban.go` - update display
- `internal/commands/task.go` - update displayTaskListTable
- `internal/tui/issue_list.go` if applicable

---

## Feature 2: Column Visibility and Resizing

### Storage

File: `~/.config/jira-go/board-columns.json`

```json
{
  "SWCSIRT": {
    "columns": {
      "REVISAR":     { "visible": true, "width": 25 },
      "EM ANDAMENTO": { "visible": true, "width": 30 },
      "CONCLUÍDO":   { "visible": false, "width": 20 }
    }
  }
}
```

### Config Change

No config.yaml changes - uses separate JSON file for persistence.

### TUI Interactions

| Key | Action |
|-----|--------|
| `x` | Toggle visibility of current column |
| `+` / `=` | Increase current column width |
| `-` | Decrease current column width |
| `h` / `←` | Navigate to previous column (skips hidden) |
| `l` / `→` | Navigate to next column (skips hidden) |

### Hidden Column Indicator

When columns are hidden, render a small "collapsed edge" indicator:
```
║ +3 hidden │ ← press x on edge to restore
```

Clicking/pressing on this indicator or using a shortcut restores the column picker.

### Implementation

1. **New file** `internal/config/board_columns.go`:
```go
type ColumnConfig struct {
    Visible bool `json:"visible"`
    Width   int  `json:"width"`
}

type BoardColumnPrefs map[string]ColumnConfig  // status -> config

func LoadBoardColumns(projectKey string) (BoardColumnPrefs, error)
func SaveBoardColumns(projectKey string, prefs BoardColumnPrefs) error
```

2. **Model change** in `KanbanBoardModel`:
```go
type KanbanBoardModel struct {
    // ... existing fields
    columnPrefs BoardColumnPrefs
    hiddenCount int  // Track how many columns hidden
}
```

3. **Key handling** in `handleNormalKeys`:
- Add cases for `x`, `+`, `-`
- Call methods to toggle width/save prefs

4. **Column rendering**:
- Skip columns where `visible == false`
- Use saved `width` instead of calculated width
- Track `hiddenCount` for indicator

5. **Auto-save** - Call `SaveBoardColumns` whenever prefs change

### Affected Files

- `internal/config/board_columns.go` (new)
- `internal/config/config.go` (ensure dirs exist)
- `internal/tui/kanban.go` - major changes to column handling

---

## Testing Considerations

1. Test merge doesn't duplicate when same user is in both assignee and owners
2. Test hidden columns don't appear but navigation still works
3. Test column width bounds (min 15, max 50)
4. Test that new columns (unknown statuses) get default width and are visible

---

## Migration

For existing configs, `merge_assignee_owner` defaults to `true`, so no migration needed.

For board column prefs, missing columns get sensible defaults (visible, auto width).
