package middleware

import (
	"context"
	"net/http"
	"strings"

	"wallace/internal/config"
	"wallace/internal/httpx"
	"wallace/internal/jwt"
)

type contextKey string

const (
	userIDKey contextKey = "user_id"
	roleKey   contextKey = "role"
)

// WithAuth validates the bearer token and stores the user ID and role in the
// request context.
func WithAuth(cfg config.Config) httpx.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			tokenStr, found := strings.CutPrefix(authHeader, "Bearer ")
			if !found {
				httpx.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing bearer token")
				return
			}

			claims, err := jwt.ValidateToken(cfg, tokenStr)
			if err != nil {
				httpx.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.Subject)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext returns the authenticated user ID set by WithAuth.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

// RoleFromContext returns the authenticated user role set by WithAuth.
func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey).(string)
	return role, ok
}
