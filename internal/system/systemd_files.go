package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func FindServiceUnitFile(serviceName string) (string, error) {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return "", fmt.Errorf("service name is empty")
	}

	candidates := []string{
		filepath.Join("/etc/systemd/system", serviceName),
		filepath.Join("/lib/systemd/system", serviceName),
		filepath.Join("/usr/lib/systemd/system", serviceName),
	}

	for _, path := range candidates {
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return path, nil
		}
	}

	// fallback: если передали без .service
	if !strings.HasSuffix(serviceName, ".service") {
		return FindServiceUnitFile(serviceName + ".service")
	}

	return "", fmt.Errorf("unit file not found for service %s", serviceName)
}