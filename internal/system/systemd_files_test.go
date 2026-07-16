package system

import (
	"strings"
	"testing"
)

func TestFindServiceUnitFileRejectsTraversal(t *testing.T) {
	// These must never resolve: joining them onto a unit root escapes it, and
	// the result would then be opened in the Files tab.
	malicious := []string{
		"../../etc/passwd",
		"../../../root/.ssh/id_rsa",
		"nginx/../../../etc/shadow",
		`..\..\windows\system32\config\sam`,
		"-x",
		"",
	}

	for _, name := range malicious {
		t.Run(name, func(t *testing.T) {
			path, err := FindServiceUnitFile(name)
			if err == nil {
				t.Fatalf("expected %q to be rejected, got path %q", name, path)
			}
			if path != "" {
				t.Errorf("expected empty path on rejection, got %q", path)
			}
		})
	}
}

func TestIsInsideUnitRoot(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/etc/systemd/system/nginx.service", true},
		{"/lib/systemd/system/docker.service", true},
		{"/usr/lib/systemd/system/ssh.service", true},
		{"/etc/systemd/system/../../passwd", false},
		{"/etc/passwd", false},
		{"/etc/systemd/system", false}, // the root itself is not a unit file
		{"/tmp/evil.service", false},
	}

	for _, tt := range tests {
		if got := isInsideUnitRoot(tt.path); got != tt.want {
			t.Errorf("isInsideUnitRoot(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestFindServiceUnitFileAcceptsPlainNames(t *testing.T) {
	// A well-formed name must not be rejected by validation; it will simply not
	// be found unless the unit exists on this host.
	_, err := FindServiceUnitFile("nginx.service")
	if err != nil && !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected rejection of a valid name: %v", err)
	}
}
