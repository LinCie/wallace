package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"

	"wallace/internal/config"
	"wallace/internal/docs"
	"wallace/internal/middleware"
	"wallace/internal/modules/auth"
	"wallace/internal/modules/products"
	"wallace/internal/modules/users"
)

func NewRouter(db *sqlx.DB, cfg config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.StripSlashes)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.WithCORS(cfg))

	r.Get("/docs", docs.Scalar)
	r.Get("/openapi.json", docs.OpenAPI)

	usersModule := users.NewModule(db)
	authModule := auth.NewModule(usersModule.Service, auth.NewMemoryRefreshStore(), cfg)
	productsModule := products.NewModule(db)

	r.Route("/v1", func(r chi.Router) {
		// r.Use(middleware.WithAuth(cfg))

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authModule.Handler.Register)
			r.Post("/login", authModule.Handler.Login)
			r.Post("/refresh", authModule.Handler.Refresh)
			r.Post("/logout", authModule.Handler.Logout)
		})

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
