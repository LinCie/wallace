package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"

	"wallace/internal/config"
	"wallace/internal/docs"
	"wallace/internal/middleware"
	"wallace/internal/modules/products"
)

func NewRouter(db *sqlx.DB, cfg config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.StripSlashes)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.WithCORS(cfg))

	r.Get("/docs", docs.Scalar)
	r.Get("/openapi.json", docs.OpenAPI)

	productsModule := products.NewModule(db)

	r.Route("/v1", func(r chi.Router) {
		// r.Use(middleware.WithAuth(cfg))

		r.Route("/products", func(r chi.Router) {
			r.Post("/", productsModule.Handler.Create)
			r.Get("/", productsModule.Handler.GetAll)
			r.Get("/{id}", productsModule.Handler.Get)
			r.Put("/{id}", productsModule.Handler.Update)
			r.Delete("/{id}", productsModule.Handler.Delete)
		})
	})

	return r
}
