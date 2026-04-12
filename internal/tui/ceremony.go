// internal/tui/ceremony.go
package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aJesus37/jira-go/internal/models"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedColumnStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00C851")).
		Padding(1)
)

// CeremonyType represents the type of ceremony
type CeremonyType int

const (
	Planning CeremonyType = iota
	DailyStandup
)

// CeremonyModel is the TUI model for ceremonies
type CeremonyModel struct {
	ceremonyType  CeremonyType
	issues        []models.Issue
	updates       []StandupUpdate
	assigneeTasks map[string]map[string][]models.Issue // member -> status -> tasks
	timer         timer.Model
	activeTimer   bool
	timerDuration time.Duration

	// For planning
	backlogList list.Model
	sprintList  list.Model
	selectedTab int // 0 = backlog, 1 = sprint

	// For standup
	currentMember int
	updateInput   textarea.Model
	addingUpdate  bool
	selectedView  int // 0 = tasks view, 1 = update form

	width  int
	height int
}

// StandupUpdate represents a team member's update
type StandupUpdate struct {
	Name      string
	Yesterday string
	Today     string
	Blockers  string
}

// NewPlanningCeremony creates a new sprint planning ceremony TUI
func NewPlanningCeremony(issues []models.Issue) CeremonyModel {
	var backlogItems []list.Item
	var sprintItems []list.Item

	for _, issue := range issues {
		backlogItems = append(backlogItems, IssueItem{issue: Issue{
			Key:      issue.Key,
			Summary:  issue.Summary,
			Type:     issue.Type,
			Status:   issue.Status,
			Assignee: getAssigneeName(issue.Assignee),
		}})
	}

	backlogList := list.New(backlogItems, list.NewDefaultDelegate(), 40, 15)
	backlogList.Title = "📋 Backlog"
	backlogList.SetShowStatusBar(true)

	sprintList := list.New(sprintItems, list.NewDefaultDelegate(), 40, 15)
	sprintList.Title = "🏃 Sprint"
	sprintList.SetShowStatusBar(true)

	return CeremonyModel{
		ceremonyType: Planning,
		issues:       issues,
		backlogList:  backlogList,
		sprintList:   sprintList,
		selectedTab:  0,
	}
}

// NewDailyStandupCeremony creates a new daily standup ceremony TUI
func NewDailyStandupCeremony(members []string, assigneeTasks map[string]map[string][]models.Issue, timerDuration time.Duration) CeremonyModel {
	input := textarea.New()
	input.Placeholder = "What did you do yesterday? What's your plan for today? Any blockers?"
	input.SetWidth(60)
	input.SetHeight(5)

	var updates []StandupUpdate
	for _, member := range members {
		updates = append(updates, StandupUpdate{Name: member})
	}

	// Default to 2 minutes if not specified
	if timerDuration <= 0 {
		timerDuration = 2 * time.Minute
	}

	t := timer.NewWithInterval(timerDuration, time.Second)

	return CeremonyModel{
		ceremonyType:  DailyStandup,
		updates:       updates,
		assigneeTasks: assigneeTasks,
		updateInput:   input,
		currentMember: 0,
		selectedView:  0,
		timer:         t,
		activeTimer:   false,
		timerDuration: timerDuration,
	}
}

func getAssigneeName(assignee *models.User) string {
	if assignee == nil {
		return ""
	}
	return assignee.DisplayName
}

func (m CeremonyModel) Init() tea.Cmd {
	if m.ceremonyType == DailyStandup && m.activeTimer {
		return m.timer.Init()
	}
	return nil
}

