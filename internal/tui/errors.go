package tui

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/edereagzi/portui/internal/types"
)

var currentErrorsGOOS = runtime.GOOS

func elevatedPrivilegesHint() string {
	if currentErrorsGOOS == "windows" {
		return "Try running as Administrator."
	}
	return "Try running with sudo."
}

func isPermissionDeniedMessage(msg string) bool {
	return strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "operation not permitted") ||
		strings.Contains(msg, "not permitted") ||
		strings.Contains(msg, "access is denied")
}

func formatScanStatus(err error, staleCount int) string {
	if err == nil {
		return fmt.Sprintf("%d ports", staleCount)
	}

	msg := strings.ToLower(err.Error())
	switch {
	case isPermissionDeniedMessage(msg):
		return fmt.Sprintf("⚠ Scan failed (permission denied) — %d ports (stale). %s", staleCount, elevatedPrivilegesHint())
	case strings.Contains(msg, "deadline exceeded") || strings.Contains(msg, "context deadline exceeded") || strings.Contains(msg, "timeout"):
		return fmt.Sprintf("⚠ Scan timed out — %d ports (stale). Press r to retry.", staleCount)
	case strings.Contains(msg, "executable file not found") && strings.Contains(msg, "lsof"):
		return fmt.Sprintf("⚠ Scan failed (lsof not found) — %d ports (stale). Install lsof and retry.", staleCount)
	default:
		return fmt.Sprintf("⚠ Scan failed — %d ports (stale). Press r to retry.", staleCount)
	}
}

func formatKillFailure(err error) string {
	msg := strings.ToLower(err.Error())
	switch {
	case isPermissionDeniedMessage(msg):
		return fmt.Sprintf("✗ Kill failed: %s. %s", err, elevatedPrivilegesHint())
	case strings.Contains(msg, "no such process") || strings.Contains(msg, "process not found") || strings.Contains(msg, "already finished"):
		return fmt.Sprintf("✗ Kill failed: %s. Process may already be gone.", err)
	default:
		return fmt.Sprintf("✗ Kill failed: %s. Try again with r then x.", err)
	}
}

func formatProcessInfoFailure(err error) string {
	msg := strings.ToLower(err.Error())
	switch {
	case isPermissionDeniedMessage(msg):
		return fmt.Sprintf("✗ Failed to read process info: %s. %s", err, elevatedPrivilegesHint())
	case strings.Contains(msg, "no such process") || strings.Contains(msg, "process not found") || strings.Contains(msg, "already finished"):
		return fmt.Sprintf("✗ Failed to read process info: %s. Process may already be gone.", err)
	default:
		return fmt.Sprintf("✗ Failed to read process info: %s.", err)
	}
}

func formatUnkillableEntryInfo(entry *types.PortEntry) string {
	if entry == nil {
		return "✗ Process details are not available for this entry."
	}
	if currentErrorsGOOS == "windows" && (entry.PID == 0 || entry.PID == 4) {
		return fmt.Sprintf("✗ Process details unavailable on Windows for PID %d listener.", entry.PID)
	}
	if entry.PID <= 0 {
		return fmt.Sprintf("✗ Process details unavailable for PID %d.", entry.PID)
	}
	return "✗ Process details are not available for this entry."
}
