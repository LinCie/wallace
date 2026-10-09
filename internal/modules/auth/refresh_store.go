package auth

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

type RefreshSession struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
}

// RefreshStore stores token digests. Get and Rotate reject missing or expired
// tokens with ErrInvalidRefreshToken. Delete succeeds for missing tokens.
type RefreshStore interface {
	Create(ctx context.Context, hash string, session RefreshSession) error
	Get(ctx context.Context, hash string) (RefreshSession, error)
	// Rotate atomically replaces a valid old token, preserving its user ID.
	Rotate(ctx context.Context, oldHash, newHash string, expiresAt time.Time) error
	Delete(ctx context.Context, hash string) error
}
