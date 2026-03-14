package tui

import (
	"testing"

	"github.com/edereagzi/portui/internal/types"
)

func TestBindHost(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "wildcard star", in: "*:7000", want: "0.0.0.0"},
		{name: "ipv4 local", in: "127.0.0.1:7265", want: "127.0.0.1"},
		{name: "ipv4 any", in: "0.0.0.0:3000", want: "0.0.0.0"},
		{name: "ipv6 any", in: ":::3000", want: "::"},
		{name: "ipv6 loopback", in: "::1:3000", want: "::1"},
		{name: "bracket ipv6", in: "[::1]:3000", want: "::1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bindHost(tt.in)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestBuildTableUsesBindHostWithoutPort(t *testing.T) {
	entries := []types.PortEntry{{
		Port:        7265,
		Protocol:    "tcp",
		PID:         1001,
		ProcessName: "api",
		User:        "eray",
		LocalAddr:   "127.0.0.1:7265",
	}}

	tbl := buildTable(entries, 0, 0)
	rows := tbl.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0][5] != "127.0.0.1" {
		t.Fatalf("expected bind host only, got %q", rows[0][5])
	}
}
