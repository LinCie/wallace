package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Memory      = 64 * 1024
	argon2Iterations  = 3
	argon2Parallelism = 2
	argon2SaltLength  = 16
	argon2KeyLength   = 32
)

type argon2idHasher struct {
	slot chan struct{}
}

func NewArgon2idHasher() *argon2idHasher {
	return &argon2idHasher{slot: make(chan struct{}, 1)}
}

func (h *argon2idHasher) acquire() error {
	select {
	case h.slot <- struct{}{}:
		return nil
	default:
		return ErrHasherBusy
	}
}

func (h *argon2idHasher) Hash(password string) (string, error) {
	if err := h.acquire(); err != nil {
		return "", err
	}
	defer func() { <-h.slot }()

	salt := make([]byte, argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)
	return argon2HashPrefix() + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(key), nil
}

func (h *argon2idHasher) Verify(password string, hash string) (bool, error) {
	// Only accept the supported parameters to bound verification's resource use.
	encoded, ok := strings.CutPrefix(hash, argon2HashPrefix())
	if !ok {
		return false, ErrInvalidPasswordHash
	}

	encodedSalt, encodedKey, ok := strings.Cut(encoded, "$")
	if !ok || len(encodedSalt) != base64.RawStdEncoding.EncodedLen(argon2SaltLength) || len(encodedKey) != base64.RawStdEncoding.EncodedLen(argon2KeyLength) {
		return false, ErrInvalidPasswordHash
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(encodedSalt)
	if err != nil || len(salt) != argon2SaltLength {
		return false, ErrInvalidPasswordHash
	}

	key, err := base64.RawStdEncoding.Strict().DecodeString(encodedKey)
	if err != nil || len(key) != argon2KeyLength {
		return false, ErrInvalidPasswordHash
	}

	if err := h.acquire(); err != nil {
		return false, err
	}
	defer func() { <-h.slot }()

	actual := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)
	return subtle.ConstantTimeCompare(actual, key) == 1, nil
}

func argon2HashPrefix() string {
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$", argon2.Version, argon2Memory, argon2Iterations, argon2Parallelism)
}
