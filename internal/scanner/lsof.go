//go:build darwin

package scanner

import (
	"context"
	"os/exec"

	"github.com/edereagzi/portui/internal/types"
)

func init() {
	newLsofScanner = func() types.PortScanner {
		return &LsofScanner{}
	}
}

type LsofScanner struct{}

func (s *LsofScanner) Scan(ctx context.Context) ([]types.PortEntry, error) {
	cmd := exec.CommandContext(ctx, "lsof", "-nP", "-iTCP", "-sTCP:LISTEN")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	entries := parseLsofOutput(output)
	if entries == nil {
		return []types.PortEntry{}, nil
	}
	return entries, nil
}
