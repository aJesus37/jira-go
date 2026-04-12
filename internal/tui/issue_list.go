// internal/tui/issue_list.go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/aJesus37/jira-go/internal/api"
	"github.com/aJesus37/jira-go/internal/models"
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

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Bold(true)

	actionMenuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)

	selectedActionStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 1)
)

// Issue represents a simplified issue for TUI display
type Issue struct {
	Key             string
	Summary         string
	Type            string
	Status          string
	Assignee        string
	AssigneeEmail   string
	Owners          []string
	OwnerEmails     []string
	OwnerAccountIDs []string
	Description     string
	Created         string
	SprintName      string
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

	if i.issue.SprintName != "" {
		parts = append(parts, "🏃 "+i.issue.SprintName)
	}

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

// Messages for async operations
type issueLoadedMsg struct {
	issue *models.Issue
	err   error
}

type markdownRenderedMsg struct {
	content string
}

type transitionsLoadedMsg struct {
	transitions []api.Transition
	err         error
}

type actionCompletedMsg struct {
	action string
	err    error
}

type sprintsLoadedMsg struct {
	sprints []models.Sprint
	err     error
}

type sprintAssignedMsg struct {
	sprintName string
	err        error
}

type usersSearchedMsg struct {
	users []models.User
	err   error
}

// ViewMode represents the current view mode
type ViewMode int

const (
	ModeList ViewMode = iota
	ModeDetail
	ModeStatusChange
	ModeAssigneeChange
	ModeManageOwners
	ModeAddComment
	ModeSprintAssign
)

// ActionMenuItem represents an action menu option
type ActionMenuItem struct {
	Key         string
	Label       string
	Description string
}

// IssueListModel is the TUI model for listing issues
type IssueListModel struct {
	list              list.Model
	issues            []Issue
	client            *api.Client
	projectKey        string
	ownerFieldID      string
	sprintFieldID     string
	boardID           int
	mode              ViewMode
	selected          Issue
	width             int
	height            int
	viewport          viewport.Model
	renderer          *RichMarkdownRenderer
	renderedDesc      string
	renderedDescWidth int
	loading           bool
	message           string // For showing success/error messages

	// Status change
	transitions     []api.Transition
	transitionIndex int

	// Assignee change
	assigneeInput string
	ownersInput   string
	searchResults []models.User
	searchIndex   int

	// Owner management
	ownerSearchInput   string
	ownerSearchResults []models.User
	ownerSearchIndex   int
	selectedOwnerIndex int

	// Comment
	commentInput textarea.Model

	// Sprint assignment
	sprints     []models.Sprint
	sprintIndex int
}

// NewIssueList creates a new issue list TUI
func NewIssueList(issues []models.Issue, client *api.Client, projectKey, ownerFieldID string, sprintFieldID string, boardID int) IssueListModel {
	var items []list.Item
	var tuiIssues []Issue

	for _, issue := range issues {
		tuiIssue := Issue{
			Key:        issue.Key,
			Summary:    issue.Summary,
			Type:       issue.Type,
			Status:     issue.Status,
			SprintName: issue.SprintName,
			Created:    issue.Created.Format("2006-01-02 15:04"),
		}

		if issue.Assignee != nil {
			tuiIssue.Assignee = issue.Assignee.DisplayName
			tuiIssue.AssigneeEmail = issue.Assignee.Email
		}

		for _, owner := range issue.Owners {
			tuiIssue.Owners = append(tuiIssue.Owners, owner.DisplayName)
			tuiIssue.OwnerEmails = append(tuiIssue.OwnerEmails, owner.Email)
			tuiIssue.OwnerAccountIDs = append(tuiIssue.OwnerAccountIDs, owner.AccountID)
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

	// Initialize viewport for scrolling
	vp := viewport.New(100, 20)

	// Initialize rich markdown renderer with syntax highlighting
	renderer := NewRichMarkdownRenderer(80)

	// Initialize comment input
	commentInput := textarea.New()
	commentInput.Placeholder = "Enter your comment..."
	commentInput.SetWidth(60)
	commentInput.SetHeight(4)

	return IssueListModel{
		list:          l,
		issues:        tuiIssues,
		client:        client,
		projectKey:    projectKey,
		ownerFieldID:  ownerFieldID,
		sprintFieldID: sprintFieldID,
		boardID:       boardID,
		mode:          ModeList,
		viewport:      vp,
		renderer:      renderer,
		loading:       false,
		commentInput:  commentInput,
		sprintIndex:   0,
	}
}

func (m IssueListModel) Init() tea.Cmd {
	return nil
}

// Async function to load issue details
func (m IssueListModel) loadIssue(key string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return issueLoadedMsg{issue: nil, err: fmt.Errorf("no client")}
		}
		issue, err := m.client.GetIssue(key, m.ownerFieldID, m.sprintFieldID)
		return issueLoadedMsg{issue: issue, err: err}
	}
}

// Async function to render markdown
func (m IssueListModel) renderMarkdown(content string, width int) tea.Cmd {
	return func() tea.Msg {
		if content == "" {
			return markdownRenderedMsg{content: helpStyle.Render("(no description)")}
		}
		if m.renderer == nil {
			return markdownRenderedMsg{content: content}
		}

		// Update renderer width
		m.renderer.SetWidth(uint(width))

		// Render markdown - FastMarkdownRenderer.Render doesn't return error
		rendered, _ := m.renderer.Render(content)
		return markdownRenderedMsg{content: rendered}
	}
}

// Async function to load transitions
func (m IssueListModel) loadTransitions(key string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return transitionsLoadedMsg{transitions: nil, err: fmt.Errorf("no client")}
		}
		transitions, err := m.client.GetTransitions(key)
		return transitionsLoadedMsg{transitions: transitions, err: err}
	}
}

