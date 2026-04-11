// internal/tui/kanban.go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/jira-go/internal/api"
	"github.com/user/jira-go/internal/config"
	"github.com/user/jira-go/internal/models"
)

var (
	columnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#555555")).
			Padding(0, 1)

	activeColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	todoColumnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#808080")).
			Padding(0, 1)

	todoActiveColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	inProgressColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00A8E8")).
				Padding(0, 1)

	inProgressActiveColumnStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#7D56F4")).
					Padding(0, 1)

	doneColumnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00C851")).
			Padding(0, 1)

	doneActiveColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	blockedColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#FF4444")).
				Padding(0, 1)

	blockedActiveColumnStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#7D56F4")).
					Padding(0, 1)

	reviewColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#FFA500")).
				Padding(0, 1)

	reviewActiveColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(2, 4).
			Width(50)

	dimmedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))
)

// KanbanMode represents the current mode of the kanban board
type KanbanMode int

const (
	ModeKanbanNormal KanbanMode = iota
	ModeKanbanStatusChange
	ModeKanbanAddComment
	ModeKanbanAssigneeChange
	ModeKanbanSprintAssign
	ModeKanbanTaskDetail
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
	Hidden bool
	Width  int
}

// KanbanBoardModel is the TUI model for kanban board
type KanbanBoardModel struct {
	columns      []KanbanColumn
	activeColumn int
	client       *api.Client
	sprintID     int
	width        int
	height       int

	columnPrefs         config.BoardColumnPrefs
	projectKey          string
	hiddenCount         int
	columnWidthOverride int

	// Action mode
	mode          KanbanMode
	selectedIssue KanbanIssue
	loading       bool
	message       string

	// Status change
	transitions     []api.Transition
	transitionIndex int
	originalStatus  string // Store original status for rollback on error

	// Sprint assignment
	sprints     []models.Sprint
	sprintIndex int

	// Comment
	commentInput textarea.Model

	// Task detail
	detailIssue  *models.Issue
	detailVP     viewport.Model
	detailOffset int

	// Column visibility focus
	focusHiddenColumns bool
}

// generateMockTasks creates realistic mock tasks for demonstration
func generateMockTasks() []KanbanIssue {
	mockAssignees := []string{"Alice Chen", "Bob Smith", "Carol Jones", "David Lee"}
	mockTasks := []struct {
		summary string
		status  string
		type_   string
	}{
		{"Setup project repository and CI/CD pipeline", "Done", "Task"},
		{"Design database schema for user management", "Done", "Story"},
		{"Implement user authentication API", "In Progress", "Story"},
		{"Create login page UI components", "In Progress", "Task"},
		{"Write unit tests for auth service", "To Do", "Task"},
		{"Configure OAuth providers (Google, GitHub)", "To Do", "Story"},
		{"Implement password reset flow", "To Do", "Story"},
		{"Add email verification system", "To Do", "Task"},
		{"Create user profile dashboard", "Blocked", "Story"},
		{"Set up monitoring and alerting", "Review", "Task"},
		{"Performance optimization for API endpoints", "Done", "Task"},
		{"Document API endpoints", "In Progress", "Task"},
		{"Implement rate limiting", "To Do", "Story"},
		{"Add caching layer with Redis", "To Do", "Task"},
	}

	var result []KanbanIssue
	for i, task := range mockTasks {
		result = append(result, KanbanIssue{
			Key:      fmt.Sprintf("MOCK-%d", i+1),
			Summary:  task.summary,
			Status:   task.status,
			Type:     task.type_,
			Assignee: mockAssignees[i%len(mockAssignees)],
		})
	}
	return result
}

