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

func TestBuildTableColumnOrder(t *testing.T) {
	entries := []types.PortEntry{{
		Port:        8080,
		Protocol:    "tcp",
		PID:         1234,
		ProcessName: "api-server",
		User:        "eray",
		LocalAddr:   "127.0.0.1:8080",
	}}

	tbl := buildTable(entries, 0, 0)
	rows := tbl.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if len(row) != 6 {
		t.Fatalf("expected 6 columns, got %d", len(row))
	}

	if row[0] != "8080" || row[1] != "127.0.0.1" || row[2] != "api-server" || row[3] != "1234" || row[4] != "tcp" || row[5] != "eray" {
		t.Fatalf("unexpected row order/content: %#v", row)
	}
}
