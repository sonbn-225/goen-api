// Package httpapi provides HTTP middleware and utilities.
package httpapi

import (
	"context"
	"net/http"

	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/platform/httpx"
)

type Claims = httpx.Claims

// UserIDFromContext is a compatibility wrapper; prefer platform/httpx.UserIDFromContext.
func UserIDFromContext(ctx context.Context) (string, bool) {
	return httpx.UserIDFromContext(ctx)
}

// LangFromContext is a compatibility wrapper; prefer platform/httpx.LangFromContext.
func LangFromContext(ctx context.Context) string {
	return httpx.LangFromContext(ctx)
}

// AuthMiddleware is a compatibility wrapper; prefer platform/httpx.AuthMiddleware.
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return httpx.AuthMiddleware(cfg)
}

// OptionalAuthMiddleware is a compatibility wrapper; prefer platform/httpx.OptionalAuthMiddleware.
func OptionalAuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return httpx.OptionalAuthMiddleware(cfg)
}

// CORSMiddleware is a compatibility wrapper; prefer platform/httpx.CORSMiddleware.
func CORSMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return httpx.CORSMiddleware(cfg)
}

// RequestLogger is a compatibility wrapper; prefer platform/httpx.RequestLogger.
func RequestLogger() func(http.Handler) http.Handler {
	return httpx.RequestLogger()
}