// NewKanbanBoard creates a new kanban board TUI
func NewKanbanBoard(issues []models.Issue, sprintID int, client *api.Client, projectKey string) KanbanBoardModel {
	// Load column preferences
	prefs, _ := config.LoadBoardColumns(projectKey)

	// If no real issues, use mock data
	var kanbanIssues []KanbanIssue
	if len(issues) == 0 {
		kanbanIssues = generateMockTasks()
	} else {
		for _, issue := range issues {
			ki := KanbanIssue{
				Key:     issue.Key,
				Summary: issue.Summary,
				Status:  issue.Status,
				Type:    issue.Type,
			}

			participants := issue.GetAllParticipants()
			if len(participants) > 0 {
				var names []string
				for _, p := range participants {
					if p.DisplayName != "" {
						names = append(names, p.DisplayName)
					}
				}
				ki.Assignee = strings.Join(names, ", ")
			} else if issue.Assignee != nil {
				ki.Assignee = issue.Assignee.DisplayName
			}

			kanbanIssues = append(kanbanIssues, ki)
		}
	}

	// Group issues by status
	statusMap := make(map[string][]KanbanIssue)
	for _, ki := range kanbanIssues {
		statusMap[ki.Status] = append(statusMap[ki.Status], ki)
	}

	// Create columns for whatever statuses exist in the issues
	statusOrder := []string{"To Do", "In Progress", "Review", "Blocked", "Done"} // Fallback order
	var columns []KanbanColumn

	// First pass: collect all unique statuses
	allStatuses := make(map[string]bool)
	for status := range statusMap {
		allStatuses[status] = true
	}

	// Build column order: known statuses first in preferred order, then any others alphabetically
	var columnOrder []string
	knownStatuses := map[string]bool{"To Do": true, "In Progress": true, "Review": true, "Blocked": true, "Done": true}
	for _, status := range statusOrder {
		if allStatuses[status] {
			columnOrder = append(columnOrder, status)
		}
	}
	for status := range allStatuses {
		if !knownStatuses[status] {
			columnOrder = append(columnOrder, status)
		}
	}

	for _, status := range columnOrder {
		statusIssues := statusMap[status]
		var items []list.Item

		for _, issue := range statusIssues {
			items = append(items, KanbanIssueItem{issue: issue})
		}

		delegate := list.NewDefaultDelegate()
		// Configure to show single line items that expand/contract with column width
		delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Copy().MaxWidth(0).Height(1)
		delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Copy().MaxWidth(0).Height(1)
		delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Copy().MaxWidth(0).Height(1)
		delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Copy().MaxWidth(0).Height(1)
		l := list.New(items, delegate, 35, 18)
		l.Title = fmt.Sprintf("%s (%d)", status, len(statusIssues))
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.SetShowHelp(false)

		hidden := false
		width := 0
		if colConfig, ok := prefs[status]; ok {
			if !colConfig.Visible {
				hidden = true
			}
			if colConfig.Width > 0 {
				width = colConfig.Width
			}
		}

		columns = append(columns, KanbanColumn{
			Name:   status,
			Issues: statusIssues,
			List:   l,
			Hidden: hidden,
			Width:  width,
		})

		// Apply saved width to list size if set (use shorter height to fit screen)
		if width > 0 {
			l.SetSize(width, 12)
		}
	}

	// Initialize comment input with light text for dark background
	commentInput := textarea.New()
	commentInput.Placeholder = "Enter your comment..."
	commentInput.SetWidth(60)
	commentInput.SetHeight(2)
	commentInput.Focus()
	// Disable line numbers and use no prompt for cleaner look
	commentInput.ShowLineNumbers = false
	commentInput.Prompt = ""

	// Configure text colors for dark terminal background (no explicit background)
	commentInput.FocusedStyle.Base = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	commentInput.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	commentInput.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	commentInput.BlurredStyle.Base = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	commentInput.BlurredStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	commentInput.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	return KanbanBoardModel{
		columns:             columns,
		activeColumn:        0,
		client:              client,
		sprintID:            sprintID,
		mode:                ModeKanbanNormal,
		commentInput:        commentInput,
		transitionIndex:     0,
		sprintIndex:         0,
		columnPrefs:         prefs,
		projectKey:          projectKey,
		columnWidthOverride: 0,
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

// Message types for async operations
type kanbanTransitionsLoadedMsg struct {
	transitions []api.Transition
	err         error
}

type kanbanActionCompletedMsg struct {
	action string
	err    error
}

type kanbanSprintsLoadedMsg struct {
	sprints []models.Sprint
	err     error
}

type kanbanIssueRefreshedMsg struct {
	issue *models.Issue
	err   error
}

type kanbanIssueDetailsLoadedMsg struct {
	issue *models.Issue
	err   error
}

func (m KanbanBoardModel) Init() tea.Cmd {
	return nil
}

// Async functions
func (m KanbanBoardModel) loadTransitions(key string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return kanbanTransitionsLoadedMsg{err: fmt.Errorf("no client")}
		}
		transitions, err := m.client.GetTransitions(key)
		return kanbanTransitionsLoadedMsg{transitions: transitions, err: err}
	}
}

func (m KanbanBoardModel) transitionIssue(key, transitionID string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return kanbanActionCompletedMsg{action: "transition", err: fmt.Errorf("no client")}
		}
		err := m.client.TransitionIssue(key, transitionID)
		return kanbanActionCompletedMsg{action: "transition", err: err}
	}
}

func (m KanbanBoardModel) addComment(key, body string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return kanbanActionCompletedMsg{action: "comment", err: fmt.Errorf("no client")}
		}
		err := m.client.AddComment(key, body)
		return kanbanActionCompletedMsg{action: "comment", err: err}
	}
}

func (m KanbanBoardModel) loadSprints() tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return kanbanSprintsLoadedMsg{err: fmt.Errorf("no client")}
		}
		// Get board ID from sprint - we need to fetch board info
		// For now, we'll use the current sprint's board
		sprints, err := m.client.GetOpenSprints(0) // This won't work without board ID
		return kanbanSprintsLoadedMsg{sprints: sprints, err: err}
	}
}

func (m KanbanBoardModel) refreshIssue(key string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return kanbanIssueRefreshedMsg{err: fmt.Errorf("no client")}
		}
		issue, err := m.client.GetIssue(key, "", "")
		return kanbanIssueRefreshedMsg{issue: issue, err: err}
	}
}

