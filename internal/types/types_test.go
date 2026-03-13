package types

import "testing"

func TestPortEntryZeroValue(t *testing.T) {
	pe := PortEntry{}
	if pe.Port != 0 || pe.Protocol != "" || pe.PID != 0 {
		t.Errorf("PortEntry zero value not valid: %+v", pe)
	}
}

func TestPortEntryInitialization(t *testing.T) {
	pe := PortEntry{
		Port:        8080,
		Protocol:    "tcp",
		PID:         1234,
		ProcessName: "myapp",
		User:        "root",
		State:       "LISTEN",
		LocalAddr:   "127.0.0.1:8080",
	}
	if pe.Port != 8080 || pe.Protocol != "tcp" || pe.PID != 1234 {
		t.Errorf("PortEntry initialization failed: %+v", pe)
	}
}

func TestProcessInfoZeroValue(t *testing.T) {
	pi := ProcessInfo{}
	if pi.PID != 0 || pi.Name != "" || pi.MemoryRSS != 0 {
		t.Errorf("ProcessInfo zero value not valid: %+v", pi)
	}
}

func TestProcessInfoInitialization(t *testing.T) {
	pi := ProcessInfo{
		PID:        5678,
		Name:       "testproc",
		Cmdline:    "/usr/bin/testproc --flag",
		User:       "user",
		MemoryRSS:  1024000,
		CreateTime: 1234567890,
		ParentPID:  1,
	}
	if pi.PID != 5678 || pi.Name != "testproc" || pi.MemoryRSS != 1024000 {
		t.Errorf("ProcessInfo initialization failed: %+v", pi)
	}
}
