//go:build windows

package system

import (
	"testing"

	"golang.org/x/sys/windows/svc"
)

func TestMapServiceState(t *testing.T) {
	// Windows states must land on the systemd vocabulary, because the UI (and
	// StatusColor in particular) only understands those words.
	tests := []struct {
		state       svc.State
		wantActive  string
		wantSubstr  string
		description string
	}{
		{svc.Running, stateActive, subRunning, "running service is active"},
		{svc.Stopped, stateInactive, subDead, "stopped service is inactive"},
		{svc.StartPending, stateActivating, "start-pre", "starting service is activating"},
		{svc.StopPending, stateDeactivating, "stop-sigterm", "stopping service is deactivating"},
		{svc.Paused, stateInactive, "paused", "paused service is inactive"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			active, sub := mapServiceState(tt.state)
			if active != tt.wantActive {
				t.Errorf("activeState = %q, want %q", active, tt.wantActive)
			}
			if sub != tt.wantSubstr {
				t.Errorf("subState = %q, want %q", sub, tt.wantSubstr)
			}
		})
	}
}

func TestMapServiceStateColoursResolve(t *testing.T) {
	// Every mapped state must produce a colour the theme knows; an unmapped word
	// would silently render grey.
	for _, state := range []svc.State{svc.Running, svc.Stopped, svc.StartPending, svc.StopPending} {
		active, _ := mapServiceState(state)
		switch active {
		case stateActive, stateInactive, stateActivating, stateDeactivating, stateFailed:
		default:
			t.Errorf("state %v mapped to unknown activeState %q", state, active)
		}
	}
}

func TestListServicesReturnsRealServices(t *testing.T) {
	// Windows always has an SCM, so listing must work without admin rights.
	services, err := listServices()
	if err != nil {
		t.Fatalf("listServices: %v", err)
	}
	if len(services) == 0 {
		t.Fatal("expected at least one Windows service")
	}

	for _, s := range services {
		if s.Name == "" {
			t.Error("service with an empty name")
		}
		if s.ActiveState == "" {
			t.Errorf("service %q has no ActiveState", s.Name)
		}
	}
}

func TestServiceManagerAvailableOnWindows(t *testing.T) {
	if !serviceManagerAvailable() {
		t.Error("the Windows service manager should always be reachable")
	}
}
