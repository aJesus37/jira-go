// internal/tui/ceremony.go
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/jira-go/internal/models"
)

// CeremonyType represents the type of ceremony
type CeremonyType int

const (
	Planning CeremonyType = iota
	Retrospective
	DailyStandup
)

// CeremonyModel is the TUI model for ceremonies
type CeremonyModel struct {
	ceremonyType  CeremonyType
	issues        []models.Issue
	cards         []RetroCard
	updates       []StandupUpdate
	timer         timer.Model
	activeTimer   bool
	timerDuration time.Duration

	// For planning
	backlogList list.Model
	sprintList  list.Model
	selectedTab int // 0 = backlog, 1 = sprint

	// For retro
	currentColumn int // 0 = went well, 1 = improve, 2 = action items
	cardInput     textarea.Model
	addingCard    bool

	// For standup
	currentMember int
	updateInput   textarea.Model
	addingUpdate  bool

	width  int
	height int
}

// RetroCard represents a card in retrospective
type RetroCard struct {
	ID      int
	Content string
	Column  int // 0 = went well, 1 = improve, 2 = action items
	Votes   int
	Author  string
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

// NewRetrospectiveCeremony creates a new retrospective ceremony TUI
func NewRetrospectiveCeremony() CeremonyModel {
	input := textarea.New()
	input.Placeholder = "Enter your card..."
	input.SetWidth(50)
	input.SetHeight(3)

	return CeremonyModel{
		ceremonyType:  Retrospective,
		cards:         []RetroCard{},
		cardInput:     input,
		currentColumn: 0,
	}
}

// NewDailyStandupCeremony creates a new daily standup ceremony TUI
func NewDailyStandupCeremony(members []string) CeremonyModel {
	input := textarea.New()
	input.Placeholder = "What did you do yesterday? What's your plan for today? Any blockers?"
	input.SetWidth(60)
	input.SetHeight(5)

	var updates []StandupUpdate
	for _, member := range members {
		updates = append(updates, StandupUpdate{Name: member})
	}

	t := timer.NewWithInterval(2*time.Minute, time.Second)

	return CeremonyModel{
		ceremonyType:  DailyStandup,
		updates:       updates,
		updateInput:   input,
		currentMember: 0,
		timer:         t,
		activeTimer:   false,
		timerDuration: 2 * time.Minute,
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
		case Retrospective:
			return m.updateRetrospective(msg)
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

func (m CeremonyModel) updateRetrospective(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.addingCard {
		switch msg.String() {
		case "esc":
			m.addingCard = false
		case "enter":
			// Add card
			content := m.cardInput.Value()
			if content != "" {
				m.cards = append(m.cards, RetroCard{
					ID:      len(m.cards),
					Content: content,
					Column:  m.currentColumn,
					Votes:   0,
				})
				m.cardInput.SetValue("")
			}
			m.addingCard = false
		default:
			var cmd tea.Cmd
			m.cardInput, cmd = m.cardInput.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	if msg.String() == "q" || msg.String() == "esc" {
		return m, tea.Quit
	}

	switch msg.String() {
	case "1":
		m.currentColumn = 0
	case "2":
		m.currentColumn = 1
	case "3":
		m.currentColumn = 2
	case "a":
		m.addingCard = true
		m.cardInput.Focus()
	case "v":
		// Vote on card under cursor
	case "e":
		// Export retrospective
		return m, tea.Quit
	}

	return m, nil
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
	case "p":
		// Previous member
		m.currentMember--
		if m.currentMember < 0 {
			m.currentMember = len(m.updates) - 1
		}
	case "a":
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
	case Retrospective:
		return m.viewRetrospective()
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

func (m CeremonyModel) viewRetrospective() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" 📝 Retrospective "))
	b.WriteString("\n\n")

	// Show columns
	columns := []string{"✅ Went Well", "⚠️  Improve", "🎯 Action Items"}

	for i, colName := range columns {
		style := columnStyle
		if i == m.currentColumn {
			style = selectedColumnStyle
		}

		// Count cards in this column
		cardCount := 0
		for _, card := range m.cards {
			if card.Column == i {
				cardCount++
			}
		}

		colTitle := fmt.Sprintf("%s (%d)", colName, cardCount)

		var cardsContent strings.Builder
		for _, card := range m.cards {
			if card.Column == i {
				cardsContent.WriteString(fmt.Sprintf("• %s", card.Content))
				if card.Votes > 0 {
					cardsContent.WriteString(fmt.Sprintf(" (+%d)", card.Votes))
				}
				cardsContent.WriteString("\n")
			}
		}

		colContent := style.Width(25).Render(colTitle + "\n" + strings.Repeat("─", 23) + "\n" + cardsContent.String())
		b.WriteString(colContent)
		b.WriteString("  ")
	}

	b.WriteString("\n\n")

	// Show card input if adding
	if m.addingCard {
		b.WriteString("Add card to " + columns[m.currentColumn] + ":\n")
		b.WriteString(m.cardInput.View())
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("1/2/3: select column • a: add card • v: vote • e: export • q: quit"))

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

	// Show team checklist
	b.WriteString("Team Members:\n")
	for i, member := range m.updates {
		prefix := "  ○ "
		if i == m.currentMember {
			prefix = "  ● "
		}
		if member.Today != "" {
			prefix = "  ✓ "
		}
		b.WriteString(fmt.Sprintf("%s%s\n", prefix, member.Name))
	}

	b.WriteString("\n")

	// Show current member
	if m.currentMember < len(m.updates) {
		member := m.updates[m.currentMember]
		b.WriteString(fmt.Sprintf("Current: %s\n", member.Name))
		b.WriteString(strings.Repeat("─", 40) + "\n")

		if member.Today != "" {
			b.WriteString(fmt.Sprintf("Today's update:\n%s\n\n", member.Today))
		}

		if m.addingUpdate {
			b.WriteString("Enter update:\n")
			b.WriteString(m.updateInput.View())
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("n: next • p: previous • a: add update • t: toggle timer • e: export • q: quit"))

	return b.String()
}

func (m CeremonyModel) GetTitle() string {
	switch m.ceremonyType {
	case Planning:
		return "Sprint Planning"
	case Retrospective:
		return "Retrospective"
	case DailyStandup:
		return "Daily Standup"
	default:
		return "Ceremony"
	}
}

// ExportPlanning exports planning results as markdown
func (m CeremonyModel) ExportPlanning() string {
	var b strings.Builder
	b.WriteString("# Sprint Planning\n\n")
	b.WriteString(fmt.Sprintf("Date: %s\n\n", time.Now().Format("2006-01-02")))

	b.WriteString("## Sprint Issues\n\n")
	for _, item := range m.sprintList.Items() {
		if issue, ok := item.(IssueItem); ok {
			b.WriteString(fmt.Sprintf("- **%s**: %s\n", issue.issue.Key, issue.issue.Summary))
		}
	}

	return b.String()
}

// ExportRetrospective exports retrospective results as markdown
func (m CeremonyModel) ExportRetrospective() string {
	var b strings.Builder
	b.WriteString("# Retrospective\n\n")
	b.WriteString(fmt.Sprintf("Date: %s\n\n", time.Now().Format("2006-01-02")))

	columns := []string{"Went Well", "To Improve", "Action Items"}
	for i, colName := range columns {
		b.WriteString(fmt.Sprintf("## %s\n\n", colName))
		for _, card := range m.cards {
			if card.Column == i {
				b.WriteString(fmt.Sprintf("- %s", card.Content))
				if card.Votes > 0 {
					b.WriteString(fmt.Sprintf(" (%d votes)", card.Votes))
				}
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// ExportDailyStandup exports standup results as markdown
func (m CeremonyModel) ExportDailyStandup() string {
	var b strings.Builder
	b.WriteString("# Daily Standup\n\n")
	b.WriteString(fmt.Sprintf("Date: %s\n\n", time.Now().Format("2006-01-02")))

	for _, member := range m.updates {
		b.WriteString(fmt.Sprintf("## %s\n\n", member.Name))
		if member.Today != "" {
			b.WriteString(fmt.Sprintf("**Today:** %s\n\n", member.Today))
		}
		if member.Blockers != "" {
			b.WriteString(fmt.Sprintf("**Blockers:** %s\n\n", member.Blockers))
		}
	}

	return b.String()
}
