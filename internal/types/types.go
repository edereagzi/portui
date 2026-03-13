package types

import "context"

// PortEntry represents a single listening TCP port and its owner process.
type PortEntry struct {
	Port        uint16
	Protocol    string
	PID         int32
	ProcessName string
	User        string
	State       string
	LocalAddr   string
}

// ProcessInfo holds details about a running process.
type ProcessInfo struct {
	PID        int32
	Name       string
	Cmdline    string
	User       string
	MemoryRSS  uint64
	CreateTime int64
	ParentPID  int32
}

// PortScanner scans the system for listening TCP ports.
type PortScanner interface {
	Scan(ctx context.Context) ([]PortEntry, error)
}

// ProcessService provides process information and management.
type ProcessService interface {
	GetInfo(pid int32) (*ProcessInfo, error)
	Kill(pid int32) error
}
