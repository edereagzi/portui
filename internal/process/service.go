package process

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	gprocess "github.com/shirou/gopsutil/v4/process"

	"github.com/edereagzi/portui/internal/types"
)

var killGracePeriod = 3 * time.Second
var killPollInterval = 200 * time.Millisecond
var currentGOOS = runtime.GOOS

type managedProcess interface {
	TerminateWithContext(ctx context.Context) error
	KillWithContext(ctx context.Context) error
	IsRunningWithContext(ctx context.Context) (bool, error)
}

var newManagedProcess = func(pid int32) (managedProcess, error) {
	return gprocess.NewProcess(pid)
}

type GopsutilProcessService struct{}

func NewProcessService() *GopsutilProcessService {
	return &GopsutilProcessService{}
}

func (s *GopsutilProcessService) GetInfo(ctx context.Context, pid int32) (*types.ProcessInfo, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("invalid PID %d", pid)
	}

	p, err := gprocess.NewProcessWithContext(ctx, pid)
	if err != nil {
		return nil, wrapProcessError(pid, err)
	}

	info := &types.ProcessInfo{PID: pid}

	if name, err := p.NameWithContext(ctx); err == nil {
		info.Name = name
	}
	if cmdline, err := p.CmdlineWithContext(ctx); err == nil {
		info.Cmdline = cmdline
	}
	if username, err := p.UsernameWithContext(ctx); err == nil {
		info.User = username
	}
	if memInfo, err := p.MemoryInfoWithContext(ctx); err == nil && memInfo != nil {
		info.MemoryRSS = memInfo.RSS
	}
	if createTime, err := p.CreateTimeWithContext(ctx); err == nil {
		info.CreateTime = createTime
	}
	if parentPID, err := p.PpidWithContext(ctx); err == nil {
		info.ParentPID = parentPID
	}

	return info, nil
}

func (s *GopsutilProcessService) Kill(ctx context.Context, pid int32) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID %d", pid)
	}
	if currentGOOS == "windows" && pid == 4 {
		return fmt.Errorf("system-owned listener (PID 4) cannot be terminated on windows")
	}

	p, err := newManagedProcess(pid)
	if err != nil {
		return wrapProcessError(pid, err)
	}

	if currentGOOS == "windows" {
		if err := p.KillWithContext(ctx); err != nil {
			return wrapProcessError(pid, err)
		}
		return nil
	}

	if err := p.TerminateWithContext(ctx); err != nil {
		return wrapProcessError(pid, err)
	}

	deadline := time.After(killGracePeriod)
	ticker := time.NewTicker(killPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			running, runningErr := p.IsRunningWithContext(ctx)
			if runningErr != nil {
				if isProcessGoneError(runningErr) {
					return nil
				}
				return wrapProcessError(pid, runningErr)
			}
			if running {
				if err := p.KillWithContext(ctx); err != nil {
					if isProcessGoneError(err) {
						return nil
					}
					return wrapProcessError(pid, err)
				}
			}
			return nil
		case <-ticker.C:
			running, runningErr := p.IsRunningWithContext(ctx)
			if runningErr != nil {
				if isProcessGoneError(runningErr) {
					return nil
				}
				return wrapProcessError(pid, runningErr)
			}
			if !running {
				return nil
			}
		}
	}
}

func wrapProcessError(pid int32, err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case isProcessGoneMessage(msg):
		return fmt.Errorf("process not found (PID %d): %w", pid, err)
	case isPermissionDeniedMessage(msg):
		return fmt.Errorf("permission denied (PID %d): %w", pid, err)
	default:
		return fmt.Errorf("process error (PID %d): %w", pid, err)
	}
}

func isProcessGoneError(err error) bool {
	return isProcessGoneMessage(strings.ToLower(err.Error()))
}

func isProcessGoneMessage(msg string) bool {
	return strings.Contains(msg, "already finished") || strings.Contains(msg, "no such process")
}

func isPermissionDeniedMessage(msg string) bool {
	return strings.Contains(msg, "not permitted") ||
		strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "operation not permitted") ||
		strings.Contains(msg, "access is denied")
}
