package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/edereagzi/portui/internal/types"
)

type mockProcessServiceDetail struct {
	info     *types.ProcessInfo
	infoErr  error
	infoPIDs []int32
}

func (m *mockProcessServiceDetail) GetInfo(ctx context.Context, pid int32) (*types.ProcessInfo, error) {
	m.infoPIDs = append(m.infoPIDs, pid)
	if m.infoErr != nil {
		return nil, m.infoErr
	}
	if m.info != nil {
		infoCopy := *m.info
		infoCopy.PID = pid
		return &infoCopy, nil
	}
	return &types.ProcessInfo{PID: pid}, nil
}

func (m *mockProcessServiceDetail) Kill(ctx context.Context, pid int32) error {
	return nil
}

func newDetailTestModel(svc types.ProcessService) Model {
	m := New(mockScanner{}, svc)
	updated, _ := m.Update(portsLoadedMsg{entries: testEntries()})
	return updated.(Model)
}

func TestDetailFormatMemoryKB(t *testing.T) {
	if got := formatMemory(512 * 1024); got != "512 KB" {
		t.Fatalf("expected 512 KB, got %q", got)
	}
}

func TestDetailFormatMemoryMB(t *testing.T) {
	if got := formatMemory(1024 * 1024); got != "1.0 MB" {
		t.Fatalf("expected 1.0 MB, got %q", got)
	}
}

func TestDetailFormatMemoryGB(t *testing.T) {
	if got := formatMemory(1024 * 1024 * 1024); got != "1.0 GB" {
		t.Fatalf("expected 1.0 GB, got %q", got)
	}
}

func TestDetailFormatUptimeHours(t *testing.T) {
	createTime := time.Now().Add(-1 * time.Hour).UnixMilli()

	if got := formatUptime(createTime); got != "1h 0m" {
		t.Fatalf("expected 1h 0m, got %q", got)
	}
}

func TestDetailFormatUptimeDays(t *testing.T) {
	createTime := time.Now().Add(-25 * time.Hour).UnixMilli()

	if got := formatUptime(createTime); got != "1d 1h" {
		t.Fatalf("expected 1d 1h, got %q", got)
	}
}

func TestDetailRenderDetailPanel(t *testing.T) {
	info := &types.ProcessInfo{
		PID:        1234,
		Name:       "nginx",
		Cmdline:    "/usr/sbin/nginx -g daemon off;",
		User:       "www-data",
		MemoryRSS:  45 * 1024 * 1024,
		CreateTime: time.Now().Add(-2*time.Hour - 15*time.Minute).UnixMilli(),
	}
	entry := &types.PortEntry{
		Port:      80,
		Protocol:  "TCP",
		LocalAddr: "0.0.0.0:80",
	}

	got := renderDetailPanel(info, entry, 32)
	for _, want := range []string{"Process Detail", "Name:", "nginx", "Memory:", "45.0 MB", "Port:", "80 (TCP)", "Address:", "0.0.0.0:80"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected detail panel to contain %q, got %q", want, got)
		}
	}
}

func TestDetailEnterKeyTransitionsToDetailState(t *testing.T) {
	info := &types.ProcessInfo{Name: "api-server", Cmdline: "/usr/bin/api-server"}
	svc := &mockProcessServiceDetail{info: info}
	m := newDetailTestModel(svc)

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := updated.(Model)

	if got.state != stateDetail {
		t.Fatalf("expected stateDetail, got %v", got.state)
	}
	if got.selectedEntry == nil {
		t.Fatal("expected selectedEntry to be set")
	}

	if cmd == nil {
		t.Fatal("expected async command for GetInfo")
	}

	msg := cmd()
	updated2, _ := got.Update(msg)
	got2 := updated2.(Model)

	if got2.detailInfo == nil {
		t.Fatal("expected detailInfo to be set after processInfoLoadedMsg")
	}
	if len(svc.infoPIDs) != 1 || svc.infoPIDs[0] != 4321 {
		t.Fatalf("expected GetInfo(4321) (first entry sorted by port), got %v", svc.infoPIDs)
	}
}

func TestDetailEnterKeyIgnoresGetInfoError(t *testing.T) {
	svc := &mockProcessServiceDetail{infoErr: errors.New("boom")}
	m := newDetailTestModel(svc)

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := updated.(Model)

	if got.state != stateDetail {
		t.Fatalf("expected stateDetail while loading, got %v", got.state)
	}

	msg := cmd()
	updated2, _ := got.Update(msg)
	got2 := updated2.(Model)

	if got2.state != stateTable {
		t.Fatalf("expected stateTable on GetInfo error, got %v", got2.state)
	}
	if got2.detailInfo != nil {
		t.Fatal("expected detailInfo to remain nil on GetInfo error")
	}
}

func TestDetailEscKeyReturnsToTable(t *testing.T) {
	svc := &mockProcessServiceDetail{info: &types.ProcessInfo{Name: "api-server"}}
	m := newDetailTestModel(svc)
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(Model)

	updated2, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	got := updated2.(Model)

	if got.state != stateTable {
		t.Fatalf("expected stateTable, got %v", got.state)
	}
	if got.detailInfo != nil {
		t.Fatal("expected detailInfo to be cleared")
	}
}

func TestDetailEnterKeyReturnsToTable(t *testing.T) {
	svc := &mockProcessServiceDetail{info: &types.ProcessInfo{Name: "api-server"}}
	m := newDetailTestModel(svc)
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(Model)

	updated2, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := updated2.(Model)

	if got.state != stateTable {
		t.Fatalf("expected stateTable, got %v", got.state)
	}
	if got.detailInfo != nil {
		t.Fatal("expected detailInfo to be cleared")
	}
}
