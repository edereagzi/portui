package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/edereagzi/portui/internal/types"
)

type portPIDKey struct {
	port uint16
	pid  int32
}

func deduplicateEntries(entries []types.PortEntry) []types.PortEntry {
	seen := make(map[portPIDKey]int, len(entries))
	result := make([]types.PortEntry, 0, len(entries))

	for _, e := range entries {
		key := portPIDKey{port: e.Port, pid: e.PID}
		if idx, exists := seen[key]; exists {
			// Prefer non-wildcard / more specific address
			if isWildcardAddr(result[idx].LocalAddr) && !isWildcardAddr(e.LocalAddr) {
				result[idx] = e
			}
			continue
		}
		seen[key] = len(result)
		result = append(result, e)
	}

	return result
}

func isWildcardAddr(addr string) bool {
	return strings.HasPrefix(addr, "*:") || strings.HasPrefix(addr, "::") || strings.HasPrefix(addr, "0.0.0.0:")
}

func filteredEntries(entries []types.PortEntry, query string) []types.PortEntry {
	deduped := deduplicateEntries(entries)
	var result []types.PortEntry

	if query == "" {
		result = deduped
	} else {
		q := strings.ToLower(query)
		result = make([]types.PortEntry, 0)

		for _, e := range deduped {
			portStr := fmt.Sprintf("%d", e.Port)
			pidStr := fmt.Sprintf("%d", e.PID)
			processLower := strings.ToLower(e.ProcessName)

			if strings.Contains(portStr, q) ||
				strings.Contains(processLower, q) ||
				strings.HasPrefix(pidStr, q) {
				result = append(result, e)
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Port < result[j].Port
	})

	return result
}
