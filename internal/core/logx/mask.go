package logx

import "strings"

const redacted = "[REDACTED]"

var sensitiveKeys = map[string]struct{}{
	"authorization":    {},
	"token":            {},
	"access_token":     {},
	"refresh_token":    {},
	"password":         {},
	"current_password": {},
	"new_password":     {},
	"password_hash":    {},
}

func MaskAttrs(attrs ...any) []any {
	if len(attrs) == 0 {
		return attrs
	}
	masked := make([]any, len(attrs))
	copy(masked, attrs)
	for i := 0; i+1 < len(masked); i += 2 {
		key, ok := masked[i].(string)
		if !ok {
			continue
		}
		masked[i+1] = MaskValue(key, masked[i+1])
	}
	return masked
}

func MaskValue(key string, value any) any {
	k := strings.ToLower(strings.TrimSpace(key))
	if _, ok := sensitiveKeys[k]; ok {
		return redacted
	}

	switch v := value.(type) {
	case string:
		return maskStringValue(v)
	case map[string]any:
		out := make(map[string]any, len(v))
		for mk, mv := range v {
			out[mk] = MaskValue(mk, mv)
		}
		return out
	default:
		return value
	}
}

func maskStringValue(s string) string {
	trimmed := strings.TrimSpace(s)
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "bearer ") {
		return "Bearer " + redacted
	}
	if looksLikeJWT(trimmed) {
		return redacted
	}
	return s
}

func looksLikeJWT(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
	}
	return true
}
