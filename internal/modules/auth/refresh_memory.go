package auth

import (
	"context"
	"sync"
	"time"
)

type memoryRefreshStore struct {
	mu       sync.Mutex
	sessions map[string]RefreshSession
}

func NewMemoryRefreshStore() *memoryRefreshStore {
	return &memoryRefreshStore{sessions: make(map[string]RefreshSession)}
}

func (s *memoryRefreshStore) Create(_ context.Context, hash string, session RefreshSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneExpired(time.Now())
	s.sessions[hash] = session
	return nil
}

func (s *memoryRefreshStore) Get(_ context.Context, hash string) (RefreshSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[hash]
	if !exists || !time.Now().Before(session.ExpiresAt) {
		delete(s.sessions, hash)
		return RefreshSession{}, ErrInvalidRefreshToken
	}
	return session, nil
}

func (s *memoryRefreshStore) Rotate(_ context.Context, oldHash, newHash string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pruneExpired(time.Now())
	session, exists := s.sessions[oldHash]
	if !exists {
		return ErrInvalidRefreshToken
	}

	delete(s.sessions, oldHash)
	session.ExpiresAt = expiresAt
	s.sessions[newHash] = session
	return nil
}

func (s *memoryRefreshStore) Delete(_ context.Context, hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, hash)
	return nil
}

// Caller must hold mu.
func (s *memoryRefreshStore) pruneExpired(now time.Time) {
	for hash, session := range s.sessions {
		if !now.Before(session.ExpiresAt) {
			delete(s.sessions, hash)
		}
	}
}
