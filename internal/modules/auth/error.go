package auth

import "errors"

var (
	ErrHasherBusy          = errors.New("password service busy")
	ErrInvalidPasswordHash = errors.New("invalid or unsupported password hash")
	ErrInvalidCredentials  = errors.New("invalid email and/or password")
	ErrFailedToRegister    = errors.New("failed to register new user")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)
