//go:build windows

package system

import (
	"fmt"
	"sort"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// ServiceManagerName labels the service backend in the UI.
const ServiceManagerName = "Службы Windows"

// The systemd vocabulary the UI speaks. Only the Windows provider needs these
// as constants: the Linux provider takes the words verbatim from systemctl's
// own output.
const (
	stateActive       = "active"
	stateInactive     = "inactive"
	stateFailed       = "failed"
	stateActivating   = "activating"
	stateDeactivating = "deactivating"

	subRunning = "running"
	subDead    = "dead"

	loadLoaded = "loaded"
)

func errUnsupportedAction(action string) error {
	return fmt.Errorf("unsupported action: %s", action)
}

// openSCM opens the service control manager with exactly the access needed.
// mgr.Connect is deliberately not used for reads: it asks for
// SC_MANAGER_ALL_ACCESS, which fails with "Access is denied" unless the app runs
// elevated — so merely listing services would require admin rights.
func openSCM(access uint32) (*mgr.Mgr, error) {
	h, err := windows.OpenSCManager(nil, nil, access)
	if err != nil {
		return nil, fmt.Errorf("connect to service manager: %w", err)
	}
	return &mgr.Mgr{Handle: h}, nil
}

// listServices enumerates services with a single EnumServicesStatusEx call.
// Opening a handle per service would mean hundreds of OpenService calls on
// every refresh; this returns name, display name and state in one syscall.
func listServices() ([]ServiceInfo, error) {
	m, err := openSCM(windows.SC_MANAGER_CONNECT | windows.SC_MANAGER_ENUMERATE_SERVICE)
	if err != nil {
		return nil, err
	}
	defer func() { _ = m.Disconnect() }()

	var bytesNeeded, servicesReturned, resumeHandle uint32

	// The first call only sizes the buffer and is expected to fail with
	// ERROR_MORE_DATA.
	err = windows.EnumServicesStatusEx(
		m.Handle,
		windows.SC_ENUM_PROCESS_INFO,
		windows.SERVICE_WIN32,
		windows.SERVICE_STATE_ALL,
		nil, 0,
		&bytesNeeded, &servicesReturned, &resumeHandle, nil,
	)
	if err != nil && err != windows.ERROR_MORE_DATA {
		return nil, fmt.Errorf("enumerate services: %w", err)
	}
	if bytesNeeded == 0 {
		return []ServiceInfo{}, nil
	}

	buf := make([]byte, bytesNeeded)
	if err := windows.EnumServicesStatusEx(
		m.Handle,
		windows.SC_ENUM_PROCESS_INFO,
		windows.SERVICE_WIN32,
		windows.SERVICE_STATE_ALL,
		&buf[0], uint32(len(buf)),
		&bytesNeeded, &servicesReturned, &resumeHandle, nil,
	); err != nil {
		return nil, fmt.Errorf("enumerate services: %w", err)
	}

	statuses := unsafe.Slice(
		(*windows.ENUM_SERVICE_STATUS_PROCESS)(unsafe.Pointer(&buf[0])),
		int(servicesReturned),
	)

	services := make([]ServiceInfo, 0, len(statuses))
	for _, s := range statuses {
		name := windows.UTF16PtrToString(s.ServiceName)
		if name == "" {
			continue
		}

		active, sub := mapServiceState(svc.State(s.ServiceStatusProcess.CurrentState))
		services = append(services, ServiceInfo{
			Name:        name,
			LoadState:   loadLoaded,
			ActiveState: active,
			SubState:    sub,
			Description: windows.UTF16PtrToString(s.DisplayName),
		})
	}

	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })
	return services, nil
}

