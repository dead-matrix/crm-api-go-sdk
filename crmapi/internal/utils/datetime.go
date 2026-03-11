package utils

import "time"

// ParseTime attempts to parse API datetime value.
// It returns nil if the input is empty or cannot be parsed.
// If timezone is missing, UTC is assumed.
func ParseTime(value string) *time.Time {
	if value == "" {
		return nil
	}

	// First try RFC3339 / ISO-compatible parsing.
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		utc := t.UTC()
		return &utc
	}

	// Common fallback: datetime without timezone, treat as UTC.
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.999999",
		"2006-01-02 15:04:05.999999",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			utc := time.Date(
				t.Year(),
				t.Month(),
				t.Day(),
				t.Hour(),
				t.Minute(),
				t.Second(),
				t.Nanosecond(),
				time.UTC,
			)
			return &utc
		}
	}

	return nil
}