func (m KanbanBoardModel) assignToSprint(sprintID int, issueKey string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return kanbanActionCompletedMsg{action: "sprint", err: fmt.Errorf("no client")}
		}
		err := m.client.MoveIssuesToSprint(sprintID, []string{issueKey})
		return kanbanActionCompletedMsg{action: "sprint", err: err}
	}
}

func (m KanbanBoardModel) loadIssueDetails(key string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return kanbanIssueDetailsLoadedMsg{err: fmt.Errorf("no client")}
		}
		// Fetch issue with description
		issue, err := m.client.GetIssue(key, "description", "")
		return kanbanIssueDetailsLoadedMsg{issue: issue, err: err}
	}
}

func (m *KanbanBoardModel) toggleColumnVisibility(idx int) {
	if idx < 0 || idx >= len(m.columns) {
		return
	}
	col := &m.columns[idx]
	col.Hidden = !col.Hidden
	m.saveColumnPrefs()
}

func (m *KanbanBoardModel) resizeColumn(idx int, delta int) {
	if idx < 0 || idx >= len(m.columns) {
		return
	}
	// Don't resize hidden columns
	if m.columns[idx].Hidden {
		return
	}
	m.columns[idx].Width += delta
	if m.columns[idx].Width < 15 {
		m.columns[idx].Width = 15
	}
	if m.columns[idx].Width > 50 {
		m.columns[idx].Width = 50
	}
	// Update list size to reflect new width
	width := m.columns[idx].Width
	if width < 20 {
		width = 20
	}
	// Cap list height to fit in terminal (same logic as WindowSizeMsg)
	listHeight := m.height - 10
	if listHeight < 8 {
		listHeight = 8
	}
	if listHeight > 20 {
		listHeight = 20
	}
	m.columns[idx].List.SetSize(width, listHeight)
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

func (m KanbanBoardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case kanbanTransitionsLoadedMsg:
		m.loading = false
		if msg.err == nil {
			m.transitions = msg.transitions
			m.transitionIndex = 0
			m.mode = ModeKanbanStatusChange
		}
		return m, nil

	case kanbanActionCompletedMsg:
		if msg.err == nil {
			m.message = fmt.Sprintf("✓ %s completed", msg.action)
			// Board was already updated optimistically, just clear loading
		} else {
			m.message = fmt.Sprintf("✗ %s failed: %v", msg.action, msg.err)
			// Revert optimistic update on error
			if msg.action == "transition" && m.originalStatus != "" {
				m.updateIssueStatus(m.selectedIssue.Key, m.originalStatus)
				m.originalStatus = "" // Clear after revert
			}
		}
		m.loading = false
		return m, nil

	case kanbanIssueRefreshedMsg:
		m.loading = false
		if msg.err == nil && msg.issue != nil {
			// Update the issue in the appropriate column
			m.updateIssueInBoard(msg.issue)
		}
		m.mode = ModeKanbanNormal
		return m, nil

	case kanbanSprintsLoadedMsg:
		m.loading = false
		if msg.err == nil {
			m.sprints = msg.sprints
			m.sprintIndex = 0
			m.mode = ModeKanbanSprintAssign
		} else {
			m.message = fmt.Sprintf("✗ Failed to load sprints: %v", msg.err)
			m.mode = ModeKanbanNormal
		}
		return m, nil

	case kanbanIssueDetailsLoadedMsg:
		m.loading = false
		if msg.err == nil && msg.issue != nil {
			m.detailIssue = msg.issue
			// Initialize the viewport for scrolling
			contentHeight := m.height - 20 // Reserve space for header and footer
			if contentHeight < 10 {
				contentHeight = 10
			}
			m.detailVP = viewport.New(m.width-28, contentHeight)
			m.detailVP.SetContent(m.taskDetailContent())
			m.mode = ModeKanbanTaskDetail
		} else {
			m.message = fmt.Sprintf("✗ Failed to load issue details: %v", msg.err)
			m.mode = ModeKanbanNormal
		}
		return m, nil

	case tea.KeyMsg:
		// Handle f key for focus toggle at top level before mode switch
		keyStr := msg.String()
		if keyStr == "f" || keyStr == "F" {
			if m.hiddenCount > 0 {
				m.focusHiddenColumns = !m.focusHiddenColumns
				m.message = ""
				if m.focusHiddenColumns {
					m.message = "Hidden columns focused"
				} else {
					m.message = "Visible columns focused"
				}
			}
			return m, nil
		}

		// Handle different modes
		switch m.mode {
		case ModeKanbanStatusChange:
			return m.handleStatusChangeKeys(msg)
		case ModeKanbanAddComment:
			return m.handleAddCommentKeys(msg)
		case ModeKanbanAssigneeChange:
			return m.handleAssigneeChangeKeys(msg)
		case ModeKanbanSprintAssign:
			return m.handleSprintAssignKeys(msg)
		case ModeKanbanTaskDetail:
			return m.handleTaskDetailKeys(msg)
		case ModeKanbanNormal:
			return m.handleNormalKeys(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update column widths - use custom width if set, otherwise calculate
		if len(m.columns) > 0 {
			columnWidth := m.calculateColumnWidth()
			// Cap list height to fit in terminal (leave room for header, message, hidden cols, help)
			// Header ~3 lines, message ~2 lines, hidden cols row ~2 lines, help ~1 line = ~8 lines
			listHeight := m.height - 10
			if listHeight < 8 {
				listHeight = 8
			}
			if listHeight > 20 {
				listHeight = 20
			}
			for i := range m.columns {
				width := columnWidth
				if m.columns[i].Width > 0 {
					width = m.columns[i].Width
				}
				m.columns[i].List.SetSize(width, listHeight)
			}
		}
	}

	return m, nil
}

func (m KanbanBoardModel) handleNormalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle 'f' key for focus toggle (check before switch to ensure it works)
	keyStr := msg.String()
	if keyStr == "f" || keyStr == "F" {
		if m.hiddenCount > 0 {
			m.focusHiddenColumns = !m.focusHiddenColumns
			m.message = ""
			if m.focusHiddenColumns {
				m.message = "Hidden columns focused"
			} else {
				m.message = "Visible columns focused"
			}
		}
		return m, nil
	}

	switch keyStr {
	case "q", "esc":
		return m, tea.Quit
	case "left", "h":
		// Move to previous column (respecting focus mode)
		if m.focusHiddenColumns {
			// Navigate only through hidden columns
			for {
				if m.activeColumn > 0 {
					m.activeColumn--
				} else {
					m.activeColumn = len(m.columns) - 1
				}
				if m.columns[m.activeColumn].Hidden {
					break
				}
			}
		} else {
			// Navigate only through visible columns
			for {
				if m.activeColumn > 0 {
					m.activeColumn--
				} else {
					m.activeColumn = len(m.columns) - 1
				}
				if !m.columns[m.activeColumn].Hidden {
					break
				}
			}
		}
	case "right", "l":
		// Move to next column (respecting focus mode)
		if m.focusHiddenColumns {
			// Navigate only through hidden columns
			for {
				if m.activeColumn < len(m.columns)-1 {
					m.activeColumn++
				} else {
					m.activeColumn = 0
				}
				if m.columns[m.activeColumn].Hidden {
					break
				}
			}
		} else {
			// Navigate only through visible columns
			for {
				if m.activeColumn < len(m.columns)-1 {
					m.activeColumn++
				} else {
					m.activeColumn = 0
				}
				if !m.columns[m.activeColumn].Hidden {
					break
				}
			}
		}
	case "up", "k":
		// Only navigate within visible columns
		if len(m.columns) > 0 && !m.columns[m.activeColumn].Hidden {
			var cmd tea.Cmd
			m.columns[m.activeColumn].List, cmd = m.columns[m.activeColumn].List.Update(msg)
			return m, cmd
		}
	case "down", "j":
		// Only navigate within visible columns
		if len(m.columns) > 0 && !m.columns[m.activeColumn].Hidden {
			var cmd tea.Cmd
			m.columns[m.activeColumn].List, cmd = m.columns[m.activeColumn].List.Update(msg)
			return m, cmd
		}
	case "s":
		// Change status
		if item, ok := m.columns[m.activeColumn].List.SelectedItem().(KanbanIssueItem); ok {
			m.selectedIssue = item.issue
			m.loading = true
			m.message = ""
			return m, m.loadTransitions(item.issue.Key)
		}
	case "c":
		// Add comment
		if item, ok := m.columns[m.activeColumn].List.SelectedItem().(KanbanIssueItem); ok {
			m.selectedIssue = item.issue
			m.mode = ModeKanbanAddComment
			m.commentInput.SetValue("")
			m.commentInput.Focus()
			m.message = ""
			return m, nil
		}
	case "a":
		// Change assignee
		if item, ok := m.columns[m.activeColumn].List.SelectedItem().(KanbanIssueItem); ok {
			m.selectedIssue = item.issue
			m.mode = ModeKanbanAssigneeChange
			m.message = ""
			return m, nil
		}
	case "d", "enter":
		// Show task details
		if item, ok := m.columns[m.activeColumn].List.SelectedItem().(KanbanIssueItem); ok {
			m.selectedIssue = item.issue
			m.loading = true
			m.message = ""
			return m, m.loadIssueDetails(item.issue.Key)
		}
	case "x":
		// Toggle current column visibility
		m.toggleColumnVisibility(m.activeColumn)
	case "+", "=":
		// Increase column width
		m.resizeColumn(m.activeColumn, 5)
	case "-":
		// Decrease column width
		m.resizeColumn(m.activeColumn, -5)
	}

	return m, nil
}

func (m KanbanBoardModel) handleStatusChangeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "backspace":
		m.mode = ModeKanbanNormal
		return m, nil
	case "up", "k":
		if m.transitionIndex > 0 {
			m.transitionIndex--
		}
	case "down", "j":
		if m.transitionIndex < len(m.transitions)-1 {
			m.transitionIndex++
		}
	case "enter":
		if len(m.transitions) > 0 && m.transitionIndex < len(m.transitions) {
			m.loading = true
			m.originalStatus = m.selectedIssue.Status // Store for potential rollback
			newStatus := m.transitions[m.transitionIndex].Name
			// Close popup immediately and show change on board
			m.mode = ModeKanbanNormal
			m.message = "Changing status..."
			// Optimistically update the board immediately
			m.updateIssueStatus(m.selectedIssue.Key, newStatus)
			return m, m.transitionIssue(m.selectedIssue.Key, m.transitions[m.transitionIndex].ID)
		}
	}
	return m, nil
}

