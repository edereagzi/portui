package tui

import (
	"fmt"
	"strconv"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"

	"github.com/edereagzi/portui/internal/types"
)

func selectedPortEntry(entries []types.PortEntry, row table.Row) *types.PortEntry {
	if len(row) < 3 {
		return nil
	}
	port, err := strconv.Atoi(row[0])
	if err != nil {
		return nil
	}
	pid, err := strconv.Atoi(row[2])
	if err != nil {
		return nil
	}
	for i := range entries {
		if int(entries[i].Port) == port && int(entries[i].PID) == pid {
			return &entries[i]
		}
	}
	return nil
}

func confirmKillView(entry *types.PortEntry, width int) string {
	if entry == nil {
		return ""
	}
	text := fmt.Sprintf("Kill %s (PID %d) on port %d? [y]es / [n]o", entry.ProcessName, entry.PID, entry.Port)
	rendered := ConfirmDialogStyle.Render(text)
	if width > 0 {
		rendered = lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(rendered)
	}
	return rendered
}
