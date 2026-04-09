package utils

import "strings"

// NormalizeOptionalString trims whitespace and returns nil if the resulting string is empty.
func NormalizeOptionalString(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

// Coalesce returns the value if not nil, else the default value.
func Coalesce(v *string, def string) string {
	if v == nil {
		return def
	}
	return *v
}

// IsNumeric checks if a string contains only digits.
func IsNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
