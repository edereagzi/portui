package process

import (
	"fmt"
	"strings"
	"time"

	gprocess "github.com/shirou/gopsutil/v4/process"

	"github.com/edereagzi/portui/internal/types"
)

const killGracePeriod = 3 * time.Second

type GopsutilProcessService struct{}

func NewProcessService() *GopsutilProcessService {
	return &GopsutilProcessService{}
}

func (s *GopsutilProcessService) GetInfo(pid int32) (*types.ProcessInfo, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("invalid PID %d", pid)
	}

	p, err := gprocess.NewProcess(pid)
	if err != nil {
		return nil, wrapProcessError(pid, err)
	}

	info := &types.ProcessInfo{PID: pid}

	if name, err := p.Name(); err == nil {
		info.Name = name
	}
	if cmdline, err := p.Cmdline(); err == nil {
		info.Cmdline = cmdline
	}
	if username, err := p.Username(); err == nil {
		info.User = username
	}
	if memInfo, err := p.MemoryInfo(); err == nil && memInfo != nil {
		info.MemoryRSS = memInfo.RSS
	}
	if createTime, err := p.CreateTime(); err == nil {
		info.CreateTime = createTime
	}
	if parentPID, err := p.Ppid(); err == nil {
		info.ParentPID = parentPID
	}

	return info, nil
}

func (s *GopsutilProcessService) Kill(pid int32) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID %d", pid)
	}

	p, err := gprocess.NewProcess(pid)
	if err != nil {
		return wrapProcessError(pid, err)
	}

	if err := p.Terminate(); err != nil {
		return wrapProcessError(pid, err)
	}

	deadline := time.After(killGracePeriod)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			if running, _ := p.IsRunning(); running {
				if err := p.Kill(); err != nil {
					return wrapProcessError(pid, err)
				}
			}
			return nil
		case <-ticker.C:
			if running, _ := p.IsRunning(); !running {
				return nil
			}
		}
	}
}

func wrapProcessError(pid int32, err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "already finished") || strings.Contains(msg, "no such process"):
		return fmt.Errorf("process not found (PID %d): %w", pid, err)
	case strings.Contains(msg, "not permitted") || strings.Contains(msg, "permission denied") || strings.Contains(msg, "operation not permitted"):
		return fmt.Errorf("permission denied (PID %d) - try running with sudo: %w", pid, err)
	default:
		return fmt.Errorf("process error (PID %d): %w", pid, err)
	}
}