// Async function to transition issue
func (m IssueListModel) transitionIssue(key, transitionID string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return actionCompletedMsg{action: "transition", err: fmt.Errorf("no client")}
		}
		err := m.client.TransitionIssue(key, transitionID)
		return actionCompletedMsg{action: "transition", err: err}
	}
}

// Async function to add comment
func (m IssueListModel) addComment(key, body string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return actionCompletedMsg{action: "comment", err: fmt.Errorf("no client")}
		}
		err := m.client.AddComment(key, body)
		return actionCompletedMsg{action: "comment", err: err}
	}
}

// Async function to refresh a single issue and update the list
func (m IssueListModel) refreshIssue(key string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return issueLoadedMsg{issue: nil, err: fmt.Errorf("no client")}
		}
		issue, err := m.client.GetIssue(key, m.ownerFieldID, m.sprintFieldID)
		return issueLoadedMsg{issue: issue, err: err}
	}
}

// Async function to load open sprints
func (m IssueListModel) loadSprints() tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return sprintsLoadedMsg{err: fmt.Errorf("no client")}
		}
		if m.boardID == 0 {
			return sprintsLoadedMsg{err: fmt.Errorf("board not configured")}
		}
		sprints, err := m.client.GetOpenSprints(m.boardID)
		return sprintsLoadedMsg{sprints: sprints, err: err}
	}
}

// Async function to assign issue to sprint
func (m IssueListModel) assignToSprint(sprintID int, issueKey string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return sprintAssignedMsg{err: fmt.Errorf("no client")}
		}
		err := m.client.MoveIssuesToSprint(sprintID, []string{issueKey})
		return sprintAssignedMsg{err: err}
	}
}

// Async function to search users for autocomplete
func (m IssueListModel) searchUsers(query string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil || query == "" {
			return usersSearchedMsg{users: nil, err: nil}
		}
		users, err := m.client.SearchUsers(query)
		return usersSearchedMsg{users: users, err: err}
	}
}

// Async function to change assignee
func (m IssueListModel) changeAssignee(issueKey, email string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return actionCompletedMsg{action: "assignee", err: fmt.Errorf("no client")}
		}
		user, err := m.client.ResolveEmail(email)
		if err != nil {
			return actionCompletedMsg{action: "assignee", err: fmt.Errorf("resolving user: %w", err)}
		}
		err = m.client.AssignIssue(issueKey, user.AccountID)
		return actionCompletedMsg{action: "assignee", err: err}
	}
}

// Async function to add owner by email
func (m IssueListModel) addOwner(issueKey, ownerFieldID, email string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return actionCompletedMsg{action: "owner", err: fmt.Errorf("no client")}
		}
		err := m.client.AddOwnerToIssue(issueKey, ownerFieldID, email)
		return actionCompletedMsg{action: "owner", err: err}
	}
}

