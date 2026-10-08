package products

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CreateRequest struct {
	ID     *uuid.UUID    `json:"id"`
	Name   string        `json:"name" validate:"required"`
	Slug   string        `json:"slug" validate:"required"`
	Status ProductStatus `json:"status" validate:"required,oneof=active inactive"`
}

type UpdateRequest struct {
	Name   string        `json:"name" validate:"required"`
	Slug   string        `json:"slug" validate:"required"`
	Status ProductStatus `json:"status" validate:"required,oneof=active inactive"`
}

type GetAllQuery struct {
	Status ProductStatus `query:"status" validate:"omitempty,oneof=active inactive"`
	Page   *int          `query:"page" validate:"omitempty,min=1"`
	Limit  *int          `query:"limit" validate:"omitempty,min=1,max=100"`
}

type ProductResponse struct {
	ID        uuid.UUID     `json:"id"`
	Name      string        `json:"name"`
	Slug      string        `json:"slug"`
	Status    ProductStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	DeleteAt  *time.Time    `json:"deleted_at"`
}
