package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"

	"github.com/edereagzi/portui/internal/types"
)

func buildTable(entries []types.PortEntry, width, height int) table.Model {
	cols := []table.Column{
		{Title: "Port", Width: 8},
		{Title: "Protocol", Width: 8},
		{Title: "PID", Width: 8},
		{Title: "Process", Width: 20},
		{Title: "User", Width: 12},
		{Title: "Bind", Width: 16},
	}

	rows := make([]table.Row, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", entry.Port),
			entry.Protocol,
			fmt.Sprintf("%d", entry.PID),
			entry.ProcessName,
			entry.User,
			bindHost(entry.LocalAddr),
		})
	}

	styles := table.DefaultStyles()
	styles.Header = TableHeaderStyle
	styles.Cell = TableRowStyle
	styles.Selected = SelectedRowStyle

	tableHeight := 8
	if height > 0 {
		tableHeight = height - 4
	}
	if tableHeight < 3 {
		tableHeight = 3
	}

	opts := []table.Option{
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
		table.WithStyles(styles),
	}
	if width > 0 {
		opts = append(opts, table.WithWidth(width))
	}

	return table.New(opts...)
}

func bindHost(localAddr string) string {
	addr := strings.TrimSpace(localAddr)
	if addr == "" {
		return ""
	}

	if strings.HasPrefix(addr, "*:") || addr == "*" {
		return "0.0.0.0"
	}
	if strings.HasPrefix(addr, ":::") {
		return "::"
	}

	if strings.HasPrefix(addr, "[") {
		if end := strings.Index(addr, "]"); end > 1 {
			return addr[1:end]
		}
	}

	if idx := strings.LastIndex(addr, ":"); idx > 0 {
		host := addr[:idx]
		if host == "*" {
			return "0.0.0.0"
		}
		return host
	}

	return addr
}
