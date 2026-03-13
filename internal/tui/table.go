package tui

import (
	"fmt"

	"charm.land/bubbles/v2/table"

	"github.com/edereagzi/portui/internal/types"
)

func buildTable(entries []types.PortEntry, width, height int) table.Model {
	cols := []table.Column{
		{Title: "Port", Width: 8},
		{Title: "Protocol", Width: 6},
		{Title: "PID", Width: 8},
		{Title: "Process", Width: 20},
		{Title: "User", Width: 12},
		{Title: "Address", Width: 20},
	}

	rows := make([]table.Row, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", entry.Port),
			entry.Protocol,
			fmt.Sprintf("%d", entry.PID),
			entry.ProcessName,
			entry.User,
			entry.LocalAddr,
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
