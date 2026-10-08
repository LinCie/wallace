package products

import "errors"

var (
	ErrFailedToCreateProduct = errors.New("failed to create product")
	ErrFailedToGetProduct    = errors.New("failed to get product")
	ErrFailedToUpdateProduct = errors.New("failed to update product")
	ErrFailedToDeleteProduct = errors.New("failed to delete product")
	ErrProductNotFound       = errors.New("product not found")
	ErrInvalidLimitRange     = errors.New("invalid limit range")
	ErrInvalidPageRange      = errors.New("invalid page range")
)
