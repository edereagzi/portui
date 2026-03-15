package scanner

import (
	"context"
	"net"
	"reflect"
	"runtime"
	"testing"
	"time"

	gnet "github.com/shirou/gopsutil/v4/net"

	"github.com/edereagzi/portui/internal/types"
)

type mockProcess struct {
	name    string
	nameErr error
	user    string
	userErr error
}

func (m *mockProcess) Name() (string, error) {
	if m.nameErr != nil {
		return "", m.nameErr
	}
	return m.name, nil
}

func (m *mockProcess) Username() (string, error) {
	if m.userErr != nil {
		return "", m.userErr
	}
	return m.user, nil
}

func TestParseLsofOutputWildcardPort(t *testing.T) {
	input := "node\t1234\tuser\t21u\tIPv4\t12345\t0t0\tTCP\t*:3000 (LISTEN)"
	entries := parseLsofOutput([]byte(input))

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Port != 3000 || entries[0].PID != 1234 || entries[0].ProcessName != "node" {
		t.Fatalf("unexpected entry: %+v", entries[0])
	}
}

func TestParseLsofOutputIPv4Localhost(t *testing.T) {
	input := "python3\t5678\tuser\t4u\tIPv4\t23456\t0t0\tTCP\t127.0.0.1:8080 (LISTEN)"
	entries := parseLsofOutput([]byte(input))

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Port != 8080 || entries[0].PID != 5678 || entries[0].LocalAddr != "127.0.0.1:8080" {
		t.Fatalf("unexpected entry: %+v", entries[0])
	}
}

func TestParseLsofOutputIPv6(t *testing.T) {
	input := "nginx\t9999\troot\t6u\tIPv6\t34567\t0t0\tTCP\t[::1]:443 (LISTEN)"
	entries := parseLsofOutput([]byte(input))

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Port != 443 || entries[0].PID != 9999 || entries[0].LocalAddr != "[::1]:443" {
		t.Fatalf("unexpected entry: %+v", entries[0])
	}
}

func TestParseLsofOutputSkipsHeader(t *testing.T) {
	input := "COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME"
	entries := parseLsofOutput([]byte(input))
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries for header row, got %d", len(entries))
	}
}

func TestParseLsofOutputWithHeader(t *testing.T) {
	input := "COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME\n" +
		"node      1234   eray   21u  IPv4 0x1234567890      0t0  TCP  *:3000 (LISTEN)\n" +
		"postgres  4321   pg     5u   IPv4 0x9876543210      0t0  TCP  127.0.0.1:5432 (LISTEN)"
	entries := parseLsofOutput([]byte(input))

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Port != 3000 || entries[0].PID != 1234 || entries[0].ProcessName != "node" || entries[0].User != "eray" {
		t.Fatalf("unexpected first entry: %+v", entries[0])
	}
	if entries[1].Port != 5432 || entries[1].PID != 4321 || entries[1].LocalAddr != "127.0.0.1:5432" {
		t.Fatalf("unexpected second entry: %+v", entries[1])
	}
}

func TestParseLsofOutputExtraColumns(t *testing.T) {
	input := "node\t1234\tuser\t21u\tIPv4\t12345\t0t0\tTCP\textra_field\t*:3000 (LISTEN)"
	entries := parseLsofOutput([]byte(input))

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry even with extra columns, got %d", len(entries))
	}
	if entries[0].Port != 3000 || entries[0].LocalAddr != "*:3000" {
		t.Fatalf("unexpected entry: %+v", entries[0])
	}
}

func TestGopsutilScannerOnlyListen(t *testing.T) {
	origConnections := connectionsWithContext
	origNewProcess := newProcess
	t.Cleanup(func() {
		connectionsWithContext = origConnections
		newProcess = origNewProcess
	})

	connectionsWithContext = func(ctx context.Context, kind string) ([]gnet.ConnectionStat, error) {
		return []gnet.ConnectionStat{
			{Status: "ESTABLISHED", Pid: 2000, Laddr: gnet.Addr{IP: "127.0.0.1", Port: 9999}},
			{Status: "LISTEN", Pid: 1234, Laddr: gnet.Addr{IP: "127.0.0.1", Port: 8080}},
		}, nil
	}
	newProcess = func(pid int32) (processInfo, error) {
		return &mockProcess{name: "svc", user: "eray"}, nil
	}

	s := &GopsutilScanner{}
	entries, err := s.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 LISTEN entry, got %d", len(entries))
	}
	if entries[0].Port != 8080 || entries[0].State != "LISTEN" {
		t.Fatalf("unexpected entry: %+v", entries[0])
	}
}

type fixedScanner struct {
	entries []types.PortEntry
	err     error
}

func (f *fixedScanner) Scan(_ context.Context) ([]types.PortEntry, error) {
	if f.entries == nil {
		return []types.PortEntry{}, f.err
	}
	return f.entries, f.err
}

func TestNewScannerFallsBackToLsofWhenPIDZero(t *testing.T) {
	origGopsutilFactory := newGopsutilScanner
	origLsofFactory := newLsofScanner
	t.Cleanup(func() {
		newGopsutilScanner = origGopsutilFactory
		newLsofScanner = origLsofFactory
	})

	gopsutilCandidate := &fixedScanner{entries: []types.PortEntry{{Port: 3000, PID: 0, State: "LISTEN"}}}
	lsofCandidate := &fixedScanner{entries: []types.PortEntry{{Port: 3000, PID: 1234, State: "LISTEN"}}}

	newGopsutilScanner = func() types.PortScanner { return gopsutilCandidate }
	newLsofScanner = func() types.PortScanner { return lsofCandidate }

	got := NewScanner()
	if runtime.GOOS == "darwin" {
		if !reflect.DeepEqual(got, lsofCandidate) {
			t.Fatalf("expected lsof scanner fallback on darwin when pid=0")
		}
		return
	}

	if !reflect.DeepEqual(got, gopsutilCandidate) {
		t.Fatalf("expected gopsutil scanner on non-darwin, got %#v", got)
	}
}

func TestGopsutilScannerIntegrationFindsRealListener(t *testing.T) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	port := ln.Addr().(*net.TCPAddr).Port

	time.Sleep(100 * time.Millisecond)

	entries, err := (&GopsutilScanner{}).Scan(context.Background())
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	found := false
	for _, entry := range entries {
		if int(entry.Port) == port {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected to find listener on port %d", port)
	}
}

func TestGopsutilScannerReturnsEmptySliceNotNil(t *testing.T) {
	origConnections := connectionsWithContext
	t.Cleanup(func() { connectionsWithContext = origConnections })

	connectionsWithContext = func(ctx context.Context, kind string) ([]gnet.ConnectionStat, error) {
		return []gnet.ConnectionStat{}, nil
	}

	entries, err := (&GopsutilScanner{}).Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries == nil {
		t.Fatalf("expected empty slice, got nil")
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty slice length, got %d", len(entries))
	}
}
