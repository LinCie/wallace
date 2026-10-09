package users

import (
	"context"
	"database/sql"
	"errors"

	"wallace/internal/database"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type postgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *postgresRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, user *User) (*User, error) {
	if user.ID == uuid.Nil {
		id, err := database.GenerateID()
		if err != nil {
			return nil, ErrFailedToCreateUser
		}
		user.ID = id
	}

	query := `
		INSERT INTO users (id, name, email, password_hash)
		VALUES (:id, :name, :email, :password_hash)
		RETURNING id, name, email, password_hash, created_at, updated_at, deleted_at
	`
	query, args, err := r.db.BindNamed(query, user)
	if err != nil {
		return nil, ErrFailedToCreateUser
	}

	if err := r.db.GetContext(ctx, user, query, args...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == database.UniqueViolation &&
			pgErr.ConstraintName == "users_email_unique_idx" {
			return nil, ErrUserEmailAlreadyExists
		}
		return nil, ErrFailedToCreateUser
	}

	return user, nil
}

func (r *postgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := `
		SELECT id, name, email, password_hash, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`
	if err := r.db.GetContext(ctx, &user, query, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, ErrFailedToGetUser
	}

	return &user, nil
}
