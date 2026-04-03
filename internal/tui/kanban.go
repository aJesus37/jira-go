// internal/tui/kanban.go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/jira-go/internal/api"
	"github.com/user/jira-go/internal/models"
)

var (
	columnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1).
			Width(30)

	selectedColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00C851")).
				Padding(1).
				Width(30)

	todoColumnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#808080")).
			Padding(1).
			Width(30)

	inProgressColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00A8E8")).
				Padding(1).
				Width(30)

	doneColumnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00C851")).
			Padding(1).
			Width(30)
)

// KanbanIssue represents an issue in kanban view
type KanbanIssue struct {
	Key      string
	Summary  string
	Assignee string
	Status   string
	Type     string
}

// KanbanColumn represents a column in the kanban board
type KanbanColumn struct {
	Name   string
	Issues []KanbanIssue
	List   list.Model
}

// KanbanBoardModel is the TUI model for kanban board
type KanbanBoardModel struct {
	columns      []KanbanColumn
	activeColumn int
	client       *api.Client
	sprintID     int
	width        int
	height       int
}

// NewKanbanBoard creates a new kanban board TUI
func NewKanbanBoard(issues []models.Issue, sprintID int, client *api.Client) KanbanBoardModel {
	// Group issues by status
	statusMap := make(map[string][]KanbanIssue)

	for _, issue := range issues {
		ki := KanbanIssue{
			Key:     issue.Key,
			Summary: issue.Summary,
			Status:  issue.Status,
			Type:    issue.Type,
		}

		if issue.Assignee != nil {
			ki.Assignee = issue.Assignee.DisplayName
		}

		statusMap[issue.Status] = append(statusMap[issue.Status], ki)
	}

	// Create columns for common statuses
	columnOrder := []string{"To Do", "In Progress", "Done"}
	var columns []KanbanColumn

	for _, status := range columnOrder {
		issues := statusMap[status]
		var items []list.Item

		for _, issue := range issues {
			items = append(items, KanbanIssueItem{issue: issue})
		}

		l := list.New(items, list.NewDefaultDelegate(), 25, 15)
		l.Title = fmt.Sprintf("%s (%d)", status, len(issues))
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)

		columns = append(columns, KanbanColumn{
			Name:   status,
			Issues: issues,
			List:   l,
		})
	}

	return KanbanBoardModel{
		columns:      columns,
		activeColumn: 0,
		client:       client,
		sprintID:     sprintID,
	}
}

// KanbanIssueItem represents a list item in kanban
type KanbanIssueItem struct {
	issue KanbanIssue
}

func (i KanbanIssueItem) Title() string {
	return fmt.Sprintf("%s: %s", i.issue.Key, i.issue.Summary)
}

func (i KanbanIssueItem) Description() string {
	if i.issue.Assignee != "" {
		return "👤 " + i.issue.Assignee
	}
	return "Unassigned"
}

func (i KanbanIssueItem) FilterValue() string {
	return i.issue.Key + " " + i.issue.Summary
}

func (m KanbanBoardModel) Init() tea.Cmd {
	return nil
}

func (m KanbanBoardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "left", "h":
			if m.activeColumn > 0 {
				m.activeColumn--
			}
		case "right", "l":
			if m.activeColumn < len(m.columns)-1 {
				m.activeColumn++
			}
		case "up", "k":
			if len(m.columns) > 0 {
				var cmd tea.Cmd
				m.columns[m.activeColumn].List, cmd = m.columns[m.activeColumn].List.Update(msg)
				return m, cmd
			}
		case "down", "j":
			if len(m.columns) > 0 {
				var cmd tea.Cmd
				m.columns[m.activeColumn].List, cmd = m.columns[m.activeColumn].List.Update(msg)
				return m, cmd
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m KanbanBoardModel) View() string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Sprint Board - Sprint %d", m.sprintID)))
	b.WriteString("\n\n")

	// Render columns side by side
	var columnViews []string
	for i, col := range m.columns {
		style := columnStyle
		if i == m.activeColumn {
			// Highlight active column
			switch col.Name {
			case "To Do":
				style = todoColumnStyle
			case "In Progress":
				style = inProgressColumnStyle
			case "Done":
				style = doneColumnStyle
			default:
				style = selectedColumnStyle
			}
		} else {
			// Dim inactive columns
			switch col.Name {
			case "To Do":
				style = todoColumnStyle
			case "In Progress":
				style = inProgressColumnStyle
			case "Done":
				style = doneColumnStyle
			}
		}

		colView := style.Render(col.List.View())
		columnViews = append(columnViews, colView)
	}

	// Join columns horizontally
	row := lipgloss.JoinHorizontal(lipgloss.Top, columnViews...)
	b.WriteString(row)

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("←/→: switch columns • ↑/↓: navigate • q: quit"))

	return b.String()
}

func (m KanbanBoardModel) GetTitle() string {
	return "Kanban Board"
}
