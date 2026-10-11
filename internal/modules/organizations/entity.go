package organizations

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type OrganizationRole string

const (
	RoleOwner  OrganizationRole = "owner"
	RoleMember OrganizationRole = "member"
)

type Organization struct {
	ID        uuid.UUID        `db:"id"`
	Name      string           `db:"name"`
	Role      OrganizationRole `db:"role"`
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt time.Time        `db:"updated_at"`
}
