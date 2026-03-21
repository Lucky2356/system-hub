package system

import (
	"fmt"
	"math"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type Stats struct {
	CPUPercent float64

	RAMUsed    uint64
	RAMTotal   uint64
	RAMPercent float64

	DiskUsed    uint64
	DiskTotal   uint64
	DiskPercent float64

	UptimeSeconds uint64
	UptimeKnown   bool

	SystemdAvailable bool
	DockerAvailable  bool

	ServiceCount      int
	ServiceCountKnown bool

	DockerContainerCount int
	DockerRunningCount   int
	DockerCountKnown     bool
}

func GetStats() (Stats, error) {
	var result Stats

	cpuPercents, err := cpu.Percent(0, false)
	if err != nil {
		return result, fmt.Errorf("get cpu percent: %w", err)
	}
	if len(cpuPercents) > 0 {
		result.CPUPercent = round(cpuPercents[0], 1)
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return result, fmt.Errorf("get virtual memory: %w", err)
	}
	result.RAMUsed = vm.Used
	result.RAMTotal = vm.Total
	result.RAMPercent = round(vm.UsedPercent, 1)

	path := getDiskPath()
	du, err := disk.Usage(path)
	if err != nil {
		return result, fmt.Errorf("get disk usage: %w", err)
	}
	result.DiskUsed = du.Used
	result.DiskTotal = du.Total
	result.DiskPercent = round(du.UsedPercent, 1)

	uptime, err := getSystemUptime()
	if err == nil {
		result.UptimeSeconds = uptime
		result.UptimeKnown = true
	}

	result.SystemdAvailable = IsSystemdAvailable()
	result.DockerAvailable = IsDockerAvailable()

	if result.SystemdAvailable {
		serviceCount, err := CountServices()
		if err == nil {
			result.ServiceCount = serviceCount
			result.ServiceCountKnown = true
		}
	}

	if result.DockerAvailable {
		total, running, err := CountDockerContainers()
		if err == nil {
			result.DockerContainerCount = total
			result.DockerRunningCount = running
			result.DockerCountKnown = true
		}
	}

	return result, nil
}

func getSystemUptime() (uint64, error) {
	if runtime.GOOS == "windows" {
		return getWindowsUptime()
	}

	hostInfo, err := host.Info()
	if err != nil {
		return 0, err
	}
	return hostInfo.Uptime, nil
}

func getWindowsUptime() (uint64, error) {
	// PowerShell-вариант надёжнее, чем wmic, потому что wmic часто отсутствует
	cmd := exec.Command(
		"powershell",
		"-NoProfile",
		"-Command",
		`(Get-Date) - (Get-CimInstance Win32_OperatingSystem).LastBootUpTime | Select-Object -ExpandProperty TotalSeconds`,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(output))
		if text == "" {
			return 0, err
		}
		return 0, fmt.Errorf("%s",text)
	}

	text := strings.TrimSpace(string(output))
	if text == "" {
		return 0, fmt.Errorf("empty uptime output")
	}

	var seconds float64
	_, err = fmt.Sscanf(text, "%f", &seconds)
	if err != nil {
		return 0, fmt.Errorf("parse uptime: %w", err)
	}

	if seconds < 0 {
		return 0, fmt.Errorf("invalid uptime")
	}

	return uint64(seconds), nil
}

func getDiskPath() string {
	if runtime.GOOS == "windows" {
		return "C:\\"
	}
	return "/"
}

func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div := float64(unit)
	exp := 0
	for n := float64(bytes) / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	value := float64(bytes) / div
	suffixes := []string{"KB", "MB", "GB", "TB", "PB"}

	if exp >= len(suffixes) {
		return fmt.Sprintf("%d B", bytes)
	}

	return fmt.Sprintf("%.1f %s", value, suffixes[exp])
}

func FormatUptime(seconds uint64) string {
	d := time.Duration(seconds) * time.Second

	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour

	hours := d / time.Hour
	d -= hours * time.Hour

	minutes := d / time.Minute

	if days > 0 {
		return fmt.Sprintf("%d д %d ч %d мин", days, hours, minutes)
	}

	if hours > 0 {
		return fmt.Sprintf("%d ч %d мин", hours, minutes)
	}

	return fmt.Sprintf("%d мин", minutes)
}

func IsSystemdAvailable() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	cmd := exec.Command("systemctl", "--version")
	err := cmd.Run()
	return err == nil
}

func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "version", "--format", "{{.Client.Version}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) != ""
}

func round(value float64, places int) float64 {
	pow := math.Pow(10, float64(places))
	return math.Round(value*pow) / pow
}