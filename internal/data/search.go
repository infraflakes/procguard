package data

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// SearchAppEvents performs a search on the app_events table in the database.
// It returns a slice of string slices, where each inner slice represents a row with the following format:
// [Time, ProcessName, PID, ParentName, ExePath]
func SearchAppEvents(db *sql.DB, query, since, until string) ([][]string, error) {
	var sinceTime, untilTime time.Time
	var err error

	if since != "" {
		sinceTime, err = ParseTime(since)
		if err != nil {
			return nil, fmt.Errorf("could not parse 'since' time: %w", err)
		}
	}

	if until != "" {
		untilTime, err = ParseTime(until)
		if err != nil {
			return nil, fmt.Errorf("could not parse 'until' time: %w", err)
		}
	}

	// Build the SQL query dynamically based on the provided filters.
	q := "SELECT process_name, pid, parent_process_name, exe_path, start_time, end_time FROM app_events WHERE 1=1"
	args := make([]interface{}, 0)

	if query != "" {
		q += " AND (process_name LIKE ? OR parent_process_name LIKE ?)"
		likeQuery := "%" + query + "%"
		args = append(args, likeQuery, likeQuery)
	}

	// The time-based filtering logic includes processes that were running within the specified time window.
	if !sinceTime.IsZero() {
		sinceUnix := sinceTime.Unix()
		// A process is considered within the window if it ended after the 'since' time, or if it hasn't ended yet.
		q += " AND (end_time IS NULL OR end_time >= ?)"
		args = append(args, sinceUnix)
	}

	if !untilTime.IsZero() {
		untilUnix := untilTime.Unix()
		// A process is considered within the window if it started before the 'until' time.
		q += " AND start_time <= ?"
		args = append(args, untilUnix)
	}

	q += " ORDER BY start_time DESC"

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			GetLogger().Printf("Failed to close rows: %v", err)
		}
	}()

	var results [][]string
	for rows.Next() {
		var processName, parentProcessName, exePath string
		var pid int32
		var startTime, endTime sql.NullInt64

		if err := rows.Scan(&processName, &pid, &parentProcessName, &exePath, &startTime, &endTime); err != nil {
			continue
		}

		startTimeStr := time.Unix(startTime.Int64, 0).Format("2006-01-02 15:04:05")

		results = append(results, []string{
			startTimeStr,
			processName,
			strconv.Itoa(int(pid)),
			parentProcessName,
			exePath,
		})
	}

	return results, nil
}
