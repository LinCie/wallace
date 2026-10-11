package products

import (
	"context"
	"database/sql"
	"errors"
	"wallace/internal/database"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
)

type postgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *postgresRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, product *Product) (*Product, error) {
	if product.ID == uuid.Nil {
		id, err := database.GenerateID()
		if err != nil {
			return nil, ErrFailedToCreateProduct
		}
		product.ID = id
	}

	query := `
		INSERT INTO products (id, organization_id, name, slug, status)
		VALUES (:id, :organization_id, :name, :slug, :status)
		RETURNING id, organization_id, name, slug, status, created_at, updated_at
	`
	query, args, err := r.db.BindNamed(query, product)
	if err != nil {
		return nil, ErrFailedToCreateProduct
	}
	if err := r.db.GetContext(ctx, product, query, args...); err != nil {
		return nil, ErrFailedToCreateProduct
	}
	return product, nil
}

func (r *postgresRepository) Get(ctx context.Context, organizationID, id uuid.UUID) (*Product, error) {
	var product Product
	query := `
		SELECT id, organization_id, name, slug, status, created_at, updated_at, deleted_at
		FROM products
		WHERE organization_id = $1 AND id = $2 AND deleted_at IS NULL
	`
	if err := r.db.GetContext(ctx, &product, query, organizationID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, ErrFailedToGetProduct
	}
	return &product, nil
}

func (r *postgresRepository) GetAll(ctx context.Context, organizationID uuid.UUID, status ProductStatus, limit, offset int) ([]Product, error) {
	query := `
		SELECT id, organization_id, name, slug, status, created_at, updated_at, deleted_at
		FROM products
		WHERE organization_id = :organization_id AND deleted_at IS NULL
	`
	params := map[string]any{"organization_id": organizationID, "limit": limit, "offset": offset}
	if status != "" {
		query += " AND status = :status"
		params["status"] = string(status)
	}
	query += " ORDER BY id DESC LIMIT :limit OFFSET :offset"

	query, args, err := r.db.BindNamed(query, params)
	if err != nil {
		return nil, ErrFailedToGetProduct
	}

	products := make([]Product, 0, limit)
	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		return nil, ErrFailedToGetProduct
	}
	return products, nil
}

func (r *postgresRepository) Update(ctx context.Context, organizationID uuid.UUID, product *Product) (*Product, error) {
	product.OrganizationID = organizationID
	query := `
		UPDATE products
		SET name = :name, slug = :slug, status = :status, updated_at = NOW()
		WHERE id = :id AND organization_id = :organization_id AND deleted_at IS NULL
		RETURNING id, organization_id, name, slug, status, created_at, updated_at, deleted_at
	`
	query, args, err := r.db.BindNamed(query, product)
	if err != nil {
		return nil, ErrFailedToUpdateProduct
	}
	if err := r.db.GetContext(ctx, product, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, ErrFailedToUpdateProduct
	}
	return product, nil
}

func (r *postgresRepository) Delete(ctx context.Context, organizationID, id uuid.UUID) error {
	query := `
		UPDATE products
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return ErrFailedToDeleteProduct
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return ErrFailedToDeleteProduct
	}
	if rows == 0 {
		return ErrProductNotFound
	}
	return nil
}
