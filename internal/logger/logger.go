package logger

import (
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	once   sync.Once
	logger *log.Logger
)

// Get returns a singleton instance of the logger.
func Get() *log.Logger {
	once.Do(func() {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			log.Fatalf("could not get user cache dir: %v", err)
		}
		logFile := filepath.Join(cacheDir, "procguard", "events.log")

		if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
			log.Fatalf("could not create log directory: %v", err)
		}

		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("could not open log file: %v", err)
		}

		logger = log.New(f, "", 0)
	})
	return logger
}
