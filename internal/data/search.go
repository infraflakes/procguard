package data

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// Search performs a search on the app_events table in the database.
func Search(db *sql.DB, query, since, until string) ([][]string, error) {
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

	// Build the query
	q := "SELECT process_name, pid, parent_process_name, start_time, end_time FROM app_events WHERE 1=1"
	args := make([]interface{}, 0)

	if query != "" {
		q += " AND (process_name LIKE ? OR parent_process_name LIKE ?)"
		likeQuery := "%" + query + "%"
		args = append(args, likeQuery, likeQuery)
	}

	if !sinceTime.IsZero() {
		sinceUnix := sinceTime.Unix()
		q += " AND (end_time IS NULL OR end_time >= ?)"
		args = append(args, sinceUnix)
	}

	if !untilTime.IsZero() {
		untilUnix := untilTime.Unix()
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
		var processName, parentProcessName string
		var pid int32
		var startTime, endTime sql.NullInt64

		if err := rows.Scan(&processName, &pid, &parentProcessName, &startTime, &endTime); err != nil {
			continue
		}

		startTimeStr := time.Unix(startTime.Int64, 0).Format("2006-01-02 15:04:05")

		// Format the results into the structure the frontend expects
		// [Time, ProcessName, PID, ParentName]
		results = append(results, []string{
			startTimeStr,
			processName,
			strconv.Itoa(int(pid)),
			parentProcessName,
		})
	}

	return results, nil
}