func (m KanbanBoardModel) handleAddCommentKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeKanbanNormal
		return m, nil
	case "ctrl+s":
		body := m.commentInput.Value()
		if body != "" {
			m.loading = true
			return m, m.addComment(m.selectedIssue.Key, body)
		}
		m.mode = ModeKanbanNormal
		return m, nil
	default:
		var cmd tea.Cmd
		m.commentInput, cmd = m.commentInput.Update(msg)
		return m, cmd
	}
}

func (m KanbanBoardModel) handleAssigneeChangeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace":
		m.mode = ModeKanbanNormal
		return m, nil
	}
	return m, nil
}

func (m KanbanBoardModel) handleSprintAssignKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "backspace":
		m.mode = ModeKanbanNormal
		return m, nil
	case "up", "k":
		if m.sprintIndex > 0 {
			m.sprintIndex--
		}
	case "down", "j":
		if m.sprintIndex < len(m.sprints)-1 {
			m.sprintIndex++
		}
	case "enter":
		if len(m.sprints) > 0 && m.sprintIndex < len(m.sprints) {
			m.loading = true
			return m, m.assignToSprint(m.sprints[m.sprintIndex].ID, m.selectedIssue.Key)
		}
	}
	return m, nil
}

func (m KanbanBoardModel) handleTaskDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "enter", "d":
		m.mode = ModeKanbanNormal
		m.detailIssue = nil
		return m, nil
	case "up", "k":
		m.detailVP.LineUp(1)
		return m, nil
	case "down", "j":
		m.detailVP.LineDown(1)
		return m, nil
	case "pgup":
		m.detailVP.HalfViewUp()
		return m, nil
	case "pgdown", " ":
		m.detailVP.HalfViewDown()
		return m, nil
	case "home", "g":
		m.detailVP.GotoTop()
		return m, nil
	case "end", "G":
		m.detailVP.GotoBottom()
		return m, nil
	}
	return m, nil
}