// Async function to add owner by account ID
func (m IssueListModel) addOwnerByAccountID(issueKey, ownerFieldID, accountID string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return actionCompletedMsg{action: "owner", err: fmt.Errorf("no client")}
		}
		err := m.client.AddOwnerByAccountID(issueKey, ownerFieldID, accountID)
		return actionCompletedMsg{action: "owner", err: err}
	}
}

// Async function to remove owner
func (m IssueListModel) removeOwner(issueKey, ownerFieldID, accountID string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return actionCompletedMsg{action: "owner", err: fmt.Errorf("no client")}
		}
		err := m.client.RemoveOwnerByAccountID(issueKey, ownerFieldID, accountID)
		return actionCompletedMsg{action: "owner", err: err}
	}
}

func (m IssueListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case issueLoadedMsg:
		m.loading = false
		if msg.err == nil && msg.issue != nil {
			m.selected.Description = msg.issue.Description
			if msg.issue.Assignee != nil {
				m.selected.Assignee = msg.issue.Assignee.DisplayName
				m.selected.AssigneeEmail = msg.issue.Assignee.Email
			}
			m.selected.Owners = []string{}
			m.selected.OwnerEmails = []string{}
			m.selected.OwnerAccountIDs = []string{}
			for _, o := range msg.issue.Owners {
				m.selected.Owners = append(m.selected.Owners, fmt.Sprintf("%s (%s)", o.DisplayName, o.Email))
				m.selected.OwnerEmails = append(m.selected.OwnerEmails, o.Email)
				m.selected.OwnerAccountIDs = append(m.selected.OwnerAccountIDs, o.AccountID)
			}
			// Trigger markdown rendering
			contentWidth := m.width - 4
			if contentWidth < 40 {
				contentWidth = 40
			}
			return m, m.renderMarkdown(m.selected.Description, contentWidth)
		}
		return m, nil

	case markdownRenderedMsg:
		m.renderedDesc = msg.content
		m.renderedDescWidth = m.width - 4
		m.loading = false
		// Update viewport with rendered content
		m.viewport.SetContent(m.renderedDesc)
		return m, nil

	case transitionsLoadedMsg:
		m.loading = false
		if msg.err == nil {
			m.transitions = msg.transitions
			m.transitionIndex = 0
			m.mode = ModeStatusChange
		}
		return m, nil

	case actionCompletedMsg:
		m.loading = false
		if msg.err == nil {
			m.message = fmt.Sprintf("✓ %s completed successfully", msg.action)
			if (msg.action == "transition" || msg.action == "assignee" || msg.action == "owner") && m.selected.Key != "" {
				return m, m.refreshIssue(m.selected.Key)
			}
		} else {
			m.message = fmt.Sprintf("✗ %s failed: %v", msg.action, msg.err)
		}
		m.mode = ModeList
		return m, nil

	case usersSearchedMsg:
		if msg.err != nil {
			m.searchResults = nil
			m.ownerSearchResults = nil
		} else {
			m.searchResults = msg.users
			m.ownerSearchResults = msg.users
			if m.searchIndex >= len(m.searchResults) {
				m.searchIndex = 0
			}
			if m.ownerSearchIndex >= len(m.ownerSearchResults) {
				m.ownerSearchIndex = 0
			}
		}
		return m, nil

	case tea.KeyMsg:
		// Handle quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle different modes
		switch m.mode {
		case ModeDetail:
			return m.handleDetailViewKeys(msg)
		case ModeStatusChange:
			return m.handleStatusChangeKeys(msg)
		case ModeAssigneeChange:
			return m.handleAssigneeChangeKeys(msg)
		case ModeManageOwners:
			return m.handleManageOwnersKeys(msg)
		case ModeAddComment:
			return m.handleAddCommentKeys(msg)
		case ModeSprintAssign:
			return m.handleSprintAssignKeys(msg)
		case ModeList:
			return m.handleListViewKeys(msg)
		}

	case tea.WindowSizeMsg:
		oldWidth := m.width
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, m.height-4)
		// Update viewport dimensions
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 12
		if m.viewport.Height < 10 {
			m.viewport.Height = 10
		}
		// Only re-render markdown if width changed significantly (>10 chars) and we're viewing
		widthDiff := oldWidth - msg.Width
		if widthDiff < 0 {
			widthDiff = -widthDiff
		}
		if m.mode == ModeDetail && m.selected.Description != "" && widthDiff > 10 {
			// Check if we already have rendered content at a different width
			if m.renderedDesc != "" && m.renderedDescWidth != m.width-4 {
				// Re-render with new width
				m.renderedDesc = "" // Clear to trigger re-render
				contentWidth := m.width - 4
				if contentWidth < 40 {
					contentWidth = 40
				}
				return m, m.renderMarkdown(m.selected.Description, contentWidth)
			}
		}
	}

	var cmd tea.Cmd
	if m.mode == ModeList {
		m.list, cmd = m.list.Update(msg)
	} else if m.mode == ModeDetail {
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

func (m IssueListModel) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "enter", "o":
		if item, ok := m.list.SelectedItem().(IssueItem); ok {
			m.selected = item.issue
			m.mode = ModeDetail
			m.loading = true
			m.renderedDesc = ""
			m.viewport.GotoTop()
			// Load issue async
			return m, m.loadIssue(item.issue.Key)
		}
	case "s":
		// Change status from list view
		if item, ok := m.list.SelectedItem().(IssueItem); ok {
			m.selected = item.issue
			m.mode = ModeStatusChange
			m.loading = true
			m.message = ""
			return m, m.loadTransitions(item.issue.Key)
		}
	case "c":
		// Add comment from list view
		if item, ok := m.list.SelectedItem().(IssueItem); ok {
			m.selected = item.issue
			m.mode = ModeAddComment
			m.commentInput.SetValue("")
			m.commentInput.Focus()
			m.message = ""
			return m, nil
		}
	case "a":
		// Change assignee from list view
		if item, ok := m.list.SelectedItem().(IssueItem); ok {
			m.selected = item.issue
			m.mode = ModeAssigneeChange
			m.assigneeInput = ""
			m.searchResults = nil
			m.searchIndex = 0
			m.message = ""
			return m, nil
		}
	case "u":
		// Manage owners from list view
		if item, ok := m.list.SelectedItem().(IssueItem); ok {
			m.selected = item.issue
			m.mode = ModeManageOwners
			m.ownerSearchInput = ""
			m.ownerSearchResults = nil
			m.ownerSearchIndex = 0
			m.selectedOwnerIndex = -1
			m.message = ""
			return m, nil
		}
	case "p":
		// Put in sprint from list view
		if item, ok := m.list.SelectedItem().(IssueItem); ok {
			m.selected = item.issue
			m.mode = ModeSprintAssign
			m.loading = true
			m.message = ""
			return m, m.loadSprints()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m IssueListModel) handleDetailViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.mode = ModeList
		m.loading = false
		m.renderedDesc = ""
		m.message = ""
		return m, nil
	case "s":
		m.loading = true
		return m, m.loadTransitions(m.selected.Key)
	case "u":
		m.mode = ModeManageOwners
		m.ownerSearchInput = ""
		m.ownerSearchResults = nil
		m.ownerSearchIndex = 0
		m.selectedOwnerIndex = -1
		m.message = ""
		return m, nil
	case "up", "k":
		m.viewport, _ = m.viewport.Update(msg)
		return m, nil
	case "down", "j":
		m.viewport, _ = m.viewport.Update(msg)
		return m, nil
	case "pgup":
		m.viewport.LineUp(10)
		return m, nil
	case "pgdown", " ":
		m.viewport.LineDown(10)
		return m, nil
	case "home":
		m.viewport.GotoTop()
		return m, nil
	case "end":
		m.viewport.GotoBottom()
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m IssueListModel) handleStatusChangeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.mode = ModeList
		return m, nil
	case "up", "k":
		if m.transitionIndex > 0 {
			m.transitionIndex--
		}
		return m, nil
	case "down", "j":
		if m.transitionIndex < len(m.transitions)-1 {
			m.transitionIndex++
		}
		return m, nil
	case "enter":
		if len(m.transitions) > 0 && m.transitionIndex < len(m.transitions) {
			m.loading = true
			return m, m.transitionIssue(m.selected.Key, m.transitions[m.transitionIndex].ID)
		}
	}

	return m, nil
}

