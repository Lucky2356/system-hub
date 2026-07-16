package system

import (
	"fmt"
	"os/user"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/host"
)

type HostInfo struct {
	Hostname        string
	OS              string
	Platform        string
	PlatformVersion string
	Kernel          string
	KernelVersion   string
	Architecture    string
	CurrentUser     string
	Uptime          string
	BootTime        string
	GoVersion       string
}

func GetHostInfo() HostInfo {
	info := HostInfo{
		Hostname:        "N/A",
		OS:              runtime.GOOS,
		Platform:        "N/A",
		PlatformVersion: "N/A",
		Kernel:          "N/A",
		KernelVersion:   "N/A",
		Architecture:    runtime.GOARCH,
		CurrentUser:     "N/A",
		Uptime:          "N/A",
		BootTime:        "N/A",
		GoVersion:       runtime.Version(),
	}

	if hi, err := host.Info(); err == nil {
		if hi.Hostname != "" {
			info.Hostname = hi.Hostname
		}
		if hi.Platform != "" {
			info.Platform = hi.Platform
		}
		if hi.PlatformVersion != "" {
			info.PlatformVersion = hi.PlatformVersion
		}
		if hi.KernelArch != "" {
			info.Architecture = hi.KernelArch
		}
		if hi.KernelVersion != "" {
			info.KernelVersion = hi.KernelVersion
		}
		if hi.OS != "" {
			info.OS = hi.OS
		}
		if hi.KernelVersion != "" {
			info.KernelVersion = hi.KernelVersion
		}
		if hi.HostID != "" && info.Hostname == "N/A" {
			info.Hostname = hi.HostID
		}

		if hi.Uptime > 0 {
			info.Uptime = formatDuration(hi.Uptime)
		}

		if hi.BootTime > 0 {
			bt := time.Unix(int64(hi.BootTime), 0)
			info.BootTime = bt.Format("2006-01-02 15:04:05")
		}

		// Для некоторых платформ поле KernelVersion есть,
		// а отдельного Kernel нет, поэтому подставим разумное значение.
		if info.Kernel == "N/A" {
			switch runtime.GOOS {
			case "windows":
				info.Kernel = "Windows NT"
			case "linux":
				info.Kernel = "Linux"
			case "darwin":
				info.Kernel = "Darwin"
			default:
				info.Kernel = runtime.GOOS
			}
		}
	}

	if u, err := user.Current(); err == nil {
		if u.Username != "" {
			info.CurrentUser = u.Username
		}
	}

	return info
}

func (h HostInfo) ToMultilineString() string {
	return fmt.Sprintf(
		"Hostname: %s\nOS: %s\nPlatform: %s\nPlatform Version: %s\nKernel: %s\nKernel Version: %s\nArchitecture: %s\nCurrent User: %s\nUptime: %s\nBoot Time: %s\nGo Version: %s",
		h.Hostname,
		h.OS,
		h.Platform,
		h.PlatformVersion,
		h.Kernel,
		h.KernelVersion,
		h.Architecture,
		h.CurrentUser,
		h.Uptime,
		h.BootTime,
		h.GoVersion,
	)
}

func formatDuration(seconds uint64) string {
	d := time.Duration(seconds) * time.Second

	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour

	hours := d / time.Hour
	d -= hours * time.Hour

	minutes := d / time.Minute

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
