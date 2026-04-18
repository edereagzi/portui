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

type processInfoLoadedMsg struct {
	reqID int64
	pid   int32
	info  *types.ProcessInfo
	err   error
}

func killProcessCmd(svc types.ProcessService, entry types.PortEntry) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := svc.Kill(ctx, entry.PID)
		return killResultMsg{entry: entry, err: err}
	}
}

func loadProcessInfoCmd(svc types.ProcessService, reqID int64, pid int32) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		info, err := svc.GetInfo(ctx, pid)
		return processInfoLoadedMsg{reqID: reqID, pid: pid, info: info, err: err}
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
