package utils

import (
	"fmt"
	"strings"
	"time"
)

// ParseTimeOrDate parses a string into a time.Time, supporting ISO 8601, RFC3339, and plain date (YYYY-MM-DD).
func ParseTimeOrDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time format: %s", s)
}

// NormalizeOccurredAt returns a time.Time and a YYYY-MM-DD string.
func NormalizeOccurredAt(occurredAtStr, dateStr, timeStr *string) (time.Time, string, error) {
	if occurredAtStr != nil && strings.TrimSpace(*occurredAtStr) != "" {
		t, err := ParseTimeOrDate(*occurredAtStr)
		if err != nil {
			return time.Time{}, "", fmt.Errorf("occurred_at is invalid: %w", err)
		}
		return t, t.Format("2006-01-02"), nil
	}

	now := time.Now().UTC()
	if dateStr != nil && strings.TrimSpace(*dateStr) != "" {
		d := strings.TrimSpace(*dateStr)
		full := d
		if timeStr != nil && strings.TrimSpace(*timeStr) != "" {
			full += "T" + strings.TrimSpace(*timeStr)
		} else {
			full += "T" + now.Format("15:04:05")
		}
		t, err := ParseTimeOrDate(full)
		if err != nil {
			// Fallback to plain date
			t, err = ParseTimeOrDate(d)
			if err != nil {
				return time.Time{}, "", fmt.Errorf("occurred_date is invalid: %w", err)
			}
		}
		return t, d, nil
	}

	return now, now.Format("2006-01-02"), nil
}
