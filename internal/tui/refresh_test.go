package tui

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/edereagzi/portui/internal/types"
)

func TestRefreshIntervalIsConst(t *testing.T) {
	if refreshInterval != 3*time.Second {
		t.Fatalf("expected 3s refresh interval, got %v", refreshInterval)
	}
}

func TestRefreshTickMsgTriggersRescan(t *testing.T) {
	m := newTestModel()
	updated, _ := m.Update(portsLoadedMsg{entries: testEntries()})
	m = updated.(Model)

	_, cmd := m.Update(tickMsg{})
	if cmd == nil {
		t.Fatal("expected tickMsg to return non-nil command (scan + tick)")
	}
}

func TestRefreshRKeyTriggersRescan(t *testing.T) {
	m := newTestModel()
	updated, _ := m.Update(portsLoadedMsg{entries: testEntries()})
	m = updated.(Model)

	_, cmd := m.Update(tea.KeyPressMsg{Text: "r"})
	if cmd == nil {
		t.Fatal("expected 'r' key to return non-nil command (scan + tick)")
	}
}

func TestRefreshSelectedRowPreservedAfterRefresh(t *testing.T) {
	entries := []types.PortEntry{
		{Port: 8080, Protocol: "tcp", PID: 1234, ProcessName: "api-server", User: "eray", LocalAddr: "127.0.0.1"},
		{Port: 5432, Protocol: "tcp", PID: 4321, ProcessName: "postgres", User: "postgres", LocalAddr: "0.0.0.0"},
		{Port: 3000, Protocol: "tcp", PID: 9999, ProcessName: "node", User: "eray", LocalAddr: "127.0.0.1"},
	}

	m := newTestModel()
	updated, _ := m.Update(portsLoadedMsg{entries: entries})
	m = updated.(Model)

	m.table.SetCursor(1)
	prevRow := m.table.SelectedRow()
	if prevRow == nil {
		t.Fatal("expected selected row to be non-nil after SetCursor")
	}

	updated2, _ := m.Update(portsLoadedMsg{entries: entries})
	m2 := updated2.(Model)

	newRow := m2.table.SelectedRow()
	if newRow == nil {
		t.Fatal("expected selected row to be non-nil after refresh")
	}

	if newRow[0] != prevRow[0] || newRow[2] != prevRow[2] {
		t.Fatalf("expected cursor on Port=%s PID=%s, got Port=%s PID=%s",
			prevRow[0], prevRow[2], newRow[0], newRow[2])
	}
}

func TestRefreshNewPortAppearsAfterRefresh(t *testing.T) {
	initial := []types.PortEntry{
		{Port: 8080, Protocol: "tcp", PID: 1234, ProcessName: "api-server", User: "eray", LocalAddr: "127.0.0.1"},
	}
	refreshed := []types.PortEntry{
		{Port: 8080, Protocol: "tcp", PID: 1234, ProcessName: "api-server", User: "eray", LocalAddr: "127.0.0.1"},
		{Port: 9090, Protocol: "tcp", PID: 5678, ProcessName: "new-service", User: "eray", LocalAddr: "127.0.0.1"},
	}

	m := newTestModel()
	m2, _ := m.Update(portsLoadedMsg{entries: initial})
	model := m2.(Model)

	m3, _ := model.Update(portsLoadedMsg{entries: refreshed})
	model2 := m3.(Model)

	rows := model2.table.Rows()
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows after refresh, got %d", len(rows))
	}

	found := false
	for _, row := range rows {
		if row[0] == "9090" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected new port 9090 to appear in table after refresh")
	}
}

func TestRefreshRemovedPortDisappearsAfterRefresh(t *testing.T) {
	initial := []types.PortEntry{
		{Port: 8080, Protocol: "tcp", PID: 1234, ProcessName: "api-server", User: "eray", LocalAddr: "127.0.0.1"},
		{Port: 5432, Protocol: "tcp", PID: 4321, ProcessName: "postgres", User: "postgres", LocalAddr: "0.0.0.0"},
	}
	afterRemoval := []types.PortEntry{
		{Port: 8080, Protocol: "tcp", PID: 1234, ProcessName: "api-server", User: "eray", LocalAddr: "127.0.0.1"},
	}

	m := newTestModel()
	m2, _ := m.Update(portsLoadedMsg{entries: initial})
	model := m2.(Model)

	m3, _ := model.Update(portsLoadedMsg{entries: afterRemoval})
	model2 := m3.(Model)

	rows := model2.table.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row after removal refresh, got %d", len(rows))
	}

	for _, row := range rows {
		if row[0] == "5432" {
			t.Fatal("expected port 5432 to be removed from table after refresh")
		}
	}
}
