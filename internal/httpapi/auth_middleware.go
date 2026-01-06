//go:build ignore

package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sonbn-225/goen-api/internal/config"
)

type ctxKey int

const ctxKeyUserID ctxKey = iota

func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxKeyUserID)
	id, ok := v.(string)
	return id, ok && id != ""
}

type AuthClaims struct {
	jwt.RegisteredClaims
}

func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	secret := []byte(cfg.JWTSecret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":{"code":"unauthorized","message":"missing Authorization header"}}`))
				return
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":{"code":"unauthorized","message":"invalid Authorization header"}}`))
				return
			}
			tokStr := strings.TrimSpace(parts[1])
			if tokStr == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":{"code":"unauthorized","message":"empty bearer token"}}`))
				return
			}

			claims := &AuthClaims{}
			token, err := jwt.ParseWithClaims(tokStr, claims, func(t *jwt.Token) (any, error) {
				return secret, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
			if err != nil || token == nil || !token.Valid || claims.Subject == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":{"code":"unauthorized","message":"invalid token"}}`))
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyUserID, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
