package system

import (
	"fmt"
	"sort"
	"strings"

	gnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

type PortProcessInfo struct {
	Proto       string
	LocalAddr   string
	LocalPort   uint32
	PID         int32
	ProcessName string
	Status      string
}

type ProcessUsageInfo struct {
	PID           int32
	ProcessName   string
	CPUPercent    float64
	MemoryBytes   uint64
	MemoryPercent float32
}

func ListListeningPorts() ([]PortProcessInfo, error) {
	conns, err := gnet.Connections("inet")
	if err != nil {
		return nil, err
	}

	result := make([]PortProcessInfo, 0)
	nameCache := make(map[int32]string)

	for _, conn := range conns {
		status := strings.ToUpper(strings.TrimSpace(conn.Status))

		if status != "LISTEN" && status != "NONE" {
			continue
		}

		proto := "tcp"
		switch conn.Type {
		case 1:
			proto = "tcp"
		case 2:
			proto = "udp"
		}

		processName := "-"
		if conn.Pid > 0 {
			if cached, ok := nameCache[conn.Pid]; ok {
				processName = cached
			} else {
				p, err := process.NewProcess(conn.Pid)
				if err == nil {
					name, nameErr := p.Name()
					if nameErr == nil && strings.TrimSpace(name) != "" {
						processName = name
					}
				}
				nameCache[conn.Pid] = processName
			}
		}

		result = append(result, PortProcessInfo{
			Proto:       proto,
			LocalAddr:   conn.Laddr.IP,
			LocalPort:   conn.Laddr.Port,
			PID:         conn.Pid,
			ProcessName: processName,
			Status:      status,
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].LocalPort == result[j].LocalPort {
			if result[i].Proto == result[j].Proto {
				return result[i].ProcessName < result[j].ProcessName
			}
			return result[i].Proto < result[j].Proto
		}
		return result[i].LocalPort < result[j].LocalPort
	})

	return result, nil
}

func GetPortProcessDetails(item PortProcessInfo) string {
	return fmt.Sprintf(
		"Protocol: %s\nAddress: %s\nPort: %d\nPID: %d\nProcess: %s\nStatus: %s",
		item.Proto,
		item.LocalAddr,
		item.LocalPort,
		item.PID,
		item.ProcessName,
		item.Status,
	)
}

func KillProcess(pid int32) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID: %d", pid)
	}

	if pid == 1 {
		return fmt.Errorf("cannot kill PID 1 (init system)")
	}

	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}

	return p.Kill()
}

func ListTopProcesses(limit int) ([]ProcessUsageInfo, error) {
	if limit <= 0 {
		limit = 5
	}

	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	result := make([]ProcessUsageInfo, 0, len(procs))

	for _, p := range procs {
		name, err := p.Name()
		if err != nil || strings.TrimSpace(name) == "" {
			name = "-"
		}

		cpuPercent, err := p.CPUPercent()
		if err != nil {
			cpuPercent = 0
		}

		memInfo, err := p.MemoryInfo()
		if err != nil || memInfo == nil {
			memInfo = nil
		}

		memPercent, err := p.MemoryPercent()
		if err != nil {
			memPercent = 0
		}

		memBytes := uint64(0)
		if memInfo != nil {
			memBytes = memInfo.RSS
		}

		result = append(result, ProcessUsageInfo{
			PID:           p.Pid,
			ProcessName:   name,
			CPUPercent:    cpuPercent,
			MemoryBytes:   memBytes,
			MemoryPercent: memPercent,
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].CPUPercent == result[j].CPUPercent {
			if result[i].MemoryBytes == result[j].MemoryBytes {
				return result[i].ProcessName < result[j].ProcessName
			}
			return result[i].MemoryBytes > result[j].MemoryBytes
		}
		return result[i].CPUPercent > result[j].CPUPercent
	})

	if len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}
