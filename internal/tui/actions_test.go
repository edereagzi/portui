package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/edereagzi/portui/internal/types"
)

type mockProcessServiceKill struct {
	killErr    error
	killedPIDs []int32
}

func (m *mockProcessServiceKill) GetInfo(ctx context.Context, pid int32) (*types.ProcessInfo, error) {
	return &types.ProcessInfo{PID: pid}, nil
}

func (m *mockProcessServiceKill) Kill(ctx context.Context, pid int32) error {
	m.killedPIDs = append(m.killedPIDs, pid)
	return m.killErr
}

func newKillTestModel(svc *mockProcessServiceKill) Model {
	m := New(mockScanner{}, svc)
	updated, _ := m.Update(portsLoadedMsg{entries: testEntries()})
	return updated.(Model)
}

func confirmKillAndCollect(t *testing.T, m Model) (Model, tea.Cmd) {
	t.Helper()
	updated, cmd := m.Update(tea.KeyPressMsg{Text: "y"})
	got := updated.(Model)
	if cmd == nil {
		t.Fatal("expected non-nil cmd from kill confirm")
	}
	msg := cmd()
	updated2, cmd2 := got.Update(msg)
	return updated2.(Model), cmd2
}

func TestKillKeyTransitionsToConfirmState(t *testing.T) {
	svc := &mockProcessServiceKill{}
	m := newKillTestModel(svc)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "x"})
	got := updated.(Model)

	if got.state != stateConfirmKill {
		t.Fatalf("expected stateConfirmKill after x, got %v", got.state)
	}
	if got.selectedEntry == nil {
		t.Fatal("expected selectedEntry to be set after x")
	}
}

func TestKillConfirmYExecutesKill(t *testing.T) {
	svc := &mockProcessServiceKill{}
	m := newKillTestModel(svc)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "x"})
	m = updated.(Model)

	wantPID := m.selectedEntry.PID

	got, _ := confirmKillAndCollect(t, m)

	if len(svc.killedPIDs) == 0 {
		t.Fatal("expected Kill() to be called")
	}
	if svc.killedPIDs[0] != wantPID {
		t.Fatalf("expected Kill(%d), got Kill(%d)", wantPID, svc.killedPIDs[0])
	}
	if got.state != stateTable {
		t.Fatalf("expected stateTable after kill, got %v", got.state)
	}
}

func TestKillConfirmNCancels(t *testing.T) {
	svc := &mockProcessServiceKill{}
	m := newKillTestModel(svc)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "x"})
	m = updated.(Model)

	updated2, _ := m.Update(tea.KeyPressMsg{Text: "n"})
	got := updated2.(Model)

	if got.state != stateTable {
		t.Fatalf("expected stateTable after n, got %v", got.state)
	}
	if len(svc.killedPIDs) != 0 {
		t.Fatalf("expected Kill() NOT to be called, but got %v", svc.killedPIDs)
	}
	if got.selectedEntry != nil {
		t.Fatal("expected selectedEntry to be nil after cancel")
	}
}

func TestKillEscCancels(t *testing.T) {
	svc := &mockProcessServiceKill{}
	m := newKillTestModel(svc)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "x"})
	m = updated.(Model)

	updated2, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	got := updated2.(Model)

	if got.state != stateTable {
		t.Fatalf("expected stateTable after Esc, got %v", got.state)
	}
	if len(svc.killedPIDs) != 0 {
		t.Fatalf("expected Kill() NOT to be called, but got %v", svc.killedPIDs)
	}
	if got.selectedEntry != nil {
		t.Fatal("expected selectedEntry to be nil after Esc cancel")
	}
}

func TestKillSuccessShowsStatusMessage(t *testing.T) {
	svc := &mockProcessServiceKill{}
	m := newKillTestModel(svc)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "x"})
	m = updated.(Model)

	got, _ := confirmKillAndCollect(t, m)

	if !strings.Contains(got.statusMsg, "✓") {
		t.Fatalf("expected statusMsg to contain '✓', got %q", got.statusMsg)
	}
	if got.statusIsErr {
		t.Fatal("expected statusIsErr to be false on success")
	}
}

