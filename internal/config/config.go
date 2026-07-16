package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	appDirName         = "system-hub"
	configFileName     = "config.json"
	defaultLogFileName = "system-hub.log"

	// AppID must match the ID in cmd/system-hub/FyneApp.toml; Fyne's
	// Preferences API refuses to work without a unique app ID.
	AppID = "com.lucky2356.systemhub"
)

type Config struct {
	RefreshIntervalSeconds int      `json:"refresh_interval_seconds"`
	DefaultLogLines        int      `json:"default_log_lines"`
	LogFile                string   `json:"log_file"`

	DashboardAutoRefresh bool `json:"dashboard_auto_refresh"`
	ServicesAutoRefresh  bool `json:"services_auto_refresh"`
	DockerAutoRefresh    bool `json:"docker_auto_refresh"`
	LogsAutoRefresh      bool `json:"logs_auto_refresh"`
	LogViewerAutoRefresh bool `json:"log_viewer_auto_refresh"`

	FavoriteServices   []string `json:"favorite_services"`
	FavoriteContainers []string `json:"favorite_containers"`

	Theme string `json:"theme"`
}

func DefaultConfig() Config {
	return Config{
		RefreshIntervalSeconds: 2,
		DefaultLogLines:        100,
		LogFile:                defaultLogFileName,
		DashboardAutoRefresh: true,
		ServicesAutoRefresh:  false,
		DockerAutoRefresh:    false,
		LogsAutoRefresh:      false,
		LogViewerAutoRefresh: false,
		FavoriteServices:   []string{},
		FavoriteContainers: []string{},
		Theme:              "dark",
	}
}

func (c *Config) Normalize() {
	if c.RefreshIntervalSeconds <= 0 {
		c.RefreshIntervalSeconds = 2
	}

	if c.DefaultLogLines <= 0 {
		c.DefaultLogLines = 100
	}

	// An empty LogFile made the caller join the config dir with "", yielding the
	// directory itself, so opening the log failed with "is a directory" and
	// logging was silently disabled.
	if strings.TrimSpace(c.LogFile) == "" {
		c.LogFile = defaultLogFileName
	}

	if c.Theme != "light" && c.Theme != "dark" {
		c.Theme = "dark"
	}

	// Keep slices non-nil so they marshal as [] rather than null.
	if c.FavoriteServices == nil {
		c.FavoriteServices = []string{}
	}
	if c.FavoriteContainers == nil {
		c.FavoriteContainers = []string{}
	}
}

func Load() (Config, error) {
	cfg := DefaultConfig()

	path, err := configFilePath()
	if err != nil {
		return cfg, fmt.Errorf("resolve config path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config file: %w", err)
	}

	if len(data) == 0 {
		cfg.Normalize()
		return cfg, nil
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), fmt.Errorf("parse config file: %w", err)
	}

	cfg.Normalize()
	return cfg, nil
}

func Save(cfg Config) error {
	cfg.Normalize()

	dir, err := configDir()
	if err != nil {
		return fmt.Errorf("resolve config dir: %w", err)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	path := filepath.Join(dir, configFileName)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

func ConfigFilePath() (string, error) {
	return configFilePath()
}

// LogFilePath resolves the log file next to the config file. It normalizes the
// config first so an empty LogFile cannot resolve to the directory itself.
func LogFilePath(cfg Config) (string, error) {
	cfg.Normalize()

	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, cfg.LogFile), nil
}

func configFilePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

func configDir() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, appDirName), nil
}