// updateIssueInBoard updates an issue in the kanban board after a refresh
func (m *KanbanBoardModel) updateIssueInBoard(issue *models.Issue) {
	if issue == nil {
		return
	}

	// Find and remove the issue from its current column
	var oldColumnIdx, oldItemIdx = -1, -1
	for colIdx, col := range m.columns {
		for itemIdx, item := range col.Issues {
			if item.Key == issue.Key {
				oldColumnIdx = colIdx
				oldItemIdx = itemIdx
				break
			}
		}
		if oldColumnIdx != -1 {
			break
		}
	}

	// Create updated KanbanIssue
	updatedIssue := KanbanIssue{
		Key:     issue.Key,
		Summary: issue.Summary,
		Status:  issue.Status,
		Type:    issue.Type,
	}

	if issue.Assignee != nil {
		updatedIssue.Assignee = issue.Assignee.DisplayName
	}

	// Determine which column the issue should be in based on its new status
	targetColumnIdx := -1
	for idx, col := range m.columns {
		if col.Name == issue.Status {
			targetColumnIdx = idx
			break
		}
	}

	// Remove from old column if found
	if oldColumnIdx != -1 && oldItemIdx != -1 {
		col := &m.columns[oldColumnIdx]
		col.Issues = append(col.Issues[:oldItemIdx], col.Issues[oldItemIdx+1:]...)

		// Rebuild list items
		var items []list.Item
		for _, iss := range col.Issues {
			items = append(items, KanbanIssueItem{issue: iss})
		}
		col.List.SetItems(items)
		col.List.Title = fmt.Sprintf("%s (%d)", col.Name, len(col.Issues))
	}

	// Add to new column if a matching column exists
	if targetColumnIdx != -1 {
		col := &m.columns[targetColumnIdx]
		col.Issues = append(col.Issues, updatedIssue)

		// Rebuild list items
		var items []list.Item
		for _, iss := range col.Issues {
			items = append(items, KanbanIssueItem{issue: iss})
		}
		col.List.SetItems(items)
		col.List.Title = fmt.Sprintf("%s (%d)", col.Name, len(col.Issues))
	}
}

