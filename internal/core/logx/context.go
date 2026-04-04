package logx

import (
	"context"
	"log/slog"
)

type contextKey string

const (
	loggerContextKey        contextKey = "logger"
	correlationIDContextKey contextKey = "correlation_id"
)

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		return ctx
	}
	return context.WithValue(ctx, loggerContextKey, logger)
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if ctx != nil {
		if l, ok := ctx.Value(loggerContextKey).(*slog.Logger); ok && l != nil {
			return l
		}
	}
	return slog.Default()
}

func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		return ctx
	}
	return context.WithValue(ctx, correlationIDContextKey, correlationID)
}

func CorrelationIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(correlationIDContextKey).(string)
	return id
}
