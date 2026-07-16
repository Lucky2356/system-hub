// Package logger writes application diagnostics to a file next to the config.
package logger

import (
	"log"
	"os"
	"path/filepath"
)

var Log *log.Logger

func Init(logPath string) error {
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
	return nil
}