// updateIssueStatus updates an issue's status locally (optimistic update)
func (m *KanbanBoardModel) updateIssueStatus(issueKey, newStatus string) {
	// Find the issue and its current column
	var oldColumnIdx, oldItemIdx = -1, -1
	var issue KanbanIssue

	for colIdx, col := range m.columns {
		for itemIdx, item := range col.Issues {
			if item.Key == issueKey {
				oldColumnIdx = colIdx
				oldItemIdx = itemIdx
				issue = item
				issue.Status = newStatus
				break
			}
		}
		if oldColumnIdx != -1 {
			break
		}
	}

	if oldColumnIdx == -1 || oldItemIdx == -1 {
		return // Issue not found
	}

	// Find target column
	targetColumnIdx := -1
	for idx, col := range m.columns {
		if col.Name == newStatus {
			targetColumnIdx = idx
			break
		}
	}

	// Remove from old column
	oldCol := &m.columns[oldColumnIdx]
	oldCol.Issues = append(oldCol.Issues[:oldItemIdx], oldCol.Issues[oldItemIdx+1:]...)

	// Rebuild old column list
	var oldItems []list.Item
	for _, iss := range oldCol.Issues {
		oldItems = append(oldItems, KanbanIssueItem{issue: iss})
	}
	oldCol.List.SetItems(oldItems)
	oldCol.List.Title = fmt.Sprintf("%s (%d)", oldCol.Name, len(oldCol.Issues))

	// Add to new column if found
	if targetColumnIdx != -1 {
		newCol := &m.columns[targetColumnIdx]
		newCol.Issues = append(newCol.Issues, issue)

		// Rebuild new column list
		var newItems []list.Item
		for _, iss := range newCol.Issues {
			newItems = append(newItems, KanbanIssueItem{issue: iss})
		}
		newCol.List.SetItems(newItems)
		newCol.List.Title = fmt.Sprintf("%s (%d)", newCol.Name, len(newCol.Issues))
	}
}

// calculateColumnWidth calculates the width for each column based on terminal width
func (m KanbanBoardModel) calculateColumnWidth() int {
	if len(m.columns) == 0 {
		return 30
	}

	// Minimum column width
	minWidth := 20

	// Calculate available width (accounting for borders and padding)
	availableWidth := m.width - 10
	if availableWidth < 0 {
		availableWidth = 0
	}

	// Calculate width per column
	columnWidth := availableWidth / len(m.columns)

	// Ensure minimum width
	if columnWidth < minWidth {
		columnWidth = minWidth
	}

	// Cap maximum width
	maxWidth := 35
	if columnWidth > maxWidth {
		columnWidth = maxWidth
	}

	return columnWidth
}

// getColumnStyle returns the appropriate style for a column based on its status and whether it's active
func (m KanbanBoardModel) getColumnStyle(status string, isActive bool) lipgloss.Style {
	switch status {
	case "To Do":
		if isActive {
			return todoActiveColumnStyle
		}
		return todoColumnStyle
	case "In Progress":
		if isActive {
			return inProgressActiveColumnStyle
		}
		return inProgressColumnStyle
	case "Done":
		if isActive {
			return doneActiveColumnStyle
		}
		return doneColumnStyle
	case "Blocked":
		if isActive {
			return blockedActiveColumnStyle
		}
		return blockedColumnStyle
	case "Review":
		if isActive {
			return reviewActiveColumnStyle
		}
		return reviewColumnStyle
	default:
		if isActive {
			return activeColumnStyle
		}
		return columnStyle
	}
}

func (m KanbanBoardModel) View() string {
	// Always render the kanban board as background
	background := m.kanbanView()

	// If in normal mode, just return the background
	if m.mode == ModeKanbanNormal {
		return background
	}

	// For popup modes, render popup centered over the board
	var popupContent string
	switch m.mode {
	case ModeKanbanStatusChange:
		popupContent = m.statusChangePopup()
	case ModeKanbanAddComment:
		popupContent = m.addCommentPopup()
	case ModeKanbanAssigneeChange:
		popupContent = m.assigneeChangePopup()
	case ModeKanbanSprintAssign:
		popupContent = m.sprintAssignPopup()
	case ModeKanbanTaskDetail:
		popupContent = m.taskDetailPopup()
	}

	return m.overlayPopup(background, popupContent)
}

func (m KanbanBoardModel) overlayPopup(background, popup string) string {
	bgLines := strings.Split(background, "\n")
	popupLines := strings.Split(popup, "\n")

	// Calculate dimensions
	bgHeight := len(bgLines)
	popupHeight := len(popupLines)
	popupWidth := lipgloss.Width(popup)

	// Center vertically
	startRow := (bgHeight - popupHeight) / 2
	if startRow < 0 {
		startRow = 0
	}

	// Build result with popup overlay
	var result []string
	for i, line := range bgLines {
		if i >= startRow && i < startRow+popupHeight {
			popupLineIdx := i - startRow
			popupLine := popupLines[popupLineIdx]

			// Center horizontally
			lineWidth := lipgloss.Width(line)
			startCol := (lineWidth - popupWidth) / 2
			if startCol < 0 {
				startCol = 0
			}

			// Build the line: [before][popup][after]
			// We need to handle ANSI codes properly
			before := m.substringByWidth(line, 0, startCol)
			after := m.substringByWidth(line, startCol+popupWidth, lineWidth)

			result = append(result, before+popupLine+after)
		} else {
			// Outside popup area, keep line as-is (dimmed)
			result = append(result, dimmedStyle.Render(line))
		}
	}

	return strings.Join(result, "\n")
}

