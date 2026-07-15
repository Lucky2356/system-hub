package config

import "testing"

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
