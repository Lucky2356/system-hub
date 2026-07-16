// Package appstate holds the live application config shared across tabs.
package appstate

import (
	"sync"

	"github.com/Lucky2356/system-hub/internal/config"
)

var (
	mu     sync.RWMutex
	Config config.Config
)

func GetConfig() config.Config {
	mu.RLock()
	defer mu.RUnlock()
	return Config
}

func SetConfig(cfg config.Config) {
	mu.Lock()
	defer mu.Unlock()
	Config = cfg
}