// substringByWidth extracts a substring by visual width, handling ANSI codes
func (m KanbanBoardModel) substringByWidth(s string, start, end int) string {
	if start >= end {
		return ""
	}

	var result strings.Builder
	currentWidth := 0
	inEscape := false
	escapeSeq := strings.Builder{}

	for _, r := range s {
		if inEscape {
			escapeSeq.WriteRune(r)
			if r == 'm' {
				// End of escape sequence
				if currentWidth >= start && currentWidth < end {
					result.WriteString(escapeSeq.String())
				}
				inEscape = false
				escapeSeq.Reset()
			}
			continue
		}

		if r == '\x1b' {
			inEscape = true
			escapeSeq.WriteRune(r)
			continue
		}

		if currentWidth >= start && currentWidth < end {
			result.WriteRune(r)
		}
		currentWidth++

		if currentWidth >= end {
			break
		}
	}

	return result.String()
}

func (m KanbanBoardModel) kanbanView() string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Sprint Board - Sprint %d", m.sprintID)))
	b.WriteString("\n\n")

	// Show message if any
	if m.message != "" {
		if strings.HasPrefix(m.message, "✓") {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00C851")).Render(m.message))
		} else if strings.HasPrefix(m.message, "✗") {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Render(m.message))
		} else {
			b.WriteString(m.message)
		}
		b.WriteString("\n\n")
	}

	// Calculate dynamic column width
	columnWidth := m.calculateColumnWidth()

	// Render columns side by side
	m.hiddenCount = 0
	var visibleColumns []string
	var hiddenColumns []string
	for i, col := range m.columns {
		isActive := i == m.activeColumn
		style := m.getColumnStyle(col.Name, isActive)

		// Hidden columns render separately below
		if col.Hidden {
			m.hiddenCount++
			// Add indicator only when in focusHiddenColumns mode and this column is active
			dimStyle := lipgloss.NewStyle().
				Padding(0, 2).
				Foreground(lipgloss.Color("#666666"))
			activeStyle := lipgloss.NewStyle().
				Padding(0, 2).
				Foreground(lipgloss.Color("#7D56F4"))
			if m.focusHiddenColumns && isActive {
				hiddenColumns = append(hiddenColumns, activeStyle.Render("▸ "+col.Name))
			} else {
				hiddenColumns = append(hiddenColumns, dimStyle.Render(col.Name))
			}
			continue
		}

		// Use custom width if set, otherwise use calculated
		if col.Width > 0 {
			style = style.Width(col.Width)
		} else {
			style = style.Width(columnWidth)
		}

		// Update column title with indicator if active (only when not focusing hidden columns)
		title := col.List.Title
		if isActive && !m.focusHiddenColumns {
			title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render("▸ " + title)
		}
		col.List.Title = title

		// Configure list delegate based on active state
		delegate := list.NewDefaultDelegate()
		// Configure to show single line items that expand/contract with column width
		delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Copy().MaxWidth(0).Height(1)
		delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Copy().MaxWidth(0).Height(1)
		delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Copy().MaxWidth(0).Height(1)
		delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Copy().MaxWidth(0).Height(1)
		if isActive && !m.focusHiddenColumns {
			// Use purple for selected items in active column (only when not focusing hidden columns)
			delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("#7D56F4"))
			delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("#a277ff"))
		} else {
			// Hide selection in inactive columns by making selected style same as normal
			delegate.Styles.SelectedTitle = delegate.Styles.NormalTitle
			delegate.Styles.SelectedDesc = delegate.Styles.NormalDesc
		}
		col.List.SetDelegate(delegate)

		colView := style.Render(col.List.View())
		visibleColumns = append(visibleColumns, colView)
	}

	// Join visible columns horizontally
	if len(visibleColumns) > 0 {
		row := lipgloss.JoinHorizontal(lipgloss.Top, visibleColumns...)
		b.WriteString(row)
	}

	// Render hidden columns in a separate row below
	if len(hiddenColumns) > 0 {
		b.WriteString("\n")
		hiddenRow := lipgloss.JoinHorizontal(lipgloss.Top, hiddenColumns...)
		b.WriteString(hiddenRow)
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("←/→: switch cols • ↑/↓: navigate • f: focus hidden cols • d: details • s: status • c: comment • x: toggle col • +/-: resize • q: quit"))

	return b.String()
}

func (m KanbanBoardModel) statusChangePopup() string {
	var b strings.Builder

	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render(fmt.Sprintf("▸ %s", m.selectedIssue.Key)))
	b.WriteString("\n\n")
	b.WriteString(whiteStyle.Bold(true).Render("Change Status"))
	b.WriteString(whiteStyle.Render(fmt.Sprintf(" (current: %s)\n\n", m.selectedIssue.Status)))

	if m.loading {
		b.WriteString(loadingStyle.Render("Loading available transitions..."))
	} else if len(m.transitions) == 0 {
		b.WriteString(whiteStyle.Render("No available transitions\n"))
	} else {
		for i, t := range m.transitions {
			if i == m.transitionIndex {
				b.WriteString(selectedActionStyle.Render(fmt.Sprintf("▸ %s", t.Name)))
			} else {
				b.WriteString(whiteStyle.Render(fmt.Sprintf("  %s", t.Name)))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: navigate • enter: select • esc: back"))

	// Wrap in popup style
	return popupStyle.Render(b.String())
}

func (m KanbanBoardModel) addCommentPopup() string {
	var b strings.Builder

	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render(fmt.Sprintf("▸ %s", m.selectedIssue.Key)))
	b.WriteString("\n\n")
	b.WriteString(whiteStyle.Bold(true).Render("Add Comment"))
	b.WriteString("\n\n")

	b.WriteString(m.commentInput.View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("ctrl+s: save • esc: cancel"))

	// Wrap in popup style
	return popupStyle.Render(b.String())
}

func (m KanbanBoardModel) assigneeChangePopup() string {
	var b strings.Builder

	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render(fmt.Sprintf("▸ %s", m.selectedIssue.Key)))
	b.WriteString("\n\n")
	b.WriteString(whiteStyle.Bold(true).Render("Change Assignee"))
	b.WriteString("\n\n")

	b.WriteString(whiteStyle.Render("Current assignee functionality requires email lookup.\n"))
	b.WriteString(whiteStyle.Render("This feature will be available in a future update.\n\n"))

	b.WriteString(helpStyle.Render("esc: back"))

	// Wrap in popup style
	return popupStyle.Render(b.String())
}

