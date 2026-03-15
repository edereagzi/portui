package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/edereagzi/portui/internal/types"
)

type mockScanner struct {
	entries []types.PortEntry
	err     error
}

func (m mockScanner) Scan(context.Context) ([]types.PortEntry, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.entries, nil
}

type mockProcessService struct{}

func (mockProcessService) GetInfo(pid int32) (*types.ProcessInfo, error) {
	return &types.ProcessInfo{PID: pid}, nil
}

func (mockProcessService) Kill(pid int32) error {
	return nil
}

type mockProcessServiceInfoErr struct{}

func (mockProcessServiceInfoErr) GetInfo(pid int32) (*types.ProcessInfo, error) {
	return nil, errors.New("permission denied")
}

func (mockProcessServiceInfoErr) Kill(pid int32) error {
	return nil
}

func testEntries() []types.PortEntry {
	return []types.PortEntry{
		{
			Port:        8080,
			Protocol:    "tcp",
			PID:         1234,
			ProcessName: "api-server",
			User:        "eray",
			LocalAddr:   "127.0.0.1",
		},
		{
			Port:        5432,
			Protocol:    "tcp",
			PID:         4321,
			ProcessName: "postgres",
			User:        "postgres",
			LocalAddr:   "0.0.0.0",
		},
	}
}

func newTestModel() Model {
	return New(mockScanner{}, mockProcessService{})
}

func TestModelInitReturnsCommand(t *testing.T) {
	m := New(mockScanner{entries: testEntries()}, mockProcessService{})

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected init command to be non-nil")
	}
}

func TestModelUpdateWindowSize(t *testing.T) {
	m := newTestModel()

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if cmd != nil {
		t.Fatal("expected window size update to return nil command")
	}

	got, ok := updated.(Model)
	if !ok {
		t.Fatal("expected updated model type")
	}

	if got.width != 120 || got.height != 40 {
		t.Fatalf("expected dimensions 120x40, got %dx%d", got.width, got.height)
	}
}

func TestModelPortsLoadedTransitionsToTableState(t *testing.T) {
	m := newTestModel()

	updated, _ := m.Update(portsLoadedMsg{entries: testEntries()})
	got := updated.(Model)

	if got.state != stateTable {
		t.Fatalf("expected state %v, got %v", stateTable, got.state)
	}
}

func TestModelPortsLoadedPopulatesEntriesAndRows(t *testing.T) {
	entries := testEntries()
	m := newTestModel()

	updated, _ := m.Update(portsLoadedMsg{entries: entries})
	got := updated.(Model)

	if len(got.entries) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(got.entries))
	}

	rows := got.table.Rows()
	if len(rows) != len(entries) {
		t.Fatalf("expected %d table rows, got %d", len(entries), len(rows))
	}

	if rows[0][0] != "5432" || rows[0][2] != "postgres" {
		t.Fatalf("unexpected first row (expected sorted by port): %#v", rows[0])
	}
}

func TestModelErrMsgSetsError(t *testing.T) {
	wantErr := errors.New("scan failed")
	m := newTestModel()

	updated, _ := m.Update(errMsg{err: wantErr})
	got := updated.(Model)

	if !errors.Is(got.err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, got.err)
	}
}

func TestModelQuitKeyReturnsQuitCommand(t *testing.T) {
	m := newTestModel()

	_, cmd := m.Update(tea.KeyPressMsg{Text: "q"})
	if cmd == nil {
		t.Fatal("expected quit command")
	}

	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", cmd())
	}
}

func TestModelViewReturnsTeaView(t *testing.T) {
	m := newTestModel()

	v := m.View()
	if !strings.Contains(v.Content, "Loading ports...") {
		t.Fatalf("expected loading view content, got %q", v.Content)
	}
}

func TestModelTableViewIncludesStatusBar(t *testing.T) {
	m := newTestModel()
	updated, _ := m.Update(portsLoadedMsg{entries: testEntries()})
	got := updated.(Model)

	v := got.View()
	if !strings.Contains(v.Content, "2 ports") {
		t.Fatalf("expected port count in view, got %q", v.Content)
	}
	if !strings.Contains(v.Content, "q: quit") {
		t.Fatalf("expected footer help in view, got %q", v.Content)
	}
}

func TestModelGetInfoErrorShowsActionableHint(t *testing.T) {
	m := New(mockScanner{}, mockProcessServiceInfoErr{})
	updated, _ := m.Update(portsLoadedMsg{entries: testEntries()})
	m = updated.(Model)

	updated2, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := updated2.(Model)

	if !strings.Contains(got.statusMsg, "Try running with sudo") && !strings.Contains(got.statusMsg, "Administrator") {
		t.Fatalf("expected actionable privilege hint in process info error, got %q", got.statusMsg)
	}
	if !got.statusIsErr {
		t.Fatal("expected statusIsErr to be true")
	}
}

func TestModelDetailGuardForWindowsSystemPIDShowsActionableMessage(t *testing.T) {
	origTuiGOOS := currentTuiGOOS
	origErrGOOS := currentErrorsGOOS
	currentTuiGOOS = "windows"
	currentErrorsGOOS = "windows"
	t.Cleanup(func() {
		currentTuiGOOS = origTuiGOOS
		currentErrorsGOOS = origErrGOOS
	})

	entries := []types.PortEntry{{
		Port:        5357,
		Protocol:    "tcp",
		PID:         4,
		ProcessName: "System",
		User:        "NT AUTHORITY",
		LocalAddr:   "0.0.0.0:5357",
	}}

	m := New(mockScanner{}, mockProcessService{})
	updated, _ := m.Update(portsLoadedMsg{entries: entries})
	m = updated.(Model)

	updated2, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := updated2.(Model)

	if got.state != stateTable {
		t.Fatalf("expected stateTable when detail is guarded, got %v", got.state)
	}
	if !strings.Contains(got.statusMsg, "unavailable") {
		t.Fatalf("expected actionable unavailable message, got %q", got.statusMsg)
	}
	if !got.statusIsErr {
		t.Fatal("expected statusIsErr to be true")
	}
}
