package products

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"

	"wallace/internal/httpx"
)

type Service interface {
	Create(ctx context.Context, product *Product) (*Product, error)
	Get(ctx context.Context, id uuid.UUID) (*Product, error)
	GetAll(ctx context.Context, status ProductStatus, limit, page int) ([]Product, error)
	Update(ctx context.Context, product *Product) (*Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body CreateRequest
	if err := httpx.BindBody(r, &body); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	id := uuid.Nil
	if body.ID != nil {
		if body.ID.Version() != 7 || body.ID.Variant() != uuid.VariantRFC4122 {
			httpx.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "id must be a valid UUIDv7")
			return
		}
		id = *body.ID
	}

	product, err := h.service.Create(r.Context(), &Product{
		ID:     id,
		Name:   body.Name,
		Slug:   body.Slug,
		Status: body.Status,
	})
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, ProductResponse(*product))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.FromString(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid product ID")
		return
	}

	product, err := h.service.Get(r.Context(), id)
	if errors.Is(err, ErrProductNotFound) {
		httpx.RespondError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}

	httpx.RespondJSON(w, http.StatusOK, ProductResponse(*product))
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	page, limit := 1, 20
	query := GetAllQuery{Page: &page, Limit: &limit}
	if err := httpx.BindQuery(r, &query); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	products, err := h.service.GetAll(r.Context(), query.Status, *query.Limit, *query.Page)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}

	response := make([]ProductResponse, len(products))
	for i, product := range products {
		response[i] = ProductResponse(product)
	}

	httpx.RespondJSON(w, http.StatusOK, response)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.FromString(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid product ID")
		return
	}

	var body UpdateRequest
	if err := httpx.BindBody(r, &body); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	product, err := h.service.Update(r.Context(), &Product{
		ID:     id,
		Name:   body.Name,
		Slug:   body.Slug,
		Status: body.Status,
	})
	if errors.Is(err, ErrProductNotFound) {
		httpx.RespondError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}

	httpx.RespondJSON(w, http.StatusOK, ProductResponse(*product))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.FromString(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid product ID")
		return
	}

	err = h.service.Delete(r.Context(), id)
	if errors.Is(err, ErrProductNotFound) {
		httpx.RespondError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