func (m IssueListModel) handleAssigneeChangeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeList
		m.assigneeInput = ""
		m.searchResults = nil
		return m, nil
	case "backspace":
		if len(m.assigneeInput) > 0 {
			m.assigneeInput = m.assigneeInput[:len(m.assigneeInput)-1]
			m.searchIndex = 0
			if m.assigneeInput != "" {
				return m, m.searchUsers(m.assigneeInput)
			}
			m.searchResults = nil
		}
		return m, nil
	case "up", "k":
		if len(m.searchResults) > 0 && m.searchIndex > 0 {
			m.searchIndex--
		}
		return m, nil
	case "down", "j":
		if len(m.searchResults) > 0 && m.searchIndex < len(m.searchResults)-1 {
			m.searchIndex++
		}
		return m, nil
	case "enter":
		if len(m.searchResults) > 0 && m.searchIndex < len(m.searchResults) {
			selected := m.searchResults[m.searchIndex]
			m.loading = true
			m.assigneeInput = ""
			m.searchResults = nil
			return m, m.changeAssignee(m.selected.Key, selected.Email)
		}
		if m.assigneeInput != "" {
			m.loading = true
			m.searchResults = nil
			return m, m.changeAssignee(m.selected.Key, m.assigneeInput)
		}
		return m, nil
	case "ctrl+s":
		if m.assigneeInput != "" {
			m.loading = true
			m.searchResults = nil
			return m, m.changeAssignee(m.selected.Key, m.assigneeInput)
		}
		return m, nil
	default:
		char := ""
		if len(msg.Runes) > 0 {
			char = string(msg.Runes[0])
		} else if len(msg.String()) == 1 {
			char = msg.String()
		}
		if char != "" {
			m.assigneeInput += char
			m.searchIndex = 0
			return m, m.searchUsers(m.assigneeInput)
		}
	}
	return m, nil
}

