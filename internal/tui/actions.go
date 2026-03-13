package tui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

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

func confirmKillView(entry *types.PortEntry, impactPorts []uint16, width int) string {
	if entry == nil {
		return ""
	}
	impact := formatImpactSummary(impactPorts)
	text := fmt.Sprintf("Kill %s (PID %d) on port %d? [y]es / [n]o\n%s", entry.ProcessName, entry.PID, entry.Port, impact)
	rendered := ConfirmDialogStyle.Render(text)
	if width > 0 {
		rendered = lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(rendered)
	}
	return rendered
}

func impactedPortsByPID(entries []types.PortEntry, pid int32) []uint16 {
	seen := make(map[uint16]struct{})
	ports := make([]uint16, 0)
	for _, e := range entries {
		if e.PID != pid {
			continue
		}
		if _, ok := seen[e.Port]; ok {
			continue
		}
		seen[e.Port] = struct{}{}
		ports = append(ports, e.Port)
	}
	sort.Slice(ports, func(i, j int) bool {
		return ports[i] < ports[j]
	})
	return ports
}

func formatImpactSummary(ports []uint16) string {
	count := len(ports)
	if count == 0 {
		return "Affects 0 listening ports"
	}
	if count == 1 {
		return fmt.Sprintf("Affects 1 listening port: %d", ports[0])
	}

	limit := 6
	shown := ports
	if len(shown) > limit {
		shown = shown[:limit]
	}

	parts := make([]string, 0, len(shown))
	for _, p := range shown {
		parts = append(parts, fmt.Sprintf("%d", p))
	}

	more := ""
	if count > limit {
		more = fmt.Sprintf(" (+%d more)", count-limit)
	}

	return fmt.Sprintf("Affects %d listening ports: %s%s", count, strings.Join(parts, ", "), more)
}
