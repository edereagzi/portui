package tui

import (
	"fmt"
	"strings"
)

func formatScanStatus(err error, staleCount int) string {
	if err == nil {
		return fmt.Sprintf("%d ports", staleCount)
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "permission denied") || strings.Contains(msg, "operation not permitted") || strings.Contains(msg, "not permitted"):
		return fmt.Sprintf("⚠ Scan failed (permission denied) — %d ports (stale). Try running with sudo.", staleCount)
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
	case strings.Contains(msg, "permission denied") || strings.Contains(msg, "operation not permitted") || strings.Contains(msg, "not permitted"):
		return fmt.Sprintf("✗ Kill failed: %s. Try running with sudo.", err)
	case strings.Contains(msg, "no such process") || strings.Contains(msg, "process not found") || strings.Contains(msg, "already finished"):
		return fmt.Sprintf("✗ Kill failed: %s. Process may already be gone.", err)
	default:
		return fmt.Sprintf("✗ Kill failed: %s. Try again with r then x.", err)
	}
}

func formatProcessInfoFailure(err error) string {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "permission denied") || strings.Contains(msg, "operation not permitted") || strings.Contains(msg, "not permitted"):
		return fmt.Sprintf("✗ Failed to read process info: %s. Try running with sudo.", err)
	case strings.Contains(msg, "no such process") || strings.Contains(msg, "process not found") || strings.Contains(msg, "already finished"):
		return fmt.Sprintf("✗ Failed to read process info: %s. Process may already be gone.", err)
	default:
		return fmt.Sprintf("✗ Failed to read process info: %s.", err)
	}
}
