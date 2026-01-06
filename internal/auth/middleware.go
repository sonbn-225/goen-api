package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/config"
)

type ctxKey int

const ctxKeyUserID ctxKey = iota

func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxKeyUserID)
	id, ok := v.(string)
	return id, ok && id != ""
}

type Claims struct {
	jwt.RegisteredClaims
}

func Middleware(cfg *config.Config) func(http.Handler) http.Handler {
	secret := []byte(cfg.JWTSecret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				apierror.Write(w, http.StatusUnauthorized, "unauthorized", "missing Authorization header", nil)
				return
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				apierror.Write(w, http.StatusUnauthorized, "unauthorized", "invalid Authorization header", nil)
				return
			}
			tokStr := strings.TrimSpace(parts[1])
			if tokStr == "" {
				apierror.Write(w, http.StatusUnauthorized, "unauthorized", "empty bearer token", nil)
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokStr, claims, func(t *jwt.Token) (any, error) {
				return secret, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
			if err != nil || token == nil || !token.Valid || claims.Subject == "" {
				apierror.Write(w, http.StatusUnauthorized, "unauthorized", "invalid token", nil)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyUserID, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
