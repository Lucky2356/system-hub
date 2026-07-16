package system

import (
	"errors"
	"testing"
)

func TestControlDockerContainerPassesSeparator(t *testing.T) {
	calls := fakeRunner(t, []byte(""), nil)

	if err := ControlDockerContainer("start", "web"); err != nil {
		t.Fatalf("ControlDockerContainer: %v", err)
	}

	if len(*calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(*calls))
	}
	// "--" must reach docker, otherwise a container named like a flag would be
	// parsed as one.
	if !argvEquals((*calls)[0], "docker", "start", "--", "web") {
		t.Errorf("argv = %s %v, want docker start -- web", (*calls)[0].name, (*calls)[0].args)
	}
}

func TestControlDockerContainerRejectsFlagLikeNames(t *testing.T) {
	calls := fakeRunner(t, []byte(""), nil)

	err := ControlDockerContainer("stop", "-H tcp://evil:2375")
	if err == nil {
		t.Fatal("expected rejection of a flag-like container name")
	}
	if len(*calls) != 0 {
		t.Errorf("no command should run for an invalid name, got %v", *calls)
	}
}

func TestControlDockerContainerRejectsUnknownAction(t *testing.T) {
	calls := fakeRunner(t, []byte(""), nil)

	if err := ControlDockerContainer("rm", "web"); err == nil {
		t.Fatal("expected unsupported action to be rejected")
	}
	if len(*calls) != 0 {
		t.Errorf("no command should run for an invalid action, got %v", *calls)
	}
}

func TestPullDockerImageRunsWithoutTimeout(t *testing.T) {
	calls := fakeRunner(t, []byte(""), nil)

	if err := PullDockerImage("nginx:1.25"); err != nil {
		t.Fatalf("PullDockerImage: %v", err)
	}

	if len(*calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(*calls))
	}
	// A large image pull must not be killed by the default 30s command timeout.
	if (*calls)[0].timeout != 0 {
		t.Errorf("timeout = %v, want 0 (no deadline) for docker pull", (*calls)[0].timeout)
	}
	if !argvEquals((*calls)[0], "docker", "pull", "--", "nginx:1.25") {
		t.Errorf("argv = %s %v", (*calls)[0].name, (*calls)[0].args)
	}
}

func TestGetDockerContainerLogsBuildsArgv(t *testing.T) {
	calls := fakeRunner(t, []byte("log line"), nil)

	if _, err := GetDockerContainerLogs("web", 50); err != nil {
		t.Fatalf("GetDockerContainerLogs: %v", err)
	}
	if !argvEquals((*calls)[0], "docker", "logs", "--tail", "50", "--", "web") {
		t.Errorf("argv = %s %v", (*calls)[0].name, (*calls)[0].args)
	}
}