func TestKillErrorShowsErrorMessage(t *testing.T) {
	svc := &mockProcessServiceKill{killErr: errors.New("operation not permitted")}
	m := newKillTestModel(svc)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "x"})
	m = updated.(Model)

	got, _ := confirmKillAndCollect(t, m)

	if !strings.Contains(got.statusMsg, "✗") {
		t.Fatalf("expected statusMsg to contain '✗', got %q", got.statusMsg)
	}
	if !got.statusIsErr {
		t.Fatal("expected statusIsErr to be true on error")
	}
	if strings.Count(got.statusMsg, "✗") != 1 {
		t.Fatalf("expected single error prefix in status message, got %q", got.statusMsg)
	}
}

func TestKillUsesSelectedRowPID(t *testing.T) {
	svc := &mockProcessServiceKill{}
	entries := []types.PortEntry{
		{Port: 9000, Protocol: "tcp", PID: 7777, ProcessName: "myapp", User: "eray", LocalAddr: "127.0.0.1"},
		{Port: 9001, Protocol: "tcp", PID: 8888, ProcessName: "otherapp", User: "eray", LocalAddr: "127.0.0.1"},
	}
	m := New(mockScanner{}, svc)
	updated, _ := m.Update(portsLoadedMsg{entries: entries})
	m = updated.(Model)

	updated2, _ := m.Update(tea.KeyPressMsg{Text: "x"})
	m = updated2.(Model)

	if m.selectedEntry == nil {
		t.Fatal("expected selectedEntry to be set")
	}
	if m.selectedEntry.PID != 7777 {
		t.Fatalf("expected selectedEntry.PID=7777, got %d", m.selectedEntry.PID)
	}

	confirmKillAndCollect(t, m)

	if len(svc.killedPIDs) == 0 || svc.killedPIDs[0] != 7777 {
		t.Fatalf("expected Kill(7777), got %v", svc.killedPIDs)
	}
}

func TestKillShowsPendingStatus(t *testing.T) {
	svc := &mockProcessServiceKill{}
	m := newKillTestModel(svc)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "x"})
	m = updated.(Model)

	updated2, cmd := m.Update(tea.KeyPressMsg{Text: "y"})
	got := updated2.(Model)

	if !strings.Contains(got.statusMsg, "Killing") {
		t.Fatalf("expected pending status with 'Killing', got %q", got.statusMsg)
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd for async kill")
	}
}

func TestConfirmKillViewShowsImpactSummaryMultiPort(t *testing.T) {
	entry := &types.PortEntry{Port: 3000, PID: 1001, ProcessName: "node", Protocol: "tcp", User: "eray", LocalAddr: "127.0.0.1"}
	entries := []types.PortEntry{
		{Port: 3000, PID: 1001, ProcessName: "node", Protocol: "tcp", User: "eray", LocalAddr: "127.0.0.1"},
		{Port: 3001, PID: 1001, ProcessName: "node", Protocol: "tcp", User: "eray", LocalAddr: "127.0.0.1"},
		{Port: 5432, PID: 4321, ProcessName: "postgres", Protocol: "tcp", User: "postgres", LocalAddr: "127.0.0.1"},
	}

	view := confirmKillView(entry, impactedPortsByPID(entries, entry.PID), 0)
	if !strings.Contains(view, "Affects 2 listening ports") {
		t.Fatalf("expected impact summary in confirm dialog, got: %q", view)
	}
	if !strings.Contains(view, "3000") || !strings.Contains(view, "3001") {
		t.Fatalf("expected affected port list in confirm dialog, got: %q", view)
	}
}

