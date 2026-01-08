package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sonbn-225/goen-api/internal/config"
)

// OptionalMiddleware tries to extract user id from Authorization Bearer token.
// If missing/invalid, it just continues without modifying context.
// This is useful for request logging where we want user_id when available.
func OptionalMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
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
