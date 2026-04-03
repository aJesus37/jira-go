// internal/tui/issue_list.go
package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/jira-go/internal/models"
)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)
)

// IssueItem represents a list item for an issue
type IssueItem struct {
	Issue models.Issue
}

func (i IssueItem) Title() string {
	return fmt.Sprintf("%s: %s", i.Issue.Key, i.Issue.Summary)
}

func (i IssueItem) Description() string {
	status := i.Issue.Status
	if i.Issue.Assignee != nil {
		status += " • " + i.Issue.Assignee.DisplayName
	}
	return status
}

func (i IssueItem) FilterValue() string {
	return i.Issue.Key + " " + i.Issue.Summary
}

// IssueListModel is the TUI model for listing issues
type IssueListModel struct {
	list   list.Model
	issues []models.Issue
	err    error
}

// NewIssueList creates a new issue list TUI
func NewIssueList(issues []models.Issue) IssueListModel {
	var items []list.Item
	for _, issue := range issues {
		items = append(items, IssueItem{Issue: issue})
	}

	l := list.New(items, list.NewDefaultDelegate(), 80, 20)
	l.Title = "Jira Issues"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return IssueListModel{
		list:   l,
		issues: issues,
	}
}

func (m IssueListModel) Init() tea.Cmd {
	return nil
}

func (m IssueListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m IssueListModel) View() string {
	return m.list.View()
}

func (m IssueListModel) GetTitle() string {
	return "Issue List"
}

// GetSelected returns the currently selected issue
func (m IssueListModel) GetSelected() *models.Issue {
	if item, ok := m.list.SelectedItem().(IssueItem); ok {
		return &item.Issue
	}
	return nil
}