func (m KanbanBoardModel) sprintAssignPopup() string {
	var b strings.Builder

	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render(fmt.Sprintf("▸ %s", m.selectedIssue.Key)))
	b.WriteString("\n\n")
	b.WriteString(whiteStyle.Bold(true).Render("Put in Sprint"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(loadingStyle.Render("Loading available sprints..."))
	} else if len(m.sprints) == 0 {
		b.WriteString(whiteStyle.Render("No open sprints available\n"))
	} else {
		for i, sprint := range m.sprints {
			if i == m.sprintIndex {
				b.WriteString(selectedActionStyle.Render(fmt.Sprintf("▸ %s", sprint.Name)))
			} else {
				b.WriteString(whiteStyle.Render(fmt.Sprintf("  %s", sprint.Name)))
			}
			// Add dates if available
			if !sprint.StartDate.IsZero() && !sprint.EndDate.IsZero() {
				b.WriteString(whiteStyle.Render(fmt.Sprintf(" (%s - %s)", sprint.StartDate.Time().Format("Jan 2"), sprint.EndDate.Time().Format("Jan 2"))))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: navigate • enter: select • esc: back"))

	// Wrap in popup style
	return popupStyle.Render(b.String())
}

func (m KanbanBoardModel) taskDetailPopup() string {
	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	if m.loading {
		return popupStyle.Render(loadingStyle.Render("Loading issue details..."))
	}

	if m.detailIssue == nil {
		return popupStyle.Render(whiteStyle.Render("No issue details available"))
	}

	// Return the viewport content
	header := m.taskDetailHeader()
	footer := helpStyle.Render("esc/enter/d: close • j/k: scroll down/up • space: page down")

	// Use a wider popup for details with viewport
	// Note: No custom background to avoid color inconsistencies with markdown
	detailPopupStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 2).
		Width(m.width - 20)

	content := header + "\n" + m.detailVP.View() + "\n" + footer
	return detailPopupStyle.Render(content)
}

func (m KanbanBoardModel) taskDetailHeader() string {
	var b strings.Builder

	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))

	issue := m.detailIssue

	// Header with key and type
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render(fmt.Sprintf("▸ %s", issue.Key)))
	b.WriteString(grayStyle.Render(fmt.Sprintf("  [%s]", issue.Type)))
	b.WriteString("\n\n")

	// Summary
	b.WriteString(whiteStyle.Bold(true).Render(issue.Summary))
	b.WriteString("\n\n")

	// Status and Assignee
	b.WriteString(grayStyle.Render("Status: "))
	b.WriteString(whiteStyle.Render(issue.Status))
	b.WriteString("\n")

	b.WriteString(grayStyle.Render("Assignee: "))
	if issue.Assignee != nil && issue.Assignee.DisplayName != "" {
		b.WriteString(whiteStyle.Render(issue.Assignee.DisplayName))
	} else {
		b.WriteString(whiteStyle.Render("Unassigned"))
	}
	b.WriteString("\n")

	// Multi-owners if available
	if len(issue.Owners) > 0 {
		b.WriteString(grayStyle.Render("Owners: "))
		var ownerNames []string
		for _, owner := range issue.Owners {
			if owner.DisplayName != "" {
				ownerNames = append(ownerNames, owner.DisplayName)
			}
		}
		if len(ownerNames) > 0 {
			b.WriteString(whiteStyle.Render(strings.Join(ownerNames, ", ")))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Description label
	b.WriteString(grayStyle.Bold(true).Render("Description"))
	b.WriteString("\n")

	return b.String()
}

func (m KanbanBoardModel) taskDetailContent() string {
	if m.detailIssue == nil || m.detailIssue.Description == "" {
		return ""
	}

	// Use the rich markdown renderer - no custom background to avoid color inconsistencies
	md := NewRichMarkdownRenderer(uint(m.width - 28)) // account for padding and borders
	rendered, err := md.Render(m.detailIssue.Description)
	if err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render("Error rendering description")
	}

	return rendered
}

func (m KanbanBoardModel) GetTitle() string {
	return "Kanban Board"
}