func TestConfirmKillViewShowsSinglePortImpact(t *testing.T) {
	entry := &types.PortEntry{Port: 8080, PID: 1234, ProcessName: "api-server", Protocol: "tcp", User: "eray", LocalAddr: "127.0.0.1"}
	entries := []types.PortEntry{
		{Port: 8080, PID: 1234, ProcessName: "api-server", Protocol: "tcp", User: "eray", LocalAddr: "127.0.0.1"},
	}

	view := confirmKillView(entry, impactedPortsByPID(entries, entry.PID), 0)
	if !strings.Contains(view, "Affects 1 listening port") {
		t.Fatalf("expected single-port impact summary, got: %q", view)
	}
}

func TestFormatImpactSummaryZero(t *testing.T) {
	got := formatImpactSummary(nil)
	if got != "Affects 0 listening ports" {
		t.Fatalf("expected zero-case summary, got %q", got)
	}
}

func TestFormatImpactSummaryTruncatesLongPortList(t *testing.T) {
	ports := []uint16{1000, 1001, 1002, 1003, 1004, 1005, 1006}
	got := formatImpactSummary(ports)
	if !strings.Contains(got, "Affects 7 listening ports") {
		t.Fatalf("expected total count in summary, got %q", got)
	}
	if !strings.Contains(got, "(+1 more)") {
		t.Fatalf("expected overflow suffix in summary, got %q", got)
	}
}

func TestConfirmImpactSnapshotStaysStableDuringRefresh(t *testing.T) {
	svc := &mockProcessServiceKill{}
	entries := []types.PortEntry{
		{Port: 3000, Protocol: "tcp", PID: 1001, ProcessName: "node", User: "eray", LocalAddr: "127.0.0.1"},
		{Port: 3001, Protocol: "tcp", PID: 1001, ProcessName: "node", User: "eray", LocalAddr: "127.0.0.1"},
	}
	m := New(mockScanner{}, svc)
	updated, _ := m.Update(portsLoadedMsg{entries: entries})
	m = updated.(Model)

	updated2, _ := m.Update(tea.KeyPressMsg{Text: "x"})
	m = updated2.(Model)

	if got := formatImpactSummary(m.confirmImpactPorts); !strings.Contains(got, "Affects 2 listening ports") {
		t.Fatalf("expected initial snapshot to show 2 ports, got %q", got)
	}

	refreshEntries := []types.PortEntry{
		{Port: 5432, Protocol: "tcp", PID: 4321, ProcessName: "postgres", User: "postgres", LocalAddr: "127.0.0.1"},
	}
	updated3, _ := m.Update(portsLoadedMsg{entries: refreshEntries})
	m = updated3.(Model)

	if got := formatImpactSummary(m.confirmImpactPorts); !strings.Contains(got, "Affects 2 listening ports") {
		t.Fatalf("expected snapshot to remain stable during confirm state, got %q", got)
	}
}

func TestKillDoesNotOpenConfirmForWindowsSystemPID(t *testing.T) {
	origGOOS := currentTuiGOOS
	t.Cleanup(func() { currentTuiGOOS = origGOOS })
	currentTuiGOOS = "windows"

	svc := &mockProcessServiceKill{}
	entries := []types.PortEntry{
		{Port: 80, Protocol: "tcp", PID: 4, ProcessName: "System", User: "NT AUTHORITY\\SYSTEM", LocalAddr: "0.0.0.0"},
	}
	m := New(mockScanner{}, svc)
	updated, _ := m.Update(portsLoadedMsg{entries: entries})
	m = updated.(Model)

	updated2, cmd := m.Update(tea.KeyPressMsg{Text: "x"})
	got := updated2.(Model)

	if cmd == nil {
		t.Fatal("expected status clear cmd for blocked kill")
	}
	if got.state != stateTable {
		t.Fatalf("expected to remain in table state, got %v", got.state)
	}
	if got.selectedEntry != nil {
		t.Fatal("expected selected entry to remain nil when kill is blocked")
	}
	if !got.statusIsErr {
		t.Fatal("expected blocked kill to set error status")
	}
	if !strings.Contains(got.statusMsg, "Kill not available") {
		t.Fatalf("expected blocked kill message, got %q", got.statusMsg)
	}
}
