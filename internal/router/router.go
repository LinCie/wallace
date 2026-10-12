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
	"wallace/internal/modules/organizations"
	"wallace/internal/modules/products"
	"wallace/internal/modules/users"
)

func NewRouter(db *sqlx.DB, cfg config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(
		chimiddleware.StripSlashes,
		chimiddleware.Logger,
		chimiddleware.Recoverer,
		middleware.WithCORS(cfg),
	)

	r.Get("/docs", docs.Scalar)
	r.Get("/openapi.json", docs.OpenAPI)

	usersModule := users.NewModule(db)
	authModule := auth.NewModule(usersModule.Service, cfg)
	organizationsModule := organizations.NewModule(db)
	productsModule := products.NewModule(db)

	r.Route("/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authModule.Handler.Register)
			r.Post("/login", authModule.Handler.Login)
			r.Post("/refresh", authModule.Handler.Refresh)
			r.Post("/logout", authModule.Handler.Logout)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.WithAuth(cfg))

			r.Route("/organizations", func(r chi.Router) {
				r.Post("/", organizationsModule.Handler.Create)
				r.Get("/", organizationsModule.Handler.GetAll)
				r.Get("/{organizationID}", organizationsModule.Handler.Get)
			})

			r.Group(func(r chi.Router) {
				r.Use(middleware.WithOrganizationAccess(organizationsModule.Service))

				r.Route("/products", func(r chi.Router) {
					r.Post("/", productsModule.Handler.Create)
					r.Get("/", productsModule.Handler.GetAll)
					r.Get("/{id}", productsModule.Handler.Get)
					r.Put("/{id}", productsModule.Handler.Update)
					r.Delete("/{id}", productsModule.Handler.Delete)
				})
			})

		})
	})

	return r
}