func (m IssueListModel) handleManageOwnersKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeList
		m.ownerSearchInput = ""
		m.ownerSearchResults = nil
		m.selectedOwnerIndex = -1
		return m, nil
	case "up", "k":
		if m.selectedOwnerIndex > 0 {
			m.selectedOwnerIndex--
		}
		return m, nil
	case "down", "j":
		maxOwners := len(m.selected.Owners)
		if m.ownerSearchResults != nil && len(m.ownerSearchResults) > 0 {
			maxOwners = len(m.selected.Owners) + len(m.ownerSearchResults)
		}
		if m.selectedOwnerIndex < maxOwners-1 {
			m.selectedOwnerIndex++
		}
		return m, nil
	case "enter":
		// First check if selecting an owner to remove
		if m.selectedOwnerIndex >= 0 && m.selectedOwnerIndex < len(m.selected.Owners) {
			accountID := m.selected.OwnerAccountIDs[m.selectedOwnerIndex]
			m.loading = true
			return m, m.removeOwner(m.selected.Key, m.ownerFieldID, accountID)
		}

		// Then check if selecting from search results to add
		if m.ownerSearchResults != nil && len(m.ownerSearchResults) > 0 {
			searchIdx := m.selectedOwnerIndex - len(m.selected.Owners)
			if searchIdx >= 0 && searchIdx < len(m.ownerSearchResults) {
				selected := m.ownerSearchResults[searchIdx]
				m.loading = true
				m.ownerSearchInput = ""
				m.ownerSearchResults = nil
				return m, m.addOwnerByAccountID(m.selected.Key, m.ownerFieldID, selected.AccountID)
			}
		}

		return m, nil
	case "backspace":
		if len(m.ownerSearchInput) > 0 {
			m.ownerSearchInput = m.ownerSearchInput[:len(m.ownerSearchInput)-1]
			m.ownerSearchIndex = 0
			if m.ownerSearchInput != "" {
				return m, m.searchUsers(m.ownerSearchInput)
			}
			m.ownerSearchResults = nil
		}
		return m, nil
	default:
		char := ""
		if len(msg.Runes) > 0 {
			char = string(msg.Runes[0])
		} else if len(msg.String()) == 1 {
			char = msg.String()
		}
		if char != "" {
			m.ownerSearchInput += char
			m.ownerSearchIndex = 0
			m.selectedOwnerIndex = len(m.selected.Owners)
			return m, m.searchUsers(m.ownerSearchInput)
		}
	}
	return m, nil
}

func (m IssueListModel) handleAddCommentKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeList
		return m, nil
	case "ctrl+s":
		body := m.commentInput.Value()
		if body != "" {
			m.loading = true
			return m, m.addComment(m.selected.Key, body)
		}
		m.mode = ModeList
		return m, nil
	default:
		var cmd tea.Cmd
		m.commentInput, cmd = m.commentInput.Update(msg)
		return m, cmd
	}
}

