// internal/tui/issue_list.go
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
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00C851"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")).
			Italic(true)

	assigneeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00A8E8"))
)

// Issue represents a simplified issue for TUI display
type Issue struct {
	Key           string
	Summary       string
	Type          string
	Status        string
	Assignee      string
	AssigneeEmail string
	Owners        []string
	Description   string
	Created       string
}

// IssueItem represents a list item for an issue
type IssueItem struct {
	issue Issue
}

func (i IssueItem) Title() string {
	return fmt.Sprintf("%s: %s", i.issue.Key, i.issue.Summary)
}

func (i IssueItem) Description() string {
	parts := []string{i.issue.Status}

	if i.issue.Assignee != "" {
		parts = append(parts, "👤 "+i.issue.Assignee)
	}

	if len(i.issue.Owners) > 0 {
		ownerStr := strings.Join(i.issue.Owners, ", ")
		if len(ownerStr) > 30 {
			ownerStr = ownerStr[:27] + "..."
		}
		parts = append(parts, "👥 "+ownerStr)
	}

	return strings.Join(parts, " • ")
}

func (i IssueItem) FilterValue() string {
	return i.issue.Key + " " + i.issue.Summary + " " + i.issue.Assignee
}

// IssueListModel is the TUI model for listing issues
type IssueListModel struct {
	list         list.Model
	issues       []Issue
	client       *api.Client
	projectKey   string
	ownerFieldID string
	viewing      bool
	selected     Issue
	width        int
	height       int
}

// NewIssueList creates a new issue list TUI
func NewIssueList(issues []models.Issue, client *api.Client, projectKey, ownerFieldID string) IssueListModel {
	var items []list.Item
	var tuiIssues []Issue

	for _, issue := range issues {
		tuiIssue := Issue{
			Key:     issue.Key,
			Summary: issue.Summary,
			Type:    issue.Type,
			Status:  issue.Status,
			Created: issue.Created.Format("2006-01-02 15:04"),
		}

		if issue.Assignee != nil {
			tuiIssue.Assignee = issue.Assignee.DisplayName
			tuiIssue.AssigneeEmail = issue.Assignee.Email
		}

		for _, owner := range issue.Owners {
			tuiIssue.Owners = append(tuiIssue.Owners, owner.DisplayName)
		}

		tuiIssues = append(tuiIssues, tuiIssue)
		items = append(items, IssueItem{issue: tuiIssue})
	}

	// Custom delegate to show more info
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.NormalDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))

	l := list.New(items, delegate, 100, 20)
	l.Title = fmt.Sprintf("Jira Issues - %s", projectKey)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.SetShowHelp(true)

	return IssueListModel{
		list:         l,
		issues:       tuiIssues,
		client:       client,
		projectKey:   projectKey,
		ownerFieldID: ownerFieldID,
		viewing:      false,
	}
}

func (m IssueListModel) Init() tea.Cmd {
	return nil
}

func (m IssueListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle back from detail view
		if m.viewing {
			if msg.String() == "q" || msg.String() == "esc" || msg.String() == "backspace" {
				m.viewing = false
				return m, nil
			}
			return m, nil
		}

		// Handle list view keys
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "enter", "o":
			if item, ok := m.list.SelectedItem().(IssueItem); ok {
				m.selected = item.issue
				// Fetch full issue details
				if m.client != nil {
					fullIssue, err := m.client.GetIssue(item.issue.Key, m.ownerFieldID)
					if err == nil && fullIssue != nil {
						m.selected.Description = fullIssue.Description
						if fullIssue.Assignee != nil {
							m.selected.Assignee = fullIssue.Assignee.DisplayName
							m.selected.AssigneeEmail = fullIssue.Assignee.Email
						}
						m.selected.Owners = []string{}
						for _, o := range fullIssue.Owners {
							m.selected.Owners = append(m.selected.Owners, fmt.Sprintf("%s (%s)", o.DisplayName, o.Email))
						}
					}
				}
				m.viewing = true
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, m.height-4)
	}

	var cmd tea.Cmd
	if !m.viewing {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m IssueListModel) View() string {
	if m.viewing {
		return m.detailView()
	}
	return m.list.View() + "\n" + helpStyle.Render("↑/↓: navigate • enter/o: open • /: filter • q: quit")
}

func (m IssueListModel) detailView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf(" %s ", m.selected.Key)))
	b.WriteString("\n\n")
	b.WriteString(selectedStyle.Render(m.selected.Summary))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Type: %s\n", m.selected.Type))
	b.WriteString(fmt.Sprintf("Status: %s\n", m.selected.Status))

	if m.selected.Assignee != "" {
		b.WriteString(assigneeStyle.Render(fmt.Sprintf("👤 Assignee: %s", m.selected.Assignee)))
		if m.selected.AssigneeEmail != "" {
			b.WriteString(fmt.Sprintf(" (%s)", m.selected.AssigneeEmail))
		}
		b.WriteString("\n")
	} else {
		b.WriteString("👤 Assignee: Unassigned\n")
	}

	if len(m.selected.Owners) > 0 {
		b.WriteString(fmt.Sprintf("👥 Owners: %s\n", strings.Join(m.selected.Owners, ", ")))
	}

	b.WriteString(fmt.Sprintf("Created: %s\n", m.selected.Created))
	b.WriteString("\n" + strings.Repeat("─", 60) + "\n")
	b.WriteString("Description:\n\n")

	if m.selected.Description != "" {
		b.WriteString(m.selected.Description)
	} else {
		b.WriteString(helpStyle.Render("(no description)"))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("esc/backspace: back • q: quit"))

	return b.String()
}

func (m IssueListModel) GetTitle() string {
	return "Issue List"
}

// GetSelected returns the currently selected issue
func (m IssueListModel) GetSelected() *Issue {
	if m.viewing {
		return &m.selected
	}
	if item, ok := m.list.SelectedItem().(IssueItem); ok {
		return &item.issue
	}
	return nil
}
