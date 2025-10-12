package daemon

import (
	"database/sql"
	"procguard/internal/app"
	"procguard/internal/data"
)

// Start runs the core daemon logic in goroutines.
func Start(appLogger data.Logger, db *sql.DB) {
	// Goroutine for event-based process logging
	go app.RunEventLogging(appLogger, db)

	// Goroutine for killing blocked processes
	go app.RunProcessKiller(appLogger)
}