func (m CeremonyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle ceremony-specific keys
		switch m.ceremonyType {
		case Planning:
			return m.updatePlanning(msg)
		case DailyStandup:
			return m.updateDailyStandup(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case timer.TickMsg:
		if m.ceremonyType == DailyStandup && m.activeTimer {
			var cmd tea.Cmd
			m.timer, cmd = m.timer.Update(msg)
			return m, cmd
		}

	case timer.TimeoutMsg:
		if m.ceremonyType == DailyStandup {
			m.activeTimer = false
			// Move to next member
			m.currentMember++
			if m.currentMember >= len(m.updates) {
				m.currentMember = 0
			}
		}
	}

	return m, nil
}

func (m CeremonyModel) updatePlanning(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "q" || msg.String() == "esc" {
		return m, tea.Quit
	}

	switch msg.String() {
	case "tab", "right":
		m.selectedTab = 1
	case "left":
		m.selectedTab = 0
	case "enter", " ":
		// Move issue between lists
		if m.selectedTab == 0 {
			// Move from backlog to sprint
			if item, ok := m.backlogList.SelectedItem().(IssueItem); ok {
				// Add to sprint
				m.sprintList.InsertItem(len(m.sprintList.Items()), item)
				// Remove from backlog
				m.backlogList.RemoveItem(m.backlogList.Index())
			}
		} else {
			// Move from sprint to backlog
			if item, ok := m.sprintList.SelectedItem().(IssueItem); ok {
				m.backlogList.InsertItem(len(m.backlogList.Items()), item)
				m.sprintList.RemoveItem(m.sprintList.Index())
			}
		}
	case "s":
		// Export sprint
		return m, tea.Quit
	}

	// Update the active list
	var cmd tea.Cmd
	if m.selectedTab == 0 {
		m.backlogList, cmd = m.backlogList.Update(msg)
	} else {
		m.sprintList, cmd = m.sprintList.Update(msg)
	}
	return m, cmd
}

func (m CeremonyModel) updateDailyStandup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.addingUpdate {
		switch msg.String() {
		case "esc":
			m.addingUpdate = false
		case "enter":
			// Save update
			update := m.updateInput.Value()
			if m.currentMember < len(m.updates) {
				m.updates[m.currentMember].Today = update
			}
			m.updateInput.SetValue("")
			m.addingUpdate = false
			// Move to next member
			m.currentMember++
			if m.currentMember >= len(m.updates) {
				m.currentMember = 0
			}
		default:
			var cmd tea.Cmd
			m.updateInput, cmd = m.updateInput.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	if msg.String() == "q" || msg.String() == "esc" {
		return m, tea.Quit
	}

	switch msg.String() {
	case "n":
		// Next member
		m.currentMember++
		if m.currentMember >= len(m.updates) {
			m.currentMember = 0
		}
		m.selectedView = 0
	case "p":
		// Previous member
		m.currentMember--
		if m.currentMember < 0 {
			m.currentMember = len(m.updates) - 1
		}
		m.selectedView = 0
	case "tab":
		// Toggle between task view and update view
		if m.selectedView == 0 {
			m.selectedView = 1
		} else {
			m.selectedView = 0
		}
	case "a":
		m.selectedView = 1
		m.addingUpdate = true
		m.updateInput.Focus()
	case "t":
		// Toggle timer
		m.activeTimer = !m.activeTimer
		if m.activeTimer {
			return m, m.timer.Init()
		}
	case "e":
		// Export standup
		return m, tea.Quit
	}

	return m, nil
}

func (m CeremonyModel) View() string {
	switch m.ceremonyType {
	case Planning:
		return m.viewPlanning()
	case DailyStandup:
		return m.viewDailyStandup()
	default:
		return "Unknown ceremony type"
	}
}

func (m CeremonyModel) viewPlanning() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" 🎯 Sprint Planning "))
	b.WriteString("\n\n")

	// Show tabs
	tabStyle := lipgloss.NewStyle().Padding(0, 2)
	activeTabStyle := tabStyle.Background(lipgloss.Color("#7D56F4")).Foreground(lipgloss.Color("#FFFFFF"))

	backlogTab := tabStyle.Render("📋 Backlog")
	if m.selectedTab == 0 {
		backlogTab = activeTabStyle.Render("📋 Backlog")
	}

	sprintTab := tabStyle.Render("🏃 Sprint")
	if m.selectedTab == 1 {
		sprintTab = activeTabStyle.Render("🏃 Sprint")
	}

	b.WriteString(backlogTab + " | " + sprintTab)
	b.WriteString("\n\n")

	// Show lists side by side
	backlogView := columnStyle.Render(m.backlogList.View())
	sprintView := columnStyle.Render(m.sprintList.View())

	if m.selectedTab == 0 {
		backlogView = selectedColumnStyle.Render(m.backlogList.View())
	} else {
		sprintView = selectedColumnStyle.Render(m.sprintList.View())
	}

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, backlogView, sprintView))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("tab/←/→: switch lists • enter/space: move issue • s: export • q: quit"))

	return b.String()
}

