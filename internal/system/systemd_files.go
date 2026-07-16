package system

import (
	"fmt"
	"os"
	"path"
	"strings"
)

// unitFileRoots are the only directories a unit file may be resolved from.
var unitFileRoots = []string{
	"/etc/systemd/system",
	"/lib/systemd/system",
	"/usr/lib/systemd/system",
}

// FindServiceUnitFile locates the unit file for a service. The name is
// validated and the resolved path is confined to a known systemd directory:
// joining an unvalidated name would let "../../etc/passwd" escape the root and
// be opened in the Files tab.
//
// Unit paths are always Unix paths, so this uses "path" rather than
// "path/filepath" — filepath would rewrite them with backslashes on Windows and
// the containment check would never match.
func FindServiceUnitFile(serviceName string) (string, error) {
	serviceName = strings.TrimSpace(serviceName)
	if err := validateName("service", serviceName); err != nil {
		return "", err
	}
	if strings.ContainsAny(serviceName, `/\`) {
		return "", fmt.Errorf("invalid service name %q: must not contain path separators", serviceName)
	}

	names := []string{serviceName}
	if !strings.HasSuffix(serviceName, ".service") {
		names = append(names, serviceName+".service")
	}

	for _, name := range names {
		for _, root := range unitFileRoots {
			candidate := path.Join(root, name)
			if !isInsideUnitRoot(candidate) {
				continue
			}
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, nil
			}
		}
	}

	return "", fmt.Errorf("unit file not found for service %s", serviceName)
}

// isInsideUnitRoot reports whether a path is still under a systemd unit
// directory after cleaning.
func isInsideUnitRoot(p string) bool {
	cleaned := path.Clean(p)
	for _, root := range unitFileRoots {
		if strings.HasPrefix(cleaned, root+"/") {
			return true
		}
	}
	return false
}