func (m IssueListModel) handleSprintAssignKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "backspace":
		m.mode = ModeList
		return m, nil
	case "up", "k":
		if m.sprintIndex > 0 {
			m.sprintIndex--
		}
		return m, nil
	case "down", "j":
		if m.sprintIndex < len(m.sprints)-1 {
			m.sprintIndex++
		}
		return m, nil
	case "enter":
		if len(m.sprints) > 0 && m.sprintIndex < len(m.sprints) {
			m.loading = true
			return m, m.assignToSprint(m.sprints[m.sprintIndex].ID, m.selected.Key)
		}
	}

	return m, nil
}

func (m IssueListModel) View() string {
	switch m.mode {
	case ModeDetail:
		return m.detailView()
	case ModeStatusChange:
		return m.statusChangeView()
	case ModeAssigneeChange:
		return m.assigneeChangeView()
	case ModeManageOwners:
		return m.manageOwnersView()
	case ModeAddComment:
		return m.addCommentView()
	case ModeSprintAssign:
		return m.sprintAssignView()
	default:
		return m.list.View() + "\n" + helpStyle.Render("↑/↓: navigate • enter/o: open • s: status • c: comment • a: assignee • u: owners • p: put in sprint • q: quit")
	}
}

func (m IssueListModel) detailView() string {
	var b strings.Builder

	// Header info
	b.WriteString(titleStyle.Render(fmt.Sprintf(" %s ", m.selected.Key)))
	b.WriteString("\n\n")
	b.WriteString(selectedStyle.Render(m.selected.Summary))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Type: %s\n", m.selected.Type))
	b.WriteString(fmt.Sprintf("Status: %s\n", m.selected.Status))

	// Show multi-owners if available, otherwise show single assignee
	if len(m.selected.Owners) > 0 {
		b.WriteString(assigneeStyle.Render(fmt.Sprintf("👥 Owners: %s\n", strings.Join(m.selected.Owners, ", "))))
	} else if m.selected.Assignee != "" {
		b.WriteString(assigneeStyle.Render(fmt.Sprintf("👤 Assignee: %s", m.selected.Assignee)))
		if m.selected.AssigneeEmail != "" {
			b.WriteString(fmt.Sprintf(" (%s)", m.selected.AssigneeEmail))
		}
		b.WriteString("\n")
	} else {
		b.WriteString("👤 Assignee: Unassigned\n")
	}

	b.WriteString(fmt.Sprintf("Created: %s\n", m.selected.Created))
	b.WriteString("\n" + strings.Repeat("─", 60) + "\n")

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

	header := b.String()

	// Show loading or content using viewport for scrolling
	var content string
	if m.loading {
		content = loadingStyle.Render("Loading...")
	} else if m.renderedDesc != "" {
		// Use viewport for scrolling
		content = m.viewport.View()
	} else {
		content = helpStyle.Render("(no description)")
	}

	// Help text
	help := helpStyle.Render("↑/↓/pgup/pgdown: scroll • s: status • u: owners • q/esc: back")

	return header + content + "\n\n" + help
}

func (m IssueListModel) statusChangeView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf(" %s ", m.selected.Key)))
	b.WriteString("\n\n")
	b.WriteString(selectedStyle.Render("Change Status"))
	b.WriteString(fmt.Sprintf(" (current: %s)\n\n", m.selected.Status))

	if m.loading {
		b.WriteString(loadingStyle.Render("Loading available transitions..."))
	} else if len(m.transitions) == 0 {
		b.WriteString("No available transitions\n")
	} else {
		for i, t := range m.transitions {
			if i == m.transitionIndex {
				b.WriteString(selectedActionStyle.Render(fmt.Sprintf("▸ %s", t.Name)))
			} else {
				b.WriteString(fmt.Sprintf("  %s", t.Name))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: navigate • enter: select • esc: back"))

	return b.String()
}

func (m IssueListModel) assigneeChangeView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf(" %s ", m.selected.Key)))
	b.WriteString("\n\n")
	b.WriteString(selectedStyle.Render("Change Assignee"))
	if m.selected.Assignee != "" {
		b.WriteString(fmt.Sprintf(" (current: %s)", m.selected.Assignee))
	}
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(loadingStyle.Render("Updating assignee..."))
	} else {
		b.WriteString(selectedStyle.Render("Type to search or enter email:"))
		b.WriteString("\n\n")

		inputStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333333")).
			Padding(1, 2).
			Width(50)

		b.WriteString(inputStyle.Render("> " + m.assigneeInput + "_"))
		b.WriteString("\n\n")

		if len(m.searchResults) > 0 {
			b.WriteString("Matching users:\n")
			for i, user := range m.searchResults {
				prefix := "  "
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
				if i == m.searchIndex {
					prefix = "▸ "
					style = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#00C851")).
						Bold(true)
				}
				displayName := user.DisplayName
				if user.Email != "" {
					displayName = fmt.Sprintf("%s (%s)", user.DisplayName, user.Email)
				}
				b.WriteString(style.Render(fmt.Sprintf("%s%s\n", prefix, displayName)))
			}
			b.WriteString("\n")
		} else if m.assigneeInput != "" {
			b.WriteString("Press Enter to assign, or type full email\n")
		}
	}

	b.WriteString(helpStyle.Render("↑/↓: navigate • Enter: select • Esc: cancel"))

	return b.String()
}

