package utils

import "time"

// moscowLoc - пояс, в котором CRM хранит наивные datetime. Если в контейнере
// нет tzdata, используется фиксированный UTC+3 (в Москве нет перехода на
// летнее время с 2014 года, поэтому результат совпадает).
var moscowLoc = loadMoscow()

func loadMoscow() *time.Location {
	if loc, err := time.LoadLocation("Europe/Moscow"); err == nil {
		return loc
	}
	return time.FixedZone("MSK", 3*60*60)
}

// ParseTime parses an API datetime value and returns it in UTC.
// It returns nil if the input is empty or cannot be parsed.
//
// Values with an explicit offset or "Z" (RFC3339) keep their instant.
// Values without a timezone are Moscow wall-clock: CRM stores naive MSK
// datetimes and most endpoints serialize them via isoformat() without offset,
// so treating them as UTC would shift every time by +3h.
func ParseTime(value string) *time.Time {
	if value == "" {
		return nil
	}

	if t, err := time.Parse(time.RFC3339, value); err == nil {
		utc := t.UTC()
		return &utc
	}

	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.999999",
		"2006-01-02 15:04:05.999999",
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, moscowLoc); err == nil {
			utc := t.UTC()
			return &utc
		}
	}

	return nil
}
