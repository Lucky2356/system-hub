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