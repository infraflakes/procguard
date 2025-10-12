package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Logger interface {
	Printf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
	Println(v ...interface{})
}

type dbLogger struct {
	db *sql.DB
}

func (l *dbLogger) Printf(format string, v ...interface{}) {
	l.write("INFO", fmt.Sprintf(format, v...))
}

func (l *dbLogger) Fatalf(format string, v ...interface{}) {
	l.write("FATAL", fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l *dbLogger) Println(v ...interface{}) {
	l.write("INFO", fmt.Sprintln(v...))
}

func (l *dbLogger) write(level, message string) {
	_, err := l.db.Exec("INSERT INTO logs (timestamp, level, message) VALUES (?, ?, ?)", time.Now().Unix(), level, message)
	if err != nil {
		log.Printf("Failed to write log to database: %v", err)
		log.Printf("[%s] %s", level, message)
	}
}

var defaultLogger Logger
var once sync.Once

func NewLogger(db *sql.DB) {
	once.Do(func() {
		defaultLogger = &dbLogger{db: db}
	})
}

func GetLogger() Logger {
	return defaultLogger
}