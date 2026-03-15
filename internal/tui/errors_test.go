package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/edereagzi/portui/internal/types"
)

func TestFormatScanStatusPermission(t *testing.T) {
	origGOOS := currentErrorsGOOS
	currentErrorsGOOS = "darwin"
	t.Cleanup(func() { currentErrorsGOOS = origGOOS })

	msg := formatScanStatus(errors.New("permission denied"), 2)
	if !strings.Contains(msg, "Try running with sudo") {
		t.Fatalf("expected actionable sudo hint, got %q", msg)
	}
}

func TestFormatScanStatusPermissionWindows(t *testing.T) {
	origGOOS := currentErrorsGOOS
	currentErrorsGOOS = "windows"
	t.Cleanup(func() { currentErrorsGOOS = origGOOS })

	msg := formatScanStatus(errors.New("access is denied"), 2)
	if !strings.Contains(msg, "Administrator") {
		t.Fatalf("expected Windows admin hint, got %q", msg)
	}
}

func TestFormatScanStatusTimeout(t *testing.T) {
	msg := formatScanStatus(errors.New("context deadline exceeded"), 3)
	if !strings.Contains(msg, "Press r to retry") {
		t.Fatalf("expected retry hint, got %q", msg)
	}
}

func TestFormatScanStatusLsofMissing(t *testing.T) {
	msg := formatScanStatus(errors.New("exec: \"lsof\": executable file not found in $PATH"), 1)
	if !strings.Contains(msg, "Install lsof and retry") {
		t.Fatalf("expected lsof installation hint, got %q", msg)
	}
}

func TestFormatKillFailurePermission(t *testing.T) {
	origGOOS := currentErrorsGOOS
	currentErrorsGOOS = "darwin"
	t.Cleanup(func() { currentErrorsGOOS = origGOOS })

	msg := formatKillFailure(errors.New("permission denied (PID 1)"))
	if !strings.Contains(msg, "Try running with sudo") {
		t.Fatalf("expected sudo hint for kill failure, got %q", msg)
	}
}

func TestFormatKillFailurePermissionWindows(t *testing.T) {
	origGOOS := currentErrorsGOOS
	currentErrorsGOOS = "windows"
	t.Cleanup(func() { currentErrorsGOOS = origGOOS })

	msg := formatKillFailure(errors.New("access is denied"))
	if !strings.Contains(msg, "Administrator") {
		t.Fatalf("expected Windows admin hint for kill failure, got %q", msg)
	}
}

func TestFormatKillFailureNotFound(t *testing.T) {
	msg := formatKillFailure(errors.New("process not found (PID 9999)"))
	if !strings.Contains(msg, "already be gone") {
		t.Fatalf("expected process missing hint, got %q", msg)
	}
}

func TestFormatProcessInfoFailurePermission(t *testing.T) {
	origGOOS := currentErrorsGOOS
	currentErrorsGOOS = "darwin"
	t.Cleanup(func() { currentErrorsGOOS = origGOOS })

	msg := formatProcessInfoFailure(errors.New("operation not permitted"))
	if !strings.Contains(msg, "Try running with sudo") {
		t.Fatalf("expected sudo hint for process info failure, got %q", msg)
	}
}

func TestFormatUnkillableEntryInfoWindowsPID4(t *testing.T) {
	origGOOS := currentErrorsGOOS
	currentErrorsGOOS = "windows"
	t.Cleanup(func() { currentErrorsGOOS = origGOOS })

	msg := formatUnkillableEntryInfo(&types.PortEntry{PID: 4})
	if !strings.Contains(msg, "Windows") {
		t.Fatalf("expected Windows-specific unkillable hint, got %q", msg)
	}
}
