package httpx

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{ResponseWriter: w, status: http.StatusOK}
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func RequestLogger() func(http.Handler) http.Handler {
	logger := slog.Default()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			correlationID := strings.TrimSpace(r.Header.Get("X-Correlation-ID"))
			if correlationID == "" {
				correlationID = chimw.GetReqID(r.Context())
			}
			if correlationID == "" {
				correlationID = uuid.NewString()
			}

			reqLogger := logger.With(
				"correlation_id", correlationID,
				"request_id", chimw.GetReqID(r.Context()),
			)
			ctx := logx.WithCorrelationID(r.Context(), correlationID)
			ctx = logx.WithLogger(ctx, reqLogger)
			r = r.WithContext(ctx)

			startedAt := time.Now()
			recorder := newStatusRecorder(w)
			next.ServeHTTP(recorder, r)

			attrs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.status,
				"bytes", recorder.bytes,
				"duration_ms", time.Since(startedAt).Milliseconds(),
				"correlation_id", correlationID,
			}
			if reqID := chimw.GetReqID(r.Context()); reqID != "" {
				attrs = append(attrs, "request_id", reqID)
			}
			if authz := strings.TrimSpace(r.Header.Get("Authorization")); authz != "" {
				attrs = append(attrs, "authorization", authz)
			}
			attrs = logx.MaskAttrs(attrs...)

			switch {
			case recorder.status >= http.StatusInternalServerError:
				reqLogger.Error("http_request", attrs...)
			case recorder.status >= http.StatusBadRequest:
				reqLogger.Warn("http_request", attrs...)
			default:
				reqLogger.Info("http_request", attrs...)
			}
		})
	}
}
