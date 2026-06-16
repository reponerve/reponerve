package sqlite

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// FlexibleTime scans SQLite DATETIME/TEXT timestamp columns into time.Time.
type FlexibleTime struct {
	Time time.Time
}

// Scan implements sql.Scanner for SQLite timestamp values.
func (ft *FlexibleTime) Scan(src interface{}) error {
	t, err := ParseDateTime(src)
	if err != nil {
		return err
	}
	ft.Time = t
	return nil
}

// ParseDateTime converts SQLite driver values to time.Time.
func ParseDateTime(src interface{}) (time.Time, error) {
	switch v := src.(type) {
	case time.Time:
		return v, nil
	case nil:
		return time.Time{}, nil
	case string:
		return parseDateTimeString(v)
	case []byte:
		return parseDateTimeString(string(v))
	default:
		return time.Time{}, fmt.Errorf("unsupported timestamp type %T", src)
	}
}

func parseDateTimeString(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, nil
	}
	s = normalizeMalformedTimestamp(s)

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse timestamp %q", s)
}

var duplicateNumericTZ = regexp.MustCompile(`^(.+\s[+-]\d{4})\s[+-]\d{4}$`)

// normalizeMalformedTimestamp repairs duplicated timezone suffixes from legacy writes.
// Example: "2026-05-07 12:40:59 -0400 -0400" -> "2026-05-07 12:40:59 -0400"
func normalizeMalformedTimestamp(s string) string {
	for {
		m := duplicateNumericTZ.FindStringSubmatch(s)
		if len(m) != 2 {
			return s
		}
		candidate := m[1]
		fields := strings.Fields(s)
		if len(fields) < 2 {
			return s
		}
		last := fields[len(fields)-1]
		prev := fields[len(fields)-2]
		if last == prev && isNumericTZ(last) {
			s = candidate
			continue
		}
		return s
	}
}

func isNumericTZ(s string) bool {
	if len(s) != 5 {
		return false
	}
	return (s[0] == '+' || s[0] == '-') && strings.Trim(s[1:], "0123456789") == ""
}

// FormatDateTime stores timestamps in a stable RFC3339 representation.
func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
