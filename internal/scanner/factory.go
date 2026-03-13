package scanner

import (
	"context"
	"runtime"
	"time"

	"github.com/edereagzi/portui/internal/types"
)

var newGopsutilScanner = func() types.PortScanner {
	return &GopsutilScanner{}
}

var newLsofScanner = func() types.PortScanner {
	return nil
}

func NewScanner() types.PortScanner {
	gopsutilScanner := newGopsutilScanner()
	if runtime.GOOS != "darwin" {
		return gopsutilScanner
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	entries, err := gopsutilScanner.Scan(ctx)
	if err != nil {
		return gopsutilScanner
	}

	lsofScanner := newLsofScanner()
	if lsofScanner == nil {
		return gopsutilScanner
	}

	if allPIDZero(entries) {
		return lsofScanner
	}

	if len(entries) == 0 {
		lsofEntries, lsofErr := lsofScanner.Scan(ctx)
		if lsofErr == nil && len(lsofEntries) > 0 {
			return lsofScanner
		}
	}

	return gopsutilScanner
}

func allPIDZero(entries []types.PortEntry) bool {
	if len(entries) == 0 {
		return false
	}
	for _, entry := range entries {
		if entry.PID != 0 {
			return false
		}
	}
	return true
}
