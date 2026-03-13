package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/edereagzi/portui/internal/types"
)

func Run(scanner types.PortScanner, processService types.ProcessService) error {
	m := New(scanner, processService)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
