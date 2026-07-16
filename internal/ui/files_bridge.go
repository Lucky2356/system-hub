package ui

import (
	"errors"
	"sync"

	"github.com/Lucky2356/system-hub/internal/i18n"
)

var (
	filesOpenMu   sync.RWMutex
	filesOpenFunc func(path string) error
)

func RegisterFilesOpener(fn func(path string) error) {
	filesOpenMu.Lock()
	defer filesOpenMu.Unlock()
	filesOpenFunc = fn
}

func OpenFileInFiles(path string) error {
	filesOpenMu.RLock()
	fn := filesOpenFunc
	filesOpenMu.RUnlock()

	if fn == nil {
		return errors.New(i18n.T("The «Files» tab is not ready yet"))
	}

	return fn(path)
}
