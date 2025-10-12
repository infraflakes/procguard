package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func InitDB() (*sql.DB, error) {
	db, err := openAndConfigureDB()
	if err != nil {
		return nil, err
	}

	if err := createSchema(db); err != nil {
		return nil, fmt.Errorf("could not create schema: %w", err)
	}

	return db, nil
}

// OpenDB opens a connection to the SQLite database with WAL mode enabled.
// It does not attempt to create the schema and should be used by reader clients.
func OpenDB() (*sql.DB, error) {
	return openAndConfigureDB()
}

// openAndConfigureDB is a helper to open and configure the database connection.
func openAndConfigureDB() (*sql.DB, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user cache dir: %w", err)
	}
	dbPath := filepath.Join(cacheDir, "procguard", "procguard.db")

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("could not create database directory: %w", err)
	}

	// Enable WAL mode for better concurrency
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_journal_mode=WAL", dbPath))
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}
	return db, nil
}

func createSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS app_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		process_name TEXT NOT NULL,
		pid INTEGER NOT NULL,
		parent_process_name TEXT,
		start_time INTEGER NOT NULL,
		end_time INTEGER
	);

	CREATE INDEX IF NOT EXISTS idx_app_events_start_time ON app_events (start_time);
	CREATE INDEX IF NOT EXISTS idx_app_events_end_time ON app_events (end_time);
	CREATE INDEX IF NOT EXISTS idx_app_events_pid ON app_events (pid);

	CREATE TABLE IF NOT EXISTS web_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL,
		timestamp INTEGER NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_web_events_timestamp ON web_events (timestamp);
	`
	_, err := db.Exec(schema)
	return err
}
