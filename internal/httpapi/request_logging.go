package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sonbn-225/goen-api/internal/auth"
)

// RequestLogger logs a single line per request to stdout/stderr (container logs).
// It includes request id, method, path, status, bytes, duration, and user id (if authenticated).
func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip very noisy probes.
			if r.URL.Path == "/healthz" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			dur := time.Since(start)

			reqID := middleware.GetReqID(r.Context())
			userID, _ := auth.UserIDFromContext(r.Context())

			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = path + "?" + r.URL.RawQuery
			}

			slog.Info(
				"http",
				"request_id", reqID,
				"user_id", userID,
				"method", r.Method,
				"path", path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration_ms", dur.Milliseconds(),
				"remote", r.RemoteAddr,
				"ua", r.UserAgent(),
			)
		})
	}
}
