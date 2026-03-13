package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/edereagzi/portui/internal/types"
)

func searchTestEntries() []types.PortEntry {
	return []types.PortEntry{
		{Port: 3000, Protocol: "tcp", PID: 1001, ProcessName: "node", User: "eray", LocalAddr: "127.0.0.1"},
		{Port: 3001, Protocol: "tcp", PID: 1002, ProcessName: "nodejs", User: "eray", LocalAddr: "127.0.0.1"},
		{Port: 8080, Protocol: "tcp", PID: 1234, ProcessName: "api-server", User: "eray", LocalAddr: "127.0.0.1"},
		{Port: 5432, Protocol: "tcp", PID: 4321, ProcessName: "postgres", User: "postgres", LocalAddr: "0.0.0.0"},
		{Port: 6379, Protocol: "tcp", PID: 9999, ProcessName: "redis-server", User: "redis", LocalAddr: "127.0.0.1"},
		{Port: 80, Protocol: "tcp", PID: 100, ProcessName: "nginx", User: "root", LocalAddr: "0.0.0.0"},
	}
}

func newLoadedModel(entries []types.PortEntry) Model {
	m := New(mockScanner{}, mockProcessService{})
	updated, _ := m.Update(portsLoadedMsg{entries: entries})
	return updated.(Model)
}

func TestSearchFilterByProcessName(t *testing.T) {
	m := newLoadedModel(searchTestEntries())

	results := filteredEntries(m.entries, "node")
	if len(results) != 2 {
		t.Fatalf("expected 2 results for 'node', got %d: %+v", len(results), results)
	}
	for _, e := range results {
		if !strings.Contains(strings.ToLower(e.ProcessName), "node") {
			t.Errorf("unexpected entry in results: %+v", e)
		}
	}
}

func TestSearchFilterByPort(t *testing.T) {
	m := newLoadedModel(searchTestEntries())

	results := filteredEntries(m.entries, "8080")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for '8080', got %d", len(results))
	}
	if results[0].Port != 8080 {
		t.Errorf("expected port 8080, got %d", results[0].Port)
	}
}

func TestSearchFilterByPortPartial(t *testing.T) {
	m := newLoadedModel(searchTestEntries())

	results := filteredEntries(m.entries, "80")
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results for '80' (ports 8080 and 80), got %d", len(results))
	}
	ports := make(map[uint16]bool)
	for _, e := range results {
		ports[e.Port] = true
	}
	if !ports[8080] {
		t.Error("expected port 8080 in results")
	}
	if !ports[80] {
		t.Error("expected port 80 in results")
	}
}

func TestSearchFilterEmpty(t *testing.T) {
	m := newLoadedModel(searchTestEntries())

	results := filteredEntries(m.entries, "zzznomatch")
	if len(results) != 0 {
		t.Fatalf("expected 0 results for 'zzznomatch', got %d", len(results))
	}
}

func TestSearchClearRestoresAll(t *testing.T) {
	entries := searchTestEntries()
	m := newLoadedModel(entries)

	filtered := filteredEntries(m.entries, "node")
	if len(filtered) == len(entries) {
		t.Fatal("filter should have narrowed results")
	}

	all := filteredEntries(m.entries, "")
	if len(all) != len(entries) {
		t.Fatalf("expected %d entries after clearing filter, got %d", len(entries), len(all))
	}
}

func TestSearchStateTransition(t *testing.T) {
	m := newLoadedModel(searchTestEntries())

	if m.state != stateTable {
		t.Fatalf("expected stateTable initially, got %v", m.state)
	}

	updated, _ := m.Update(tea.KeyPressMsg{Text: "/"})
	got := updated.(Model)
	if got.state != stateSearch {
		t.Fatalf("expected stateSearch after '/', got %v", got.state)
	}

	updated2, _ := got.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	got2 := updated2.(Model)
	if got2.state != stateTable {
		t.Fatalf("expected stateTable after Esc, got %v", got2.state)
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	m := newLoadedModel(searchTestEntries())

	results := filteredEntries(m.entries, "NODE")
	if len(results) == 0 {
		t.Fatal("expected results for 'NODE' (case-insensitive), got 0")
	}
	for _, e := range results {
		if !strings.Contains(strings.ToLower(e.ProcessName), "node") {
			t.Errorf("unexpected entry in case-insensitive results: %+v", e)
		}
	}
}

func TestSearchFilterByPIDPrefix(t *testing.T) {
	m := newLoadedModel(searchTestEntries())

	results := filteredEntries(m.entries, "12")
	found := false
	for _, e := range results {
		if e.PID == 1234 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected PID 1234 in results for '12', got: %+v", results)
	}
}

func TestSearchViewShowsFilterIndicator(t *testing.T) {
	m := newLoadedModel(searchTestEntries())

	updated, _ := m.Update(tea.KeyPressMsg{Text: "/"})
	got := updated.(Model)

	got.filterText = "node"
	got.table = buildTable(filteredEntries(got.entries, "node"), got.width, got.height)

	v := got.View()
	if !strings.Contains(v.Content, "Filter:") {
		t.Fatalf("expected 'Filter:' in view when search active, got: %q", v.Content)
	}
}

func TestSearchEscClearsFilterText(t *testing.T) {
	m := newLoadedModel(searchTestEntries())

	updated, _ := m.Update(tea.KeyPressMsg{Text: "/"})
	got := updated.(Model)
	got.filterText = "node"

	updated2, _ := got.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	got2 := updated2.(Model)

	if got2.filterText != "" {
		t.Fatalf("expected filterText to be empty after Esc, got %q", got2.filterText)
	}
	if got2.state != stateTable {
		t.Fatalf("expected stateTable after Esc, got %v", got2.state)
	}
}

func TestSearchEnterAppliesFilterAndReturnsToTable(t *testing.T) {
	m := newLoadedModel(searchTestEntries())

	updated, _ := m.Update(tea.KeyPressMsg{Text: "/"})
	got := updated.(Model)
	got.filterText = "node"
	got.filterInput.SetValue("node")

	updated2, _ := got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got2 := updated2.(Model)

	if got2.state != stateTable {
		t.Fatalf("expected stateTable after Enter, got %v", got2.state)
	}
	if got2.filterText != "node" {
		t.Fatalf("expected filterText to remain 'node' after Enter, got %q", got2.filterText)
	}
}

func TestDeduplicateByPortPID(t *testing.T) {
	entries := []types.PortEntry{
		{Port: 3000, PID: 100, ProcessName: "node", LocalAddr: "*:3000"},
		{Port: 3000, PID: 100, ProcessName: "node", LocalAddr: "127.0.0.1:3000"},
		{Port: 5432, PID: 200, ProcessName: "postgres", LocalAddr: "127.0.0.1:5432"},
	}

	deduped := deduplicateEntries(entries)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 deduplicated entries, got %d", len(deduped))
	}
	if deduped[0].LocalAddr != "127.0.0.1:3000" {
		t.Fatalf("expected specific addr preferred over wildcard, got %q", deduped[0].LocalAddr)
	}
}

func TestDeduplicatePreservesUniqueEntries(t *testing.T) {
	entries := []types.PortEntry{
		{Port: 3000, PID: 100, ProcessName: "node", LocalAddr: "127.0.0.1:3000"},
		{Port: 5432, PID: 200, ProcessName: "postgres", LocalAddr: "127.0.0.1:5432"},
	}

	deduped := deduplicateEntries(entries)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 entries (no duplicates), got %d", len(deduped))
	}
}

func TestIsWildcardAddr(t *testing.T) {
	wildcards := []string{"*:3000", ":::3000", "0.0.0.0:3000"}
	for _, addr := range wildcards {
		if !isWildcardAddr(addr) {
			t.Errorf("expected %q to be wildcard", addr)
		}
	}

	specifics := []string{"127.0.0.1:3000", "192.168.1.1:8080", "10.0.0.1:443"}
	for _, addr := range specifics {
		if isWildcardAddr(addr) {
			t.Errorf("expected %q to NOT be wildcard", addr)
		}
	}
}

func TestFilteredEntriesSortedByPort(t *testing.T) {
	entries := []types.PortEntry{
		{Port: 8080, PID: 1, ProcessName: "a", LocalAddr: "127.0.0.1:8080"},
		{Port: 80, PID: 2, ProcessName: "b", LocalAddr: "127.0.0.1:80"},
		{Port: 3000, PID: 3, ProcessName: "c", LocalAddr: "127.0.0.1:3000"},
	}

	result := filteredEntries(entries, "")
	for i := 1; i < len(result); i++ {
		if result[i].Port < result[i-1].Port {
			t.Fatalf("entries not sorted by port: %d came after %d", result[i].Port, result[i-1].Port)
		}
	}
}
