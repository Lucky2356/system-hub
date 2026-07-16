//go:build linux

package system

import "testing"

func TestControlServicePassesSeparator(t *testing.T) {
	tests := []struct {
		action string
		want   []string
	}{
		{"start", []string{"start", "--", "nginx.service"}},
		{"stop", []string{"stop", "--", "nginx.service"}},
		{"restart", []string{"restart", "--", "nginx.service"}},
		{"enable", []string{"enable", "--", "nginx.service"}},
		{"disable", []string{"disable", "--", "nginx.service"}},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			calls := fakeRunner(t, []byte(""), nil)

			if err := ControlService(tt.action, "nginx.service"); err != nil {
				t.Fatalf("ControlService(%q): %v", tt.action, err)
			}
			if len(*calls) != 1 {
				t.Fatalf("expected 1 call, got %d", len(*calls))
			}
			// "--" must reach systemctl so a unit named like a flag cannot be
			// parsed as one.
			if !argvEquals((*calls)[0], "systemctl", tt.want...) {
				t.Errorf("argv = %s %v, want systemctl %v", (*calls)[0].name, (*calls)[0].args, tt.want)
			}
		})
	}
}

func TestControlServiceRejectsFlagLikeNames(t *testing.T) {
	calls := fakeRunner(t, []byte(""), nil)

	for _, name := range []string{"--version", "-H", "", "  ", "foo bar"} {
		if err := ControlService("start", name); err == nil {
			t.Errorf("expected rejection of service name %q", name)
		}
	}
	if len(*calls) != 0 {
		t.Errorf("no command should run for invalid names, got %+v", *calls)
	}
}

func TestControlServiceRejectsUnknownAction(t *testing.T) {
	calls := fakeRunner(t, []byte(""), nil)

	if err := ControlService("mask", "nginx.service"); err == nil {
		t.Fatal("expected unsupported action to be rejected")
	}
	if len(*calls) != 0 {
		t.Errorf("no command should run for an unsupported action, got %+v", *calls)
	}
}

func TestControlServiceInvalidatesCache(t *testing.T) {
	output := "nginx.service loaded active running Web server"
	calls := fakeRunner(t, []byte(output), nil)

	if _, err := ListServices(); err != nil {
		t.Fatalf("list: %v", err)
	}
	if err := ControlService("restart", "nginx.service"); err != nil {
		t.Fatalf("restart: %v", err)
	}
	if _, err := ListServices(); err != nil {
		t.Fatalf("list after restart: %v", err)
	}

	if len(*calls) != 3 {
		t.Fatalf("expected 3 calls (list, restart, list), got %d", len(*calls))
	}
}

func TestListServicesIsCached(t *testing.T) {
	output := "nginx.service loaded active running Web server"
	calls := fakeRunner(t, []byte(output), nil)

	for i := 0; i < 3; i++ {
		if _, err := ListServices(); err != nil {
			t.Fatalf("ListServices: %v", err)
		}
	}
	// The dashboard consults the listing several times per tick; it must fork
	// systemctl only once within the cache TTL.
	if len(*calls) != 1 {
		t.Errorf("expected 1 systemctl call, got %d", len(*calls))
	}
}

func TestGetServiceLogsBuildsArgv(t *testing.T) {
	calls := fakeRunner(t, []byte("log"), nil)

	if _, err := GetServiceLogs("nginx.service", 50); err != nil {
		t.Fatalf("GetServiceLogs: %v", err)
	}
	if !argvEquals((*calls)[0], "journalctl", "-u", "nginx.service", "-n", "50", "--no-pager") {
		t.Errorf("argv = %s %v", (*calls)[0].name, (*calls)[0].args)
	}

	if _, err := GetServiceLogs("-u evil", 10); err == nil {
		t.Error("expected a flag-like service name to be rejected")
	}
}

func TestParseSystemctlListUnits(t *testing.T) {
	output := "" +
		"nginx.service loaded active running A high performance web server\n" +
		"docker.service loaded active running Docker Application Container Engine\n" +
		"  \n" +
		"cron.service loaded inactive dead Regular background program processing daemon\n" +
		"too few fields\n"

	got := parseSystemctlListUnits(output)

	if len(got) != 3 {
		t.Fatalf("expected 3 services, got %d: %+v", len(got), got)
	}

	if got[0].Name != "nginx.service" {
		t.Errorf("Name = %q, want nginx.service", got[0].Name)
	}
	if got[0].ActiveState != "active" || got[0].SubState != "running" {
		t.Errorf("states = %q/%q, want active/running", got[0].ActiveState, got[0].SubState)
	}
	if got[0].Description != "A high performance web server" {
		t.Errorf("Description = %q", got[0].Description)
	}
	if got[2].ActiveState != "inactive" {
		t.Errorf("cron ActiveState = %q, want inactive", got[2].ActiveState)
	}
}
