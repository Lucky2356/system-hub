package config

import (
	"path/filepath"
	"testing"
)

func TestNormalize(t *testing.T) {
	c := Config{RefreshIntervalSeconds: 0, DefaultLogLines: -5}
	c.Normalize()
	if c.RefreshIntervalSeconds != 2 {
		t.Errorf("RefreshIntervalSeconds = %d, want 2", c.RefreshIntervalSeconds)
	}
	if c.DefaultLogLines != 100 {
		t.Errorf("DefaultLogLines = %d, want 100", c.DefaultLogLines)
	}

	c2 := Config{RefreshIntervalSeconds: 5, DefaultLogLines: 50}
	c2.Normalize()
	if c2.RefreshIntervalSeconds != 5 || c2.DefaultLogLines != 50 {
		t.Errorf("Normalize altered valid values: %+v", c2)
	}
}

func TestNormalizeFixesEmptyLogFile(t *testing.T) {
	// An empty LogFile previously resolved to the config directory itself,
	// so logger.Init failed with "is a directory" and logging was disabled.
	c := Config{LogFile: ""}
	c.Normalize()
	if c.LogFile != defaultLogFileName {
		t.Errorf("LogFile = %q, want %q", c.LogFile, defaultLogFileName)
	}

	path, err := LogFilePath(Config{LogFile: "   "})
	if err != nil {
		t.Fatalf("LogFilePath: %v", err)
	}
	if filepath.Base(path) != defaultLogFileName {
		t.Errorf("LogFilePath base = %q, want %q", filepath.Base(path), defaultLogFileName)
	}
}

func TestNormalizeThemeAndSlices(t *testing.T) {
	c := Config{Theme: "bogus"}
	c.Normalize()
	if c.Theme != "dark" {
		t.Errorf("Theme = %q, want dark", c.Theme)
	}
	if c.FavoriteServices == nil || c.FavoriteContainers == nil {
		t.Error("favorite slices should be non-nil after Normalize")
	}

	light := Config{Theme: "light"}
	light.Normalize()
	if light.Theme != "light" {
		t.Errorf("Theme = %q, want light preserved", light.Theme)
	}
}

func TestDefaultConfig(t *testing.T) {
	c := DefaultConfig()
	if c.RefreshIntervalSeconds <= 0 {
		t.Errorf("default RefreshIntervalSeconds = %d", c.RefreshIntervalSeconds)
	}
	if c.Theme != "dark" {
		t.Errorf("default Theme = %q, want dark", c.Theme)
	}
	if c.FavoriteServices == nil || c.FavoriteContainers == nil {
		t.Error("favorite slices should be non-nil")
	}
}
