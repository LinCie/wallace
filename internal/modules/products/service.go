package products

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

type Repository interface {
	Create(ctx context.Context, product *Product) (*Product, error)
	Get(ctx context.Context, organizationID, id uuid.UUID) (*Product, error)
	GetAll(ctx context.Context, organizationID uuid.UUID, status ProductStatus, limit, offset int) ([]Product, error)
	Update(ctx context.Context, organizationID uuid.UUID, product *Product) (*Product, error)
	Delete(ctx context.Context, organizationID, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) *service {
	return &service{
		repo: repo,
	}
}

func (s *service) Create(ctx context.Context, product *Product) (*Product, error) {
	return s.repo.Create(ctx, product)
}

func (s *service) Get(ctx context.Context, organizationID, id uuid.UUID) (*Product, error) {
	return s.repo.Get(ctx, organizationID, id)
}

func (s *service) GetAll(ctx context.Context, organizationID uuid.UUID, status ProductStatus, limit, page int) ([]Product, error) {
	if limit < 1 || limit > 100 {
		return nil, ErrInvalidLimitRange
	}
	if page < 1 {
		return nil, ErrInvalidPageRange
	}

	offset := (page - 1) * limit
	return s.repo.GetAll(ctx, organizationID, status, limit, offset)
}

func (s *service) Update(ctx context.Context, organizationID uuid.UUID, product *Product) (*Product, error) {
	return s.repo.Update(ctx, organizationID, product)
}

func (s *service) Delete(ctx context.Context, organizationID, id uuid.UUID) error {
	return s.repo.Delete(ctx, organizationID, id)
}
