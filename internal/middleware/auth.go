package middleware

import (
	"context"
	"net/http"
	"strings"

	"wallace/internal/config"
	"wallace/internal/httpx"
	"wallace/internal/jwt"

	"github.com/gofrs/uuid/v5"
)

type contextKey string

const userIDKey contextKey = "user_id"

// WithAuth validates the bearer token and stores the UUID user ID in the
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

			userID, err := uuid.FromString(claims.Subject)
			if err != nil || userID == uuid.Nil {
				httpx.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token subject")
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext returns the authenticated user ID set by WithAuth.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}
