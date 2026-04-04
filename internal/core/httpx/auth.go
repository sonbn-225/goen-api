package httpx

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/core/response"
)

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logx.LoggerFromContext(r.Context()).With("layer", "auth_middleware")

			authz := r.Header.Get("Authorization")
			if authz == "" {
				logger.Warn("auth_failed", logx.MaskAttrs("reason", "missing authorization header")...)
				response.WriteError(w, apperrors.New(apperrors.KindUnauth, "missing authorization header"))
				return
			}

			parts := strings.SplitN(authz, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				logger.Warn("auth_failed", logx.MaskAttrs("reason", "invalid authorization scheme", "authorization", authz)...)
				response.WriteError(w, apperrors.New(apperrors.KindUnauth, "invalid authorization scheme"))
				return
			}

			token, err := jwt.Parse(parts[1], func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, apperrors.New(apperrors.KindUnauth, "unexpected signing method")
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				logger.Warn("auth_failed", logx.MaskAttrs("reason", "invalid access token", "token", parts[1])...)
				response.WriteError(w, apperrors.New(apperrors.KindUnauth, "invalid access token"))
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				logger.Warn("auth_failed", logx.MaskAttrs("reason", "invalid token claims")...)
				response.WriteError(w, apperrors.New(apperrors.KindUnauth, "invalid token claims"))
				return
			}

			userID, _ := claims["sub"].(string)
			if userID == "" {
				logger.Warn("auth_failed", logx.MaskAttrs("reason", "token missing subject")...)
				response.WriteError(w, apperrors.New(apperrors.KindUnauth, "token missing subject"))
				return
			}

			ctx := withUserID(r.Context(), userID)
			ctx = logx.WithLogger(ctx, logger.With("user_id", userID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
