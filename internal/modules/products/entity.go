package products

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type ProductStatus string

const (
	StatusActive   ProductStatus = "active"
	StatusInactive ProductStatus = "inactive"
)

type Product struct {
	ID        uuid.UUID     `db:"id"`
	Name      string        `db:"name"`
	Slug      string        `db:"slug"`
	Status    ProductStatus `db:"status"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
	DeleteAt  *time.Time    `db:"deleted_at"`
}
