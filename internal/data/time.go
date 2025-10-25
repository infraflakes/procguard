package data

import (
	"fmt"
	"strings"
	"time"
)

// ParseTime is a helper function that parses a string into a time.Time object.
// It supports both relative time strings (e.g., "now", "1 hour ago") and absolute time strings in specific layouts.
func ParseTime(input string) (time.Time, error) {
	now := time.Now()
	lowerInput := strings.ToLower(strings.TrimSpace(input))

	// Handle relative time strings.
	switch lowerInput {
	case "now":
		return now, nil
	case "1 hour ago":
		return now.Add(-1 * time.Hour), nil
	case "24 hours ago":
		return now.Add(-24 * time.Hour), nil
	case "7 days ago":
		return now.AddDate(0, 0, -7), nil
	}

	// Handle absolute time strings.
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
	}

	for _, layout := range layouts {
		parsedTime, err := time.ParseInLocation(layout, input, time.Local)
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse time: %s", input)
}
