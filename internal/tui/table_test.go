package tui

import (
	"testing"

	"github.com/edereagzi/portui/internal/types"
)

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

	if row[0] != "8080" || row[1] != "127.0.0.1:8080" || row[2] != "api-server" || row[3] != "1234" || row[4] != "tcp" || row[5] != "eray" {
		t.Fatalf("unexpected row order/content: %#v", row)
	}
}
