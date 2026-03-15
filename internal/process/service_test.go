package process

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

type mockManagedProcess struct {
	terminateErr error
	killErr      error
	runningSeq   []bool
	runningErrs  []error
	runningIdx   int

	terminateCalls int
	killCalls      int
}

func (m *mockManagedProcess) Terminate() error {
	m.terminateCalls++
	return m.terminateErr
}

func (m *mockManagedProcess) Kill() error {
	m.killCalls++
	return m.killErr
}

func (m *mockManagedProcess) IsRunning() (bool, error) {
	err := error(nil)
	if m.runningIdx < len(m.runningErrs) {
		err = m.runningErrs[m.runningIdx]
	}

	if len(m.runningSeq) == 0 {
		m.runningIdx++
		return false, err
	}
	if m.runningIdx >= len(m.runningSeq) {
		m.runningIdx++
		return m.runningSeq[len(m.runningSeq)-1], err
	}
	v := m.runningSeq[m.runningIdx]
	m.runningIdx++
	return v, err
}

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

func TestKillUsesHardKillOnWindows(t *testing.T) {
	origNewManagedProcess := newManagedProcess
	origGOOS := currentGOOS
	t.Cleanup(func() {
		newManagedProcess = origNewManagedProcess
		currentGOOS = origGOOS
	})

	proc := &mockManagedProcess{}
	newManagedProcess = func(pid int32) (managedProcess, error) {
		if pid != 4242 {
			return nil, fmt.Errorf("unexpected pid %d", pid)
		}
		return proc, nil
	}
	currentGOOS = "windows"

	svc := NewProcessService()
	if err := svc.Kill(4242); err != nil {
		t.Fatalf("Kill returned error: %v", err)
	}

	if proc.terminateCalls != 0 {
		t.Fatalf("expected no terminate call on windows, got %d", proc.terminateCalls)
	}
	if proc.killCalls != 1 {
		t.Fatalf("expected exactly one kill call on windows, got %d", proc.killCalls)
	}
}

func TestKillBlocksWindowsSystemPID4(t *testing.T) {
	origNewManagedProcess := newManagedProcess
	origGOOS := currentGOOS
	t.Cleanup(func() {
		newManagedProcess = origNewManagedProcess
		currentGOOS = origGOOS
	})

	called := false
	newManagedProcess = func(pid int32) (managedProcess, error) {
		called = true
		return &mockManagedProcess{}, nil
	}
	currentGOOS = "windows"

	svc := NewProcessService()
	err := svc.Kill(4)
	if err == nil {
		t.Fatal("expected error for PID 4 on windows")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "cannot be terminated") {
		t.Fatalf("expected system pid message, got %v", err)
	}
	if called {
		t.Fatal("expected PID 4 guard to fail before process lookup")
	}
}

func TestKillUsesTerminateFirstOnUnix(t *testing.T) {
	origNewManagedProcess := newManagedProcess
	origGOOS := currentGOOS
	origKillGracePeriod := killGracePeriod
	origKillPollInterval := killPollInterval
	t.Cleanup(func() {
		newManagedProcess = origNewManagedProcess
		currentGOOS = origGOOS
		killGracePeriod = origKillGracePeriod
		killPollInterval = origKillPollInterval
	})

	proc := &mockManagedProcess{runningSeq: []bool{false}}
	newManagedProcess = func(pid int32) (managedProcess, error) {
		if pid != 3131 {
			return nil, fmt.Errorf("unexpected pid %d", pid)
		}
		return proc, nil
	}
	currentGOOS = "linux"
	killGracePeriod = 20 * time.Millisecond
	killPollInterval = 1 * time.Millisecond

	svc := NewProcessService()
	if err := svc.Kill(3131); err != nil {
		t.Fatalf("Kill returned error: %v", err)
	}

	if proc.terminateCalls != 1 {
		t.Fatalf("expected terminate call on unix path, got %d", proc.terminateCalls)
	}
	if proc.killCalls != 0 {
		t.Fatalf("expected no force kill when process exits in grace period, got %d", proc.killCalls)
	}
}

