package logger

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
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
		logDir := filepath.Join(cacheDir, "procguard", "logs")

		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Fatalf("could not create log directory: %v", err)
		}

		lumberjackLogger := &lumberjack.Logger{
			Filename:   filepath.Join(logDir, "events.log"),
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, //days
			Compress:   true,
		}

		logger = log.New(lumberjackLogger, "", 0)
	})
	return logger
}