package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Logger interface {
	Printf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
	Println(v ...interface{})
}

type multiLogger struct {
	db     *sql.DB
	file   *os.File
	logger *log.Logger
}

func (l *multiLogger) Printf(format string, v ...interface{}) {
	l.write("INFO", fmt.Sprintf(format, v...))
}

func (l *multiLogger) Fatalf(format string, v ...interface{}) {
	l.write("FATAL", fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l *multiLogger) Println(v ...interface{}) {
	l.write("INFO", fmt.Sprintln(v...))
}

func (l *multiLogger) write(level, message string) {
	// Write to database
	_, err := l.db.Exec("INSERT INTO logs (timestamp, level, message) VALUES (?, ?, ?)", time.Now().Unix(), level, message)
	if err != nil {
		log.Printf("Failed to write log to database: %v", err)
		log.Printf("[%s] %s", level, message)
	}
	// Write to file
	l.logger.Printf("[%s] %s", level, message)
}

var defaultLogger Logger
var once sync.Once

func NewLogger(db *sql.DB) {
	once.Do(func() {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			log.Fatalf("Failed to get user cache dir: %v", err)
		}
		logPath := filepath.Join(cacheDir, "procguard", "procguard.log")
		file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		logger := log.New(file, "", log.LstdFlags)
		defaultLogger = &multiLogger{db: db, file: file, logger: logger}
	})
}

func GetLogger() Logger {
	return defaultLogger
}
