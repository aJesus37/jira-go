// internal/tui/tui.go
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model is the base interface for all TUI models
type Model interface {
	tea.Model
	GetTitle() string
}

// Run starts a TUI program
func Run(model tea.Model) error {
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