// mapServiceState translates a Windows service state into the systemd
// vocabulary the UI and StatusColor already understand.
func mapServiceState(state svc.State) (activeState, subState string) {
	switch state {
	case svc.Running:
		return stateActive, subRunning
	case svc.Stopped:
		return stateInactive, subDead
	case svc.StartPending:
		return stateActivating, "start-pre"
	case svc.StopPending:
		return stateDeactivating, "stop-sigterm"
	case svc.Paused:
		return stateInactive, "paused"
	case svc.PausePending, svc.ContinuePending:
		return stateActivating, "pending"
	default:
		return stateInactive, subDead
	}
}

// openServiceForAction opens a service with only the rights the action needs,
// so a user granted rights on one service does not need full admin.
func openServiceForAction(action, serviceName string) (*mgr.Mgr, *mgr.Service, error) {
	var access uint32
	switch action {
	case "start":
		access = windows.SERVICE_START | windows.SERVICE_QUERY_STATUS
	case "stop":
		access = windows.SERVICE_STOP | windows.SERVICE_QUERY_STATUS
	case "restart":
		access = windows.SERVICE_START | windows.SERVICE_STOP | windows.SERVICE_QUERY_STATUS
	case "enable", "disable":
		access = windows.SERVICE_CHANGE_CONFIG | windows.SERVICE_QUERY_CONFIG
	default:
		return nil, nil, errUnsupportedAction(action)
	}

	m, err := openSCM(windows.SC_MANAGER_CONNECT)
	if err != nil {
		return nil, nil, err
	}

	namePtr, err := windows.UTF16PtrFromString(serviceName)
	if err != nil {
		_ = m.Disconnect()
		return nil, nil, fmt.Errorf("invalid service name %q: %w", serviceName, err)
	}

	h, err := windows.OpenService(m.Handle, namePtr, access)
	if err != nil {
		_ = m.Disconnect()
		return nil, nil, fmt.Errorf("open service %s: %w", serviceName, err)
	}

	return m, &mgr.Service{Name: serviceName, Handle: h}, nil
}

func controlService(action, serviceName string) error {
	m, s, err := openServiceForAction(action, serviceName)
	if err != nil {
		return err
	}
	defer func() {
		_ = s.Close()
		_ = m.Disconnect()
	}()

	switch action {
	case "start":
		if err := s.Start(); err != nil {
			return fmt.Errorf("start %s: %w", serviceName, err)
		}
	case "stop":
		if err := stopService(s); err != nil {
			return fmt.Errorf("stop %s: %w", serviceName, err)
		}
	case "restart":
		if err := stopService(s); err != nil {
			return fmt.Errorf("restart %s: %w", serviceName, err)
		}
		if err := s.Start(); err != nil {
			return fmt.Errorf("restart %s: %w", serviceName, err)
		}
	case "enable":
		if err := setStartType(s, mgr.StartAutomatic); err != nil {
			return fmt.Errorf("enable %s: %w", serviceName, err)
		}
	case "disable":
		if err := setStartType(s, mgr.StartDisabled); err != nil {
			return fmt.Errorf("disable %s: %w", serviceName, err)
		}
	}

	return nil
}

// stopService sends a stop and waits for the service to settle, so a restart
// does not try to start a service that is still stopping.
func stopService(s *mgr.Service) error {
	status, err := s.Control(svc.Stop)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(20 * time.Second)
	for status.State != svc.Stopped {
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for the service to stop")
		}
		time.Sleep(300 * time.Millisecond)

		status, err = s.Query()
		if err != nil {
			return err
		}
	}
	return nil
}

// setStartType changes only the start type, leaving the rest of the config
// untouched.
func setStartType(s *mgr.Service, startType uint32) error {
	cfg, err := s.Config()
	if err != nil {
		return err
	}
	cfg.StartType = startType
	return s.UpdateConfig(cfg)
}

// serviceManagerAvailable reports whether the SCM can be read.
func serviceManagerAvailable() bool {
	m, err := openSCM(windows.SC_MANAGER_CONNECT | windows.SC_MANAGER_ENUMERATE_SERVICE)
	if err != nil {
		return false
	}
	_ = m.Disconnect()
	return true
}
