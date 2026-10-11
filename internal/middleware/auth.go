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

const (
	userIDKey         contextKey = "user_id"
	organizationIDKey contextKey = "organization_id"
)

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

type OrganizationAccess interface {
	IsMember(ctx context.Context, organizationID, userID uuid.UUID) (bool, error)
}

// WithOrganizationAccess restricts a route to members and stores its organization ID in context.
func WithOrganizationAccess(access OrganizationAccess) httpx.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			organizationID, err := uuid.FromString(r.Header.Get("X-Organization-ID"))
			if err != nil {
				httpx.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "missing or invalid X-Organization-ID header")
				return
			}
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				httpx.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
				return
			}
			isMember, err := access.IsMember(r.Context(), organizationID, userID)
			if err != nil {
				httpx.RespondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to verify organization membership")
				return
			}
			if !isMember {
				httpx.RespondError(w, http.StatusNotFound, "NOT_FOUND", "organization not found")
				return
			}

			ctx := context.WithValue(r.Context(), organizationIDKey, organizationID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OrganizationIDFromContext returns the organization ID set by WithOrganizationAccess.
func OrganizationIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	organizationID, ok := ctx.Value(organizationIDKey).(uuid.UUID)
	return organizationID, ok
}
