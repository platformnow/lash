package ui

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	table    table.Model
	progress progress.Model
	state    State
	err      error
}

type State int

const (
	StateInit State = iota
	StateSelectingProviders
	StateInstalling
	StateDone
)

func NewModel() Model {
	return Model{
		table:    initTable(),
		progress: progress.New(progress.WithDefaultGradient()),
		state:    StateInit,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	switch m.state {
	case StateInit:
		return "Loading..."
	case StateSelectingProviders:
		return m.table.View()
	case StateInstalling:
		return m.progress.View()
	case StateDone:
		return "Installation complete!"
	default:
		return "Unknown state"
	}
}