func (m IssueListModel) manageOwnersView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf(" %s ", m.selected.Key)))
	b.WriteString("\n\n")
	b.WriteString(selectedStyle.Render("Manage Owners"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(loadingStyle.Render("Updating owners..."))
	} else {
		if len(m.selected.Owners) > 0 {
			b.WriteString("Current owners (↑/↓ to select, Enter to remove):\n\n")
			for i, owner := range m.selected.Owners {
				prefix := "  "
				style := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#808080")).
					Width(m.width - 4)
				if m.selectedOwnerIndex == i {
					prefix = "▸ "
					style = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#FF4444")).
						Bold(true).
						Width(m.width - 4)
				}
				b.WriteString(style.Render(fmt.Sprintf("%s%s", prefix, owner)))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		} else {
			b.WriteString("No owners assigned.\n\n")
		}

		b.WriteString(selectedStyle.Render("Type to search and add owner:"))
		b.WriteString("\n\n")

		inputStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333333")).
			Padding(1, 2).
			Width(50)

		b.WriteString(inputStyle.Render("> " + m.ownerSearchInput + "_"))
		b.WriteString("\n\n")

		if len(m.ownerSearchResults) > 0 {
			b.WriteString("Matching users (Enter to add):\n")
			for i, user := range m.ownerSearchResults {
				prefix := "  "
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
				idx := len(m.selected.Owners) + i
				if m.selectedOwnerIndex == idx {
					prefix = "▸ "
					style = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#00C851")).
						Bold(true)
				}
				displayName := user.DisplayName
				if user.Email != "" {
					displayName = fmt.Sprintf("%s (%s)", user.DisplayName, user.Email)
				}
				b.WriteString(style.Render(fmt.Sprintf("%s%s\n", prefix, displayName)))
			}
			b.WriteString("\n")
		} else if m.ownerSearchInput != "" {
			b.WriteString("Press Enter to add as owner, or type full email\n")
		}
	}

	b.WriteString(helpStyle.Render("↑/↓: navigate • Enter: add/remove • Esc: cancel • Backspace: delete"))

	return b.String()
}

func (m IssueListModel) addCommentView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf(" %s ", m.selected.Key)))
	b.WriteString("\n\n")
	b.WriteString(selectedStyle.Render("Add Comment"))
	b.WriteString("\n\n")

	b.WriteString(m.commentInput.View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("ctrl+s: save • esc: cancel"))

	return b.String()
}

func (m IssueListModel) sprintAssignView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf(" %s ", m.selected.Key)))
	b.WriteString("\n\n")
	b.WriteString(selectedStyle.Render("Put in Sprint"))
	if m.selected.SprintName != "" {
		b.WriteString(fmt.Sprintf(" (current: %s)", m.selected.SprintName))
	}
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(loadingStyle.Render("Loading available sprints..."))
	} else if len(m.sprints) == 0 {
		b.WriteString("No open sprints available\n")
	} else {
		for i, sprint := range m.sprints {
			if i == m.sprintIndex {
				b.WriteString(selectedActionStyle.Render(fmt.Sprintf("▸ %s", sprint.Name)))
			} else {
				b.WriteString(fmt.Sprintf("  %s", sprint.Name))
			}
			// Add dates if available
			if !sprint.StartDate.IsZero() && !sprint.EndDate.IsZero() {
				b.WriteString(fmt.Sprintf(" (%s - %s)", sprint.StartDate.Time().Format("Jan 2"), sprint.EndDate.Time().Format("Jan 2")))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: navigate • enter: select • esc: back"))

	return b.String()
}
