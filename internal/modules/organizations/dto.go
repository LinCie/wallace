package organizations

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CreateRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

type GetAllQuery struct {
	Page  *int `query:"page" validate:"omitempty,min=1"`
	Limit *int `query:"limit" validate:"omitempty,min=1,max=100"`
}

type OrganizationResponse struct {
	ID        uuid.UUID        `json:"id"`
	Name      string           `json:"name"`
	Role      OrganizationRole `json:"role"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}
