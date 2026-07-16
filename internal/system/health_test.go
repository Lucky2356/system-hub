package system

import (
	"strings"
	"testing"
)

func TestBuildProblemsReturnsEmptyWhenHealthy(t *testing.T) {
	stats := Stats{
		RAMPercent:       10,
		DiskPercent:      20,
		SystemdAvailable: true,
		DockerAvailable:  true,
	}

	problems := BuildProblems(stats, nil, nil)

	// "Healthy" must be an empty slice, not a human-readable sentinel: callers
	// treated the sentinel as a real problem and notified the user about it.
	if len(problems) != 0 {
		t.Errorf("expected no problems, got %v", problems)
	}
}

func TestBuildProblemsFlagsThresholdsAndAvailability(t *testing.T) {
	stats := Stats{
		RAMPercent:       DefaultRAMAlertPercent + 1,
		DiskPercent:      DefaultDiskAlertPercent + 1,
		SystemdAvailable: false,
		DockerAvailable:  false,
	}

	problems := BuildProblems(stats, nil, nil)
	if len(problems) != 4 {
		t.Fatalf("expected 4 problems (ram, disk, systemd, docker), got %d: %v", len(problems), problems)
	}

	joined := strings.ToLower(strings.Join(problems, "|"))
	for _, want := range []string{"ram", "disk", "systemd", "docker"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected a problem mentioning %q, got %v", want, problems)
		}
	}
}

func TestBuildProblemsIgnoresThresholdsBelowLimit(t *testing.T) {
	stats := Stats{
		RAMPercent:       DefaultRAMAlertPercent - 0.1,
		DiskPercent:      DefaultDiskAlertPercent - 0.1,
		SystemdAvailable: true,
		DockerAvailable:  true,
	}

	if problems := BuildProblems(stats, nil, nil); len(problems) != 0 {
		t.Errorf("values just under the thresholds must not alert, got %v", problems)
	}
}
