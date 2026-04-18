package tui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/edereagzi/portui/internal/types"
)

func TestHelpOverlayTogglesWithQuestionMark(t *testing.T) {
	m := newLoadedModel(testEntries())

	updated, _ := m.Update(tea.KeyPressMsg{Text: "?"})
	got := updated.(Model)
	if !got.showHelp {
		t.Fatalf("expected showHelp=true after '?', got %v", got.showHelp)
	}

	updated2, _ := got.Update(tea.KeyPressMsg{Text: "?"})
	got2 := updated2.(Model)
	if got2.showHelp {
		t.Fatalf("expected showHelp=false after second '?', got %v", got2.showHelp)
	}
}

func TestHelpOverlayDismissesWithEsc(t *testing.T) {
	m := newLoadedModel(testEntries())
	m.showHelp = true

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	got := updated.(Model)
	if got.showHelp {
		t.Fatalf("expected showHelp=false after Esc, got %v", got.showHelp)
	}
}

func TestHelpViewContainsKeybindings(t *testing.T) {
	m := newLoadedModel(testEntries())
	m.showHelp = true
	m.width = 100
	m.height = 30

	v := m.View()
	for _, want := range []string{"j / ↓", "x", "q / Ctrl+C"} {
		if !strings.Contains(v.Content, want) {
			t.Fatalf("expected help view to contain %q, got %q", want, v.Content)
		}
	}
}

func TestTerminalTooSmallView(t *testing.T) {
	m := newLoadedModel(testEntries())
	m.width = 40
	m.height = 10

	v := m.View()
	if !strings.Contains(v.Content, "too small") {
		t.Fatalf("expected too small warning, got %q", v.Content)
	}
}

func TestTerminalJustLargeEnough(t *testing.T) {
	m := newLoadedModel(testEntries())
	m.width = 60
	m.height = 15

	v := m.View()
	if strings.Contains(v.Content, "too small") {
		t.Fatalf("expected normal view at minimum size, got %q", v.Content)
	}
}

func TestScanErrorShowsInStatusBar(t *testing.T) {
	m := newLoadedModel(testEntries())
	staleCount := len(m.entries)
	wantErr := errors.New("permission denied")

	updated, _ := m.Update(errMsg{err: wantErr})
	got := updated.(Model)

	if len(got.entries) != staleCount {
		t.Fatalf("expected stale entries to remain, got %d entries", len(got.entries))
	}
	if !errors.Is(got.err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, got.err)
	}

	v := got.View()
	if !strings.Contains(v.Content, "⚠ Scan failed") {
		t.Fatalf("expected warning in status bar, got %q", v.Content)
	}
	if !strings.Contains(v.Content, "Try running with sudo") && !strings.Contains(v.Content, "Administrator") {
		t.Fatalf("expected actionable privilege hint in scan error, got %q", v.Content)
	}
	if !strings.Contains(v.Content, "2 ports") {
		t.Fatalf("expected stale port count to remain visible, got %q", v.Content)
	}
}

func TestEmptyStateHasTipText(t *testing.T) {
	m := newLoadedModel(nil)

	v := m.View()
	if !strings.Contains(v.Content, "No listening TCP ports found.") {
		t.Fatalf("expected empty state message, got %q", v.Content)
	}
	if !strings.Contains(v.Content, "Tip:") {
		t.Fatalf("expected tip text in empty state, got %q", v.Content)
	}
}

func TestStatusBarShowsPortCount(t *testing.T) {
	m := newLoadedModel(testEntries())

	v := m.View()
	if !strings.Contains(v.Content, "2 ports") {
		t.Fatalf("expected port count in status area, got %q", v.Content)
	}
}

func TestFooterHintsTable(t *testing.T) {
	m := newLoadedModel(testEntries())
	hints := m.footerHints()
	for _, want := range []string{"q: quit", "?: help", "x: kill"} {
		if !strings.Contains(hints, want) {
			t.Fatalf("expected table footer to contain %q, got %q", want, hints)
		}
	}
}

func TestFooterHintsSearch(t *testing.T) {
	m := newLoadedModel(testEntries())
	m.state = stateSearch
	hints := m.footerHints()
	if !strings.Contains(hints, "enter: apply filter") {
		t.Fatalf("expected search footer with 'enter: apply filter', got %q", hints)
	}
}

func TestFooterHintsConfirmKill(t *testing.T) {
	m := newLoadedModel(testEntries())
	m.state = stateConfirmKill
	hints := m.footerHints()
	if !strings.Contains(hints, "y: confirm kill") {
		t.Fatalf("expected confirm kill footer, got %q", hints)
	}
}

func TestFooterHintsHelp(t *testing.T) {
	m := newLoadedModel(testEntries())
	m.showHelp = true
	hints := m.footerHints()
	if !strings.Contains(hints, "close help") {
		t.Fatalf("expected help footer with 'close help', got %q", hints)
	}
}

func TestDetailViewShowsProcessInfo(t *testing.T) {
	m := newLoadedModel(testEntries())
	m.state = stateDetail
	m.width = 80
	m.height = 30
	m.detailInfo = &types.ProcessInfo{
		PID:     4321,
		Name:    "postgres",
		Cmdline: "/usr/bin/postgres",
		User:    "postgres",
	}
	entry := testEntries()[1]
	m.selectedEntry = &entry

	v := m.View()
	for _, want := range []string{"Process Detail", "postgres", "4321"} {
		if !strings.Contains(v.Content, want) {
			t.Fatalf("expected detail view to contain %q, got %q", want, v.Content)
		}
	}
}

func TestErrMsgFromLoadingTransitionsToTable(t *testing.T) {
	m := newTestModel()
	if m.state != stateLoading {
		t.Fatalf("expected stateLoading, got %v", m.state)
	}

	updated, _ := m.Update(errMsg{err: errors.New("network error")})
	got := updated.(Model)

	if got.state != stateTable {
		t.Fatalf("expected stateTable after error from loading, got %v", got.state)
	}
}
