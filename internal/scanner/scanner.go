package scanner

import (
	"context"
	"strconv"

	gnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"

	"github.com/edereagzi/portui/internal/types"
)

type processInfo interface {
	Name() (string, error)
	Username() (string, error)
}

var connectionsWithContext = gnet.ConnectionsWithContext

var newProcess = func(pid int32) (processInfo, error) {
	return process.NewProcess(pid)
}

type GopsutilScanner struct{}

func (s *GopsutilScanner) Scan(ctx context.Context) ([]types.PortEntry, error) {
	conns, err := connectionsWithContext(ctx, "tcp")
	if err != nil {
		return nil, err
	}

	entries := make([]types.PortEntry, 0, len(conns))
	for _, conn := range conns {
		if conn.Status != "LISTEN" {
			continue
		}
		if conn.Laddr.Port > 65535 {
			continue
		}

		entry := types.PortEntry{
			Port:      uint16(conn.Laddr.Port),
			Protocol:  "tcp",
			PID:       int32(conn.Pid),
			State:     conn.Status,
			LocalAddr: localAddr(conn.Laddr),
		}

		if entry.PID > 0 {
			p, procErr := newProcess(entry.PID)
			if procErr == nil {
				if processName, nameErr := p.Name(); nameErr == nil {
					entry.ProcessName = processName
				}
				if username, userErr := p.Username(); userErr == nil {
					entry.User = username
				}
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func localAddr(addr gnet.Addr) string {
	ip := addr.IP
	if ip == "" {
		ip = "*"
	}
	return ip + ":" + strconv.FormatUint(uint64(addr.Port), 10)
}
