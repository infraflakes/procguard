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
	appLogger *log.Logger
)

// Get returns a singleton instance of the main application logger.
func Get() *log.Logger {
	once.Do(func() {
		appLogger = createLogger()
	})
	return appLogger
}

// createLogger is a helper function to create a new logger instance.
func createLogger() *log.Logger {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("could not get user cache dir: %v", err)
	}
	logDir := filepath.Join(cacheDir, "procguard")

	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("could not create log directory: %v", err)
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "app.log"),
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Compress:   true,
	}

	return log.New(lumberjackLogger, "", log.Ldate|log.Ltime)
}
