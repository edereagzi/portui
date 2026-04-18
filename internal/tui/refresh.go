package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

const refreshInterval = 3 * time.Second

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

type statusClearMsg struct {
	generation int64
}

func statusClearCmd(generation int64) tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return statusClearMsg{generation: generation}
	})
}
