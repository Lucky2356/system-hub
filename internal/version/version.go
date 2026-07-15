// Package version exposes the application version, injected at build time.
package version

// Version is the application version. It is overridden at build time via
//
//	go build -ldflags "-X github.com/Lucky2356/system-hub/internal/version.Version=v1.2.3"
//
// and falls back to "dev" for local/unversioned builds.
var Version = "dev"
