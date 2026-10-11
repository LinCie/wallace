package organizations

import (
	"context"
	"strings"

	"github.com/gofrs/uuid/v5"
)

type Repository interface {
	Create(ctx context.Context, name string, userID uuid.UUID) (*Organization, error)
	Get(ctx context.Context, id, userID uuid.UUID) (*Organization, error)
	GetAll(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Organization, error)
	IsMember(ctx context.Context, organizationID, userID uuid.UUID) (bool, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) *service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, name string, userID uuid.UUID) (*Organization, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidOrganizationName
	}
	return s.repo.Create(ctx, name, userID)
}

func (s *service) Get(ctx context.Context, id, userID uuid.UUID) (*Organization, error) {
	return s.repo.Get(ctx, id, userID)
}

func (s *service) GetAll(ctx context.Context, userID uuid.UUID, limit, page int) ([]Organization, error) {
	if limit < 1 || limit > 100 {
		return nil, ErrInvalidLimitRange
	}
	if page < 1 {
		return nil, ErrInvalidPageRange
	}
	return s.repo.GetAll(ctx, userID, limit, (page-1)*limit)
}

func (s *service) IsMember(ctx context.Context, organizationID, userID uuid.UUID) (bool, error) {
	return s.repo.IsMember(ctx, organizationID, userID)
}
