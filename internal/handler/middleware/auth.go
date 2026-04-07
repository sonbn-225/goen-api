package middleware
 
import (
	"context"
	"fmt"
	"net/http"
	"strings"
 
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)
 
type contextKey string
 
const UserIDKey contextKey = "user_id"
 
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing authorization header", nil)
				return
			}
 
			// Case-insensitive check for "Bearer " prefix
			const bearerPrefix = "bearer "
			if len(authHeader) <= len(bearerPrefix) || strings.ToLower(authHeader[:len(bearerPrefix)]) != bearerPrefix {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid authorization header format", nil)
				return
			}
 
			tokenString := strings.TrimSpace(authHeader[len(bearerPrefix):])
			if tokenString == "" {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized", "empty token", nil)
				return
			}
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(cfg.JWTSecret), nil
			})
 
			if err != nil || !token.Valid {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid or expired token", nil)
				return
			}
 
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid token claims", nil)
				return
			}
 
			userIDStr, ok := claims["sub"].(string)
			if !ok {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid user id in token", nil)
				return
			}
 
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid user id format", nil)
				return
			}
 
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
 
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if userID, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return userID, true
	}
	return uuid.Nil, false
}
