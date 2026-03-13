package process

import (
	"os"
	"strings"
	"testing"
)

func TestGetInfoCurrentProcess(t *testing.T) {
	svc := NewProcessService()

	info, err := svc.GetInfo(int32(os.Getpid()))
	if err != nil {
		t.Fatalf("GetInfo returned error: %v", err)
	}
	if info == nil {
		t.Fatal("GetInfo returned nil info")
	}
	if info.Name == "" {
		t.Fatal("expected process name to be non-empty")
	}
	if info.MemoryRSS == 0 {
		t.Fatal("expected memory RSS to be non-zero")
	}
}

func TestGetInfoInvalidPID(t *testing.T) {
	svc := NewProcessService()

	info, err := svc.GetInfo(-1)
	if err == nil {
		t.Fatal("expected error for invalid PID")
	}
	if info != nil {
		t.Fatal("expected nil info for invalid PID")
	}
}

func TestGetInfoNonExistentPID(t *testing.T) {
	svc := NewProcessService()

	info, err := svc.GetInfo(99999999)
	if err == nil {
		t.Fatal("expected error for non-existent PID")
	}
	if info != nil {
		t.Fatal("expected nil info for non-existent PID")
	}
}

func TestKillNonExistent(t *testing.T) {
	svc := NewProcessService()

	err := svc.Kill(99999999)
	if err == nil {
		t.Fatal("expected error for non-existent PID")
	}
}

func TestErrorMessagesFriendly(t *testing.T) {
	svc := NewProcessService()

	_, err := svc.GetInfo(-1)
	if err == nil {
		t.Fatal("expected error for invalid PID")
	}

	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "invalid") && !strings.Contains(msg, "not found") && !strings.Contains(msg, "no such process") {
		t.Fatalf("expected user-friendly error message, got %q", err.Error())
	}
}
