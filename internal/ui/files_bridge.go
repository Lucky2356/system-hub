package ui

import (
	"fmt"
	"sync"
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
		return fmt.Errorf("вкладка «Файлы» ещё не готова")
	}

	return fn(path)
}