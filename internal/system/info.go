package system

import (
	"fmt"
	"math"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type NetStats struct {
	BytesSent uint64
	BytesRecv uint64
}

type CPUCoreStat struct {
	Core    int
	Percent float64
}

type Stats struct {
	CPUPercent float64
	PerCPU     []CPUCoreStat

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

	NetSent uint64
	NetRecv uint64

	Temperatures []SensorInfo
}

// GetStats returns a full snapshot including systemd/docker counts. Suitable
// for one-off consumers (reports, diagnostics bundle).
func GetStats() (Stats, error) {
	return collectStats(true)
}

// GetStatsLight returns only the cheap metrics (CPU/RAM/disk/network/uptime/
// temperature) plus cached availability flags. It performs no `systemctl
// list-units` / `docker ps` calls, so it is safe to poll at a high frequency.
func GetStatsLight() (Stats, error) {
	return collectStats(false)
}

func collectStats(heavy bool) (Stats, error) {
	var result Stats

	cpuPercents, err := cpu.Percent(0, false)
	if err != nil {
		return result, fmt.Errorf("get cpu percent: %w", err)
	}
	if len(cpuPercents) > 0 {
		result.CPUPercent = round(cpuPercents[0], 1)
	}

	perCPU, err := cpu.Percent(0, true)
	if err == nil {
		result.PerCPU = make([]CPUCoreStat, len(perCPU))
		for i, p := range perCPU {
			result.PerCPU[i] = CPUCoreStat{Core: i, Percent: round(p, 1)}
		}
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

	netIO, err := net.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		result.NetSent = netIO[0].BytesSent
		result.NetRecv = netIO[0].BytesRecv
	}

	result.SystemdAvailable = IsSystemdAvailable()
	result.DockerAvailable = IsDockerAvailable()

	if !heavy {
		return result, nil
	}

	result.Temperatures = GetTemperatures()

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
	output, err := runCmd("powershell",
		"-NoProfile",
		"-Command",
		`(Get-Date) - (Get-CimInstance Win32_OperatingSystem).LastBootUpTime | Select-Object -ExpandProperty TotalSeconds`,
	)
	if err != nil {
		return 0, err
	}

	text := strings.TrimSpace(string(output))
	if text == "" {
		return 0, fmt.Errorf("empty uptime output")
	}

	seconds, err := strconv.ParseFloat(text, 64)
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
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}

	return fmt.Sprintf("%dm", minutes)
}

// Availability of systemd/docker rarely changes during a session, and each
// check spawns a subprocess. Cache results with a short TTL so high-frequency
// dashboard polling does not repeatedly fork `systemctl`/`docker`.
const availabilityTTL = 15 * time.Second

type availabilityCache struct {
	mu       sync.Mutex
	value    bool
	checkedAt time.Time
	valid    bool
}

func (c *availabilityCache) get(check func() bool) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.valid && time.Since(c.checkedAt) < availabilityTTL {
		return c.value
	}
	c.value = check()
	c.checkedAt = time.Now()
	c.valid = true
	return c.value
}

var (
	systemdAvailCache availabilityCache
	dockerAvailCache  availabilityCache
)

func IsSystemdAvailable() bool {
	return systemdAvailCache.get(func() bool {
		if runtime.GOOS != "linux" {
			return false
		}
		_, err := runCmd("systemctl", "--version")
		return err == nil
	})
}

func IsDockerAvailable() bool {
	return dockerAvailCache.get(func() bool {
		output, err := runCmd("docker", "version", "--format", "{{.Client.Version}}")
		if err != nil {
			return false
		}
		return strings.TrimSpace(string(output)) != ""
	})
}

func round(value float64, places int) float64 {
	pow := math.Pow(10, float64(places))
	return math.Round(value*pow) / pow
}