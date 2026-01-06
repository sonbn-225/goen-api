package httpapi

import (
	"net/http"
	"strings"

	"github.com/sonbn-225/goen-api/internal/config"
)

func CORSMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	allowed := cfg.CORSOrigins
	allowAll := len(allowed) == 1 && allowed[0] == "*"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" {
				if allowAll {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if isOriginAllowed(origin, allowed) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Add("Vary", "Origin")
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,If-Match,If-None-Match,X-Client-Id,Idempotency-Key")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isOriginAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		if strings.EqualFold(a, origin) {
			return true
		}
	}
	return false
}