func TestListDockerContainersParsesAndCaches(t *testing.T) {
	output := `{"ID":"abc","Image":"nginx","Names":"web","State":"running","Status":"Up 2 hours"}
{"ID":"def","Image":"redis","Names":"cache","State":"exited","Status":"Exited (0)"}`

	calls := fakeRunner(t, []byte(output), nil)

	got, err := ListDockerContainers()
	if err != nil {
		t.Fatalf("ListDockerContainers: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(got))
	}
	if got[0].Names != "web" || got[0].State != "running" || got[0].Image != "nginx" {
		t.Errorf("unexpected first container: %+v", got[0])
	}

	// A second call within the TTL must be served from cache, so the dashboard
	// does not fork `docker ps` several times per tick.
	if _, err := ListDockerContainers(); err != nil {
		t.Fatalf("second ListDockerContainers: %v", err)
	}
	if len(*calls) != 1 {
		t.Errorf("expected listing to be cached, got %d calls", len(*calls))
	}
}

func TestControlDockerContainerInvalidatesCache(t *testing.T) {
	output := `{"ID":"abc","Image":"nginx","Names":"web","State":"running","Status":"Up"}`
	calls := fakeRunner(t, []byte(output), nil)

	if _, err := ListDockerContainers(); err != nil {
		t.Fatalf("list: %v", err)
	}
	if err := ControlDockerContainer("restart", "web"); err != nil {
		t.Fatalf("restart: %v", err)
	}
	if _, err := ListDockerContainers(); err != nil {
		t.Fatalf("list after restart: %v", err)
	}

	// ps, restart, ps — the listing must be refetched after a mutating action.
	if len(*calls) != 3 {
		t.Fatalf("expected 3 calls (ps, restart, ps), got %d: %+v", len(*calls), *calls)
	}
}

func TestListDockerContainersDoesNotCacheErrors(t *testing.T) {
	calls := fakeRunner(t, nil, errors.New("docker daemon not running"))

	if _, err := ListDockerContainers(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := ListDockerContainers(); err == nil {
		t.Fatal("expected error on retry")
	}
	// A transient failure must be retried, not remembered.
	if len(*calls) != 2 {
		t.Errorf("expected 2 attempts, got %d", len(*calls))
	}
}

func TestListDockerImagesSkipsUntagged(t *testing.T) {
	output := `{"ID":"1","Repository":"nginx","Tag":"latest","Size":"140MB"}
{"ID":"2","Repository":"<none>","Tag":"<none>","Size":"5MB"}`
	fakeRunner(t, []byte(output), nil)

	images, err := ListDockerImages()
	if err != nil {
		t.Fatalf("ListDockerImages: %v", err)
	}
	if len(images) != 1 || images[0].Repository != "nginx" {
		t.Errorf("expected only the tagged image, got %+v", images)
	}
}

func TestRunCmdContextTimeoutIsForwarded(t *testing.T) {
	calls := fakeRunner(t, []byte("ok"), nil)

	if _, err := runCmd("docker", "ps"); err != nil {
		t.Fatalf("runCmd: %v", err)
	}
	if (*calls)[0].timeout != defaultTimeout {
		t.Errorf("timeout = %v, want %v", (*calls)[0].timeout, defaultTimeout)
	}
}

func TestListDockerContainerStatsBuildsArgv(t *testing.T) {
	statsCache.invalidate()
	calls := fakeRunner(t, []byte(`{"Name":"web","CPUPerc":"0.15%","MemUsage":"12.4MiB / 1.94GiB","MemPerc":"0.62%"}`), nil)

	if _, err := ListDockerContainerStats(); err != nil {
		t.Fatalf("ListDockerContainerStats: %v", err)
	}

	if len(*calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(*calls))
	}
	// --no-stream is load-bearing: without it docker streams forever and the
	// call never returns.
	if !argvEquals((*calls)[0], "docker", "stats", "--no-stream", "--format", "{{json .}}") {
		t.Errorf("argv = %s %v, want docker stats --no-stream --format {{json .}}",
			(*calls)[0].name, (*calls)[0].args)
	}
}

func TestParseDockerStats(t *testing.T) {
	output := `{"Name":"web","CPUPerc":"0.15%","MemUsage":"12.4MiB / 1.94GiB","MemPerc":"0.62%"}
{"Name":"db","CPUPerc":"1.20%","MemUsage":"48MiB / 1.94GiB","MemPerc":"2.41%"}`

	stats, err := parseDockerStats(output)
	if err != nil {
		t.Fatalf("parseDockerStats: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("got %d stats, want 2", len(stats))
	}

	if stats[0].Name != "web" || stats[0].CPUPercent != "0.15%" || stats[0].MemPercent != "0.62%" {
		t.Errorf("first entry = %+v", stats[0])
	}
	if stats[1].MemUsage != "48MiB / 1.94GiB" {
		t.Errorf("MemUsage = %q", stats[1].MemUsage)
	}
}

func TestParseDockerStatsHandlesNoRunningContainers(t *testing.T) {
	// With nothing running docker prints only a newline; that is not an error,
	// and it must not become a one-element slice of an empty container.
	stats, err := parseDockerStats("\n")
	if err != nil {
		t.Fatalf("parseDockerStats: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("got %d stats, want 0: %+v", len(stats), stats)
	}
}

func TestParseDockerStatsRejectsGarbage(t *testing.T) {
	if _, err := parseDockerStats("not json"); err == nil {
		t.Error("expected an error for non-JSON output")
	}
}

func TestDockerStatsByNameIndexesEntries(t *testing.T) {
	statsCache.invalidate()
	fakeRunner(t, []byte(`{"Name":"web","CPUPerc":"0.15%","MemUsage":"12.4MiB / 1.94GiB","MemPerc":"0.62%"}`), nil)

	byName, err := DockerStatsByName()
	if err != nil {
		t.Fatalf("DockerStatsByName: %v", err)
	}

	got, ok := byName["web"]
	if !ok {
		t.Fatalf("no entry for web, got %v", byName)
	}
	if got.CPUPercent != "0.15%" {
		t.Errorf("CPUPercent = %q, want 0.15%%", got.CPUPercent)
	}
}

func TestInvalidateContainersCacheAlsoDropsStats(t *testing.T) {
	statsCache.invalidate()
	calls := fakeRunner(t, []byte(`{"Name":"web","CPUPerc":"0.15%","MemUsage":"1MiB / 2MiB","MemPerc":"0.5%"}`), nil)

	if _, err := ListDockerContainerStats(); err != nil {
		t.Fatalf("first call: %v", err)
	}
	// Stale stats after a start/stop would keep showing the usage of a
	// container that is no longer running.
	InvalidateContainersCache()
	if _, err := ListDockerContainerStats(); err != nil {
		t.Fatalf("second call: %v", err)
	}

	if len(*calls) != 2 {
		t.Errorf("expected the invalidated cache to refetch, got %d calls", len(*calls))
	}
}
