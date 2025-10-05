package logger

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	once      sync.Once
	webOnce   sync.Once
	appLogger *log.Logger
	webLogger *log.Logger
)

// Get returns a singleton instance of the main application logger.
func Get() *log.Logger {
	once.Do(func() {
		appLogger = createLogger("logs")
	})
	return appLogger
}

// GetWebLogger returns a singleton instance of the web logger.
func GetWebLogger() *log.Logger {
	webOnce.Do(func() {
		webLogger = createLogger("web-logs")
	})
	return webLogger
}

// createLogger is a helper function to create a new logger instance.
func createLogger(logSubDir string) *log.Logger {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("could not get user cache dir: %v", err)
	}
	logDir := filepath.Join(cacheDir, "procguard", logSubDir)

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

	return log.New(lumberjackLogger, "", 0)
}
