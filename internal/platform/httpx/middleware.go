package httpx

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/response"
	"github.com/sonbn-225/goen-api/internal/shared/contracts"
)

type ctxKey int

const (
	ctxKeyUserID ctxKey = iota
	ctxKeyLang
)

func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxKeyUserID)
	id, ok := v.(string)
	return id, ok && id != ""
}

func LangFromContext(ctx context.Context) string {
	v := ctx.Value(ctxKeyLang)
	if s, ok := v.(string); ok {
		return s
	}
	return "en"
}

type Claims struct {
	jwt.RegisteredClaims
}

func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	secret := []byte(cfg.JWTSecret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := UserIDFromContext(r.Context()); ok {
				next.ServeHTTP(w, r)
				return
			}

			auth := r.Header.Get("Authorization")
			if auth == "" {
				response.WriteError(w, http.StatusUnauthorized, contracts.ErrorCodeUnauthorized, "missing Authorization header", nil)
				return
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.WriteError(w, http.StatusUnauthorized, contracts.ErrorCodeUnauthorized, "invalid Authorization header", nil)
				return
			}
			tokStr := strings.TrimSpace(parts[1])
			if tokStr == "" {
				response.WriteError(w, http.StatusUnauthorized, contracts.ErrorCodeUnauthorized, "empty bearer token", nil)
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokStr, claims, func(t *jwt.Token) (any, error) {
				return secret, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
			if err != nil || token == nil || !token.Valid || claims.Subject == "" {
				response.WriteError(w, http.StatusUnauthorized, contracts.ErrorCodeUnauthorized, "invalid token", nil)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyUserID, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	secret := []byte(cfg.JWTSecret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := UserIDFromContext(r.Context()); ok {
				next.ServeHTTP(w, r)
				return
			}

			authz := r.Header.Get("Authorization")
			if authz == "" {
				next.ServeHTTP(w, r)
				return
			}
			parts := strings.SplitN(authz, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				next.ServeHTTP(w, r)
				return
			}
			tokStr := strings.TrimSpace(parts[1])
			if tokStr == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokStr, claims, func(t *jwt.Token) (any, error) {
				return secret, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
			if err != nil || token == nil || !token.Valid || claims.Subject == "" {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyUserID, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

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
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,If-Match,If-None-Match,X-Client-Id,Idempotency-Key,X-Goen-Language")
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

func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := r.Header.Get("Accept-Language")
			if lang == "" {
				lang = r.Header.Get("X-Goen-Language")
			}
			if lang == "" {
				lang = "en"
			}
			ctx := context.WithValue(r.Context(), ctxKeyLang, lang)
			r = r.WithContext(ctx)

			if r.URL.Path == "/healthz" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			dur := time.Since(start)

			reqID := middleware.GetReqID(r.Context())
			userID, _ := UserIDFromContext(r.Context())

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
