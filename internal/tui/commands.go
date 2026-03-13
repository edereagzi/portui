package tui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/edereagzi/portui/internal/types"
)

const scanTimeout = 10 * time.Second

type portsLoadedMsg struct {
	entries []types.PortEntry
}

type errMsg struct {
	err error
}

type killResultMsg struct {
	entry types.PortEntry
	err   error
}

func killProcessCmd(svc types.ProcessService, entry types.PortEntry) tea.Cmd {
	return func() tea.Msg {
		err := svc.Kill(entry.PID)
		return killResultMsg{entry: entry, err: err}
	}
}

func scanPortsCmd(scanner types.PortScanner) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
		defer cancel()

		entries, err := scanner.Scan(ctx)
		if err != nil {
			return errMsg{err: err}
		}

		return portsLoadedMsg{entries: entries}
	}
}
