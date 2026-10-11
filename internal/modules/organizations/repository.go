package organizations

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

func (r *postgresRepository) Create(ctx context.Context, name string, userID uuid.UUID) (*Organization, error) {
	id, err := database.GenerateID()
	if err != nil {
		return nil, ErrFailedToCreateOrganization
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, ErrFailedToCreateOrganization
	}
	defer tx.Rollback()

	organization := &Organization{ID: id, Name: name}
	if err := tx.GetContext(ctx, organization, `
		INSERT INTO organizations (id, name)
		VALUES ($1, $2)
		RETURNING id, name, created_at, updated_at
	`, id, name); err != nil {
		return nil, ErrFailedToCreateOrganization
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, $3)
	`, id, userID, RoleOwner); err != nil {
		return nil, ErrFailedToCreateOrganization
	}
	if err := tx.Commit(); err != nil {
		return nil, ErrFailedToCreateOrganization
	}

	organization.Role = RoleOwner
	return organization, nil
}

func (r *postgresRepository) Get(ctx context.Context, id, userID uuid.UUID) (*Organization, error) {
	var organization Organization
	err := r.db.GetContext(ctx, &organization, `
		SELECT o.id, o.name, m.role, o.created_at, o.updated_at
		FROM organizations o
		JOIN organization_members m ON m.organization_id = o.id
		WHERE o.id = $1 AND m.user_id = $2
	`, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrganizationNotFound
		}
		return nil, ErrFailedToGetOrganization
	}
	return &organization, nil
}

func (r *postgresRepository) GetAll(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Organization, error) {
	organizations := make([]Organization, 0, limit)
	err := r.db.SelectContext(ctx, &organizations, `
		SELECT o.id, o.name, m.role, o.created_at, o.updated_at
		FROM organizations o
		JOIN organization_members m ON m.organization_id = o.id
		WHERE m.user_id = $1
		ORDER BY o.id DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, ErrFailedToGetOrganization
	}
	return organizations, nil
}

func (r *postgresRepository) IsMember(ctx context.Context, organizationID, userID uuid.UUID) (bool, error) {
	var isMember bool
	err := r.db.GetContext(ctx, &isMember, `
		SELECT EXISTS (
			SELECT 1 FROM organization_members
			WHERE organization_id = $1 AND user_id = $2
		)
	`, organizationID, userID)
	if err != nil {
		return false, ErrFailedToGetOrganization
	}
	return isMember, nil
}