func (m CeremonyModel) viewDailyStandup() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" 📅 Daily Standup "))
	b.WriteString("\n\n")

	// Show timer if active
	if m.activeTimer {
		b.WriteString(fmt.Sprintf("⏱️  Time remaining: %s\n\n", m.timer.View()))
	}

	if m.currentMember >= len(m.updates) {
		return b.String()
	}

	member := m.updates[m.currentMember]

	// Two-column layout: sidebar + main content
	sidebarWidth := 25
	mainWidth := m.width - sidebarWidth - 5
	if mainWidth < 40 {
		mainWidth = 40
	}

	// Build sidebar with team members
	var sidebar strings.Builder
	sidebar.WriteString(lipgloss.NewStyle().Bold(true).Render("Team"))
	sidebar.WriteString("\n\n")
	for i, u := range m.updates {
		prefix := "○ "
		style := lipgloss.NewStyle()
		if i == m.currentMember {
			prefix = "● "
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00C851"))
		}
		if u.Today != "" {
			prefix = "✓ "
		}
		sidebar.WriteString(style.Render(fmt.Sprintf("%s%s\n", prefix, u.Name)))
	}

	sidebarStyle := lipgloss.NewStyle().
		Width(sidebarWidth).
		PaddingRight(2)

	// Build main content - either tasks view or update form
	var mainContent strings.Builder

	if m.selectedView == 1 || m.addingUpdate {
		// Update form view
		mainContent.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("📝 Update for %s", member.Name)))
		mainContent.WriteString("\n")
		mainContent.WriteString(strings.Repeat("─", mainWidth))
		mainContent.WriteString("\n\n")

		if member.Today != "" {
			mainContent.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Render("Previous update:"))
			mainContent.WriteString("\n")
			mainContent.WriteString(member.Today)
			mainContent.WriteString("\n\n")
		}

		mainContent.WriteString(m.updateInput.View())
	} else {
		// Tasks view
		mainContent.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("👤 %s's Tasks", member.Name)))
		mainContent.WriteString("\n")
		mainContent.WriteString(strings.Repeat("─", mainWidth))
		mainContent.WriteString("\n")

		// Show tasks grouped by status
		tasksByStatus := m.assigneeTasks[member.Name]

		// Known status patterns for ordering (case-insensitive)
		statusPriority := map[string]int{
			"done":        1,
			"closed":      1,
			"review":      2,
			"in review":   2,
			"in progress": 3,
			"progress":    3,
			"blocked":     4,
			"to do":       5,
			"todo":        5,
			"open":        5,
			"new":         6,
		}

		// Default colors for common statuses
		statusColors := map[string]string{
			"done":        "#00C851",
			"closed":      "#00C851",
			"review":      "#FFA500",
			"in review":   "#FFA500",
			"in progress": "#00A8E8",
			"progress":    "#00A8E8",
			"blocked":     "#FF4444",
			"to do":       "#808080",
			"todo":        "#808080",
			"open":        "#808080",
			"new":         "#808080",
		}

		// Get unique statuses and sort them
		var statuses []string
		for status := range tasksByStatus {
			statuses = append(statuses, status)
		}
		sort.SliceStable(statuses, func(i, j int) bool {
			si, sji := strings.ToLower(statuses[i]), strings.ToLower(statuses[j])
			pi, oki := statusPriority[si]
			pj, okj := statusPriority[sji]
			if oki && okj {
				if pi != pj {
					return pi < pj
				}
			} else if oki {
				return true
			} else if okj {
				return false
			}
			return si < sji
		})

		hasTasks := false
		for _, status := range statuses {
			tasks := tasksByStatus[status]
			if len(tasks) == 0 {
				continue
			}
			hasTasks = true

			color := statusColors[strings.ToLower(status)]
			if color == "" {
				color = "#808080"
			}

			statusStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(color))

			mainContent.WriteString(fmt.Sprintf("\n%s (%d)\n", statusStyle.Render(status), len(tasks)))

			for _, task := range tasks {
				taskStyle := lipgloss.NewStyle().
					Width(mainWidth - 2)

				keyStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#a277ff"))

				taskLine := fmt.Sprintf("  %s %s",
					keyStyle.Render(task.Key),
					task.Summary)

				mainContent.WriteString(taskStyle.Render(taskLine) + "\n")
			}
		}

		if !hasTasks {
			mainContent.WriteString("\n")
			mainContent.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#808080")).
				Italic(true).
				Render("No tasks assigned in current sprint"))
			mainContent.WriteString("\n")
		}

		// Show current update if exists
		if member.Today != "" {
			mainContent.WriteString("\n")
			mainContent.WriteString(lipgloss.NewStyle().Bold(true).Render("📝 Today's Update"))
			mainContent.WriteString("\n")
			mainContent.WriteString(strings.Repeat("─", mainWidth))
			mainContent.WriteString("\n")
			mainContent.WriteString(member.Today)
			mainContent.WriteString("\n")
		}
	}

	mainStyle := lipgloss.NewStyle().Width(mainWidth)

	// Combine sidebar and main content
	row := lipgloss.JoinHorizontal(lipgloss.Top,
		sidebarStyle.Render(sidebar.String()),
		mainStyle.Render(mainContent.String()),
	)
	b.WriteString(row)
	b.WriteString("\n\n")

	// Help text
	if m.addingUpdate {
		b.WriteString(helpStyle.Render("enter: save • esc: cancel"))
	} else if m.selectedView == 1 {
		b.WriteString(helpStyle.Render("esc: back to tasks • enter: save update"))
	} else {
		b.WriteString(helpStyle.Render("n: next • p: prev • a: add update • tab: toggle view • t: timer • q: quit"))
	}

	return b.String()
}

func (m CeremonyModel) GetTitle() string {
	switch m.ceremonyType {
	case Planning:
		return "Sprint Planning"
	case DailyStandup:
		return "Daily Standup"
	default:
		return "Ceremony"
	}
}
