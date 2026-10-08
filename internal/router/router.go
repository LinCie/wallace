package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"wallace/internal/config"
	"wallace/internal/middleware"
)

func NewRouter(db *sqlx.DB, cfg config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.WithCORS(cfg))

	return r
}
