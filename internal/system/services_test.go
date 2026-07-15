package system

import "testing"

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