func TestKillForceKillsAfterGracePeriodOnUnix(t *testing.T) {
	origNewManagedProcess := newManagedProcess
	origGOOS := currentGOOS
	origKillGracePeriod := killGracePeriod
	origKillPollInterval := killPollInterval
	t.Cleanup(func() {
		newManagedProcess = origNewManagedProcess
		currentGOOS = origGOOS
		killGracePeriod = origKillGracePeriod
		killPollInterval = origKillPollInterval
	})

	proc := &mockManagedProcess{runningSeq: []bool{true, true, true, true}}
	newManagedProcess = func(pid int32) (managedProcess, error) {
		if pid != 5151 {
			return nil, fmt.Errorf("unexpected pid %d", pid)
		}
		return proc, nil
	}
	currentGOOS = "darwin"
	killGracePeriod = 5 * time.Millisecond
	killPollInterval = 1 * time.Millisecond

	svc := NewProcessService()
	if err := svc.Kill(5151); err != nil {
		t.Fatalf("Kill returned error: %v", err)
	}

	if proc.terminateCalls != 1 {
		t.Fatalf("expected terminate call on unix path, got %d", proc.terminateCalls)
	}
	if proc.killCalls != 1 {
		t.Fatalf("expected force kill after grace period, got %d", proc.killCalls)
	}
}

func TestKillReturnsErrorWhenIsRunningFails(t *testing.T) {
	origNewManagedProcess := newManagedProcess
	origGOOS := currentGOOS
	origKillGracePeriod := killGracePeriod
	origKillPollInterval := killPollInterval
	t.Cleanup(func() {
		newManagedProcess = origNewManagedProcess
		currentGOOS = origGOOS
		killGracePeriod = origKillGracePeriod
		killPollInterval = origKillPollInterval
	})

	proc := &mockManagedProcess{
		runningSeq:  []bool{true},
		runningErrs: []error{fmt.Errorf("permission denied")},
	}
	newManagedProcess = func(pid int32) (managedProcess, error) {
		if pid != 6262 {
			return nil, fmt.Errorf("unexpected pid %d", pid)
		}
		return proc, nil
	}
	currentGOOS = "linux"
	killGracePeriod = 50 * time.Millisecond
	killPollInterval = 1 * time.Millisecond

	svc := NewProcessService()
	err := svc.Kill(6262)
	if err == nil {
		t.Fatal("expected error when IsRunning fails")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "permission denied") {
		t.Fatalf("expected wrapped permission error, got %v", err)
	}
}

func TestKillTreatsGoneAfterDeadlineAsSuccess(t *testing.T) {
	origNewManagedProcess := newManagedProcess
	origGOOS := currentGOOS
	origKillGracePeriod := killGracePeriod
	origKillPollInterval := killPollInterval
	t.Cleanup(func() {
		newManagedProcess = origNewManagedProcess
		currentGOOS = origGOOS
		killGracePeriod = origKillGracePeriod
		killPollInterval = origKillPollInterval
	})

	proc := &mockManagedProcess{
		killErr:    fmt.Errorf("no such process"),
		runningSeq: []bool{true, true, true, true},
	}
	newManagedProcess = func(pid int32) (managedProcess, error) {
		if pid != 7272 {
			return nil, fmt.Errorf("unexpected pid %d", pid)
		}
		return proc, nil
	}
	currentGOOS = "darwin"
	killGracePeriod = 5 * time.Millisecond
	killPollInterval = 1 * time.Millisecond

	svc := NewProcessService()
	if err := svc.Kill(7272); err != nil {
		t.Fatalf("expected success when process vanishes before final kill, got %v", err)
	}
	if proc.killCalls != 1 {
		t.Fatalf("expected final kill attempt once, got %d", proc.killCalls)
	}
}

func TestIsPermissionDeniedMessageWindowsPhrase(t *testing.T) {
	if !isPermissionDeniedMessage("access is denied") {
		t.Fatal("expected windows permission phrase to be recognized")
	}
}
