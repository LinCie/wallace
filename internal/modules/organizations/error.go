package organizations

import "errors"

var (
	ErrFailedToCreateOrganization = errors.New("failed to create organization")
	ErrFailedToGetOrganization    = errors.New("failed to get organization")
	ErrOrganizationNotFound       = errors.New("organization not found")
	ErrInvalidOrganizationName    = errors.New("organization name must not be empty")
	ErrInvalidLimitRange          = errors.New("invalid limit range")
	ErrInvalidPageRange           = errors.New("invalid page range")
)
