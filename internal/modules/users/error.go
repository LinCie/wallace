package users

import "errors"

var (
	ErrFailedToCreateUser     = errors.New("failed to create user")
	ErrUserEmailAlreadyExists = errors.New("user email already exists")
	ErrFailedToGetUser        = errors.New("failed to get user")
	ErrUserNotFound           = errors.New("user not found")
)
