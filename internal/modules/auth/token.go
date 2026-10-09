package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"
)

// Tokens is an internal result. Only AccessToken and ExpiresIn are sent in JSON;
// RefreshToken is delivered exclusively through an HttpOnly cookie.
type Tokens struct {
	AccessToken      string
	RefreshToken     string
	ExpiresIn        int64
	RefreshExpiresAt time.Time
}

func generateRefreshToken() (string, error) {
	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(token[:]), nil
}

func hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
