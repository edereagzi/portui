package tui

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/edereagzi/portui/internal/types"
)

func formatMemory(bytes uint64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)

	if bytes >= gb {
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	}
	if bytes >= mb {
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	}
	if bytes == 0 {
		return "0 KB"
	}
	return fmt.Sprintf("%d KB", (bytes+kb-1)/kb)
}

func formatUptime(createTimeMs int64) string {
	if createTimeMs <= 0 {
		return "0m 0s"
	}

	age := time.Since(time.UnixMilli(createTimeMs))
	if age < 0 {
		age = 0
	}

	if age < time.Hour {
		minutes := int(age / time.Minute)
		seconds := int((age % time.Minute) / time.Second)
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	if age < 24*time.Hour {
		hours := int(age / time.Hour)
		minutes := int((age % time.Hour) / time.Minute)
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}

	days := int(age / (24 * time.Hour))
	hours := int((age % (24 * time.Hour)) / time.Hour)
	return fmt.Sprintf("%dd %dh", days, hours)
}

func renderDetailPanel(info *types.ProcessInfo, entry *types.PortEntry, width int) string {
	if info == nil || entry == nil {
		content := "Process Detail\n\nNo process selected."
		style := DetailPanelStyle
		if width > 0 {
			style = style.Width(width)
		}
		return style.Render(content)
	}

	style := DetailPanelStyle
	if width > 0 {
		style = style.Width(width)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		"Process Detail",
		fmt.Sprintf("%-8s %s", "Name:", info.Name),
		fmt.Sprintf("%-8s %d", "PID:", info.PID),
		fmt.Sprintf("%-8s %s", "Command:", info.Cmdline),
		fmt.Sprintf("%-8s %s", "User:", info.User),
		fmt.Sprintf("%-8s %s", "Memory:", formatMemory(info.MemoryRSS)),
		fmt.Sprintf("%-8s %d (%s)", "Port:", entry.Port, entry.Protocol),
		fmt.Sprintf("%-8s %s", "Address:", entry.LocalAddr),
		fmt.Sprintf("%-8s %s", "Uptime:", formatUptime(info.CreateTime)),
	)

	return style.Render(content)
}
