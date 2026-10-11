package organizations

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"

	"wallace/internal/httpx"
	"wallace/internal/middleware"
)

type Service interface {
	Create(ctx context.Context, name string, userID uuid.UUID) (*Organization, error)
	Get(ctx context.Context, id, userID uuid.UUID) (*Organization, error)
	GetAll(ctx context.Context, userID uuid.UUID, limit, page int) ([]Organization, error)
	IsMember(ctx context.Context, organizationID, userID uuid.UUID) (bool, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	var body CreateRequest
	if err := httpx.BindBody(r, &body); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	organization, err := h.service.Create(r.Context(), body.Name, userID)
	if err != nil {
		h.respondError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, OrganizationResponse(*organization))
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	page, limit := 1, 20
	query := GetAllQuery{Page: &page, Limit: &limit}
	if err := httpx.BindQuery(r, &query); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	organizations, err := h.service.GetAll(r.Context(), userID, *query.Limit, *query.Page)
	if err != nil {
		h.respondError(w, err)
		return
	}

	response := make([]OrganizationResponse, len(organizations))
	for i, organization := range organizations {
		response[i] = OrganizationResponse(organization)
	}
	httpx.RespondJSON(w, http.StatusOK, response)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.FromString(chi.URLParam(r, "organizationID"))
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid organization ID")
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	organization, err := h.service.Get(r.Context(), id, userID)
	if err != nil {
		h.respondError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, OrganizationResponse(*organization))
}

func (h *Handler) respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidOrganizationName), errors.Is(err, ErrInvalidLimitRange), errors.Is(err, ErrInvalidPageRange):
		httpx.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
	case errors.Is(err, ErrOrganizationNotFound):
		httpx.RespondError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	default:
		httpx.RespondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
	}
}
