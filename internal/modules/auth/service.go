package auth

import (
	"context"
	"errors"
	"time"

	"wallace/internal/config"
	"wallace/internal/jwt"
	"wallace/internal/modules/users"

	"github.com/gofrs/uuid/v5"
)

type HasherService interface {
	Hash(password string) (string, error)
	Verify(password string, hash string) (bool, error)
}

type UsersService interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type service struct {
	users   UsersService
	hasher  HasherService
	refresh RefreshStore
	cfg     config.Config
}

func NewService(users UsersService, hasher HasherService, refresh RefreshStore, cfg config.Config) *service {
	return &service{users: users, hasher: hasher, refresh: refresh, cfg: cfg}
}

func (s *service) Register(ctx context.Context, name, email, password string) (*User, error) {
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	user, err := s.users.Create(ctx, &User{Name: name, Email: email, Hash: hash})
	if err != nil {
		if errors.Is(err, users.ErrUserEmailAlreadyExists) {
			return nil, err
		}
		return nil, ErrFailedToRegister
	}
	return user, nil
}

func (s *service) Login(ctx context.Context, email, password string) (*Tokens, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, users.ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	ok, err := s.hasher.Verify(password, user.Hash)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidCredentials
	}

	tokens, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, err
	}
	if err := s.refresh.Create(ctx, hashRefreshToken(tokens.RefreshToken), RefreshSession{
		UserID: user.ID, ExpiresAt: tokens.RefreshExpiresAt,
	}); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (*Tokens, error) {
	oldHash := hashRefreshToken(refreshToken)
	session, err := s.refresh.Get(ctx, oldHash)
	if err != nil {
		return nil, err
	}

	tokens, err := s.generateTokens(session.UserID)
	if err != nil {
		return nil, err
	}
	if err := s.refresh.Rotate(ctx, oldHash, hashRefreshToken(tokens.RefreshToken), tokens.RefreshExpiresAt); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	return s.refresh.Delete(ctx, hashRefreshToken(refreshToken))
}

func (s *service) generateTokens(userID uuid.UUID) (*Tokens, error) {
	accessToken, err := jwt.GenerateToken(s.cfg, userID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	return &Tokens{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int64(s.cfg.JWTExpiry / time.Second),
		RefreshExpiresAt: time.Now().Add(s.cfg.RefreshTokenExpiry),
	}, nil
}
