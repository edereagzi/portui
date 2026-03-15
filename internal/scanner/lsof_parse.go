package scanner

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/edereagzi/portui/internal/types"
)

func parseLsofOutput(output []byte) []types.PortEntry {
	lines := bytes.Split(output, []byte("\n"))
	entries := make([]types.PortEntry, 0, len(lines))

	for _, rawLine := range lines {
		entry, ok := parseLsofLine(string(rawLine))
		if !ok {
			continue
		}
		entries = append(entries, entry)
	}

	return entries
}

const listenSuffix = " (LISTEN)"

func parseLsofLine(line string) (types.PortEntry, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || !strings.HasSuffix(trimmed, "(LISTEN)") {
		return types.PortEntry{}, false
	}

	fields := strings.Fields(trimmed)
	if len(fields) < 4 {
		return types.PortEntry{}, false
	}
	if fields[0] == "COMMAND" {
		return types.PortEntry{}, false
	}

	pid64, err := strconv.ParseInt(fields[1], 10, 32)
	if err != nil {
		return types.PortEntry{}, false
	}

	localAddr := trimmed[:len(trimmed)-len(listenSuffix)]
	if lastWS := strings.LastIndexAny(localAddr, " \t"); lastWS != -1 {
		localAddr = localAddr[lastWS+1:]
	}

	separator := strings.LastIndex(localAddr, ":")
	if separator == -1 || separator == len(localAddr)-1 {
		return types.PortEntry{}, false
	}

	port64, err := strconv.ParseUint(localAddr[separator+1:], 10, 16)
	if err != nil {
		return types.PortEntry{}, false
	}

	return types.PortEntry{
		Port:        uint16(port64),
		Protocol:    "tcp",
		PID:         int32(pid64),
		ProcessName: fields[0],
		User:        fields[2],
		State:       "LISTEN",
		LocalAddr:   localAddr,
	}, true
}
