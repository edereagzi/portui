package tui

import (
	"fmt"
	"sort"
	"strconv"
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
	} else if !hasOperatorToken(query) {
		q := strings.ToLower(query)
		result = make([]types.PortEntry, 0)

		for _, e := range deduped {
			if matchesPlainToken(e, q) {
				result = append(result, e)
			}
		}
	} else {
		result = make([]types.PortEntry, 0)
		tokens := strings.Fields(query)

		for _, e := range deduped {
			if matchesAllTokens(e, tokens) {
				result = append(result, e)
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Port < result[j].Port
	})

	return result
}

func hasOperatorToken(query string) bool {
	for _, token := range strings.Fields(query) {
		field, _, ok := strings.Cut(token, ":")
		if !ok {
			continue
		}
		switch strings.ToLower(field) {
		case "port", "proc", "pid", "user":
			return true
		}
	}
	return false
}

func matchesAllTokens(e types.PortEntry, tokens []string) bool {
	for _, token := range tokens {
		if !matchesToken(e, token) {
			return false
		}
	}
	return true
}

func matchesToken(e types.PortEntry, token string) bool {
	field, value, ok := strings.Cut(token, ":")
	if ok {
		field = strings.ToLower(strings.TrimSpace(field))
		value = strings.TrimSpace(value)
		switch field {
		case "port":
			port, err := strconv.Atoi(value)
			if err != nil || port < 0 || port > 65535 {
				return false
			}
			return uint16(port) == e.Port
		case "proc":
			if value == "" {
				return false
			}
			return strings.Contains(strings.ToLower(e.ProcessName), strings.ToLower(value))
		case "pid":
			pid, err := strconv.Atoi(value)
			if err != nil || pid < 0 || pid > 2147483647 {
				return false
			}
			return int32(pid) == e.PID
		case "user":
			if value == "" {
				return false
			}
			return strings.Contains(strings.ToLower(e.User), strings.ToLower(value))
		}
	}

	return matchesPlainToken(e, strings.ToLower(token))
}

func matchesPlainToken(e types.PortEntry, token string) bool {
	portStr := fmt.Sprintf("%d", e.Port)
	pidStr := fmt.Sprintf("%d", e.PID)
	processLower := strings.ToLower(e.ProcessName)

	return strings.Contains(portStr, token) ||
		strings.Contains(processLower, token) ||
		strings.HasPrefix(pidStr, token)
}
