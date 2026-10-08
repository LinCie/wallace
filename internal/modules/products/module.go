package products

import "github.com/jmoiron/sqlx"

type Module struct {
	Service Service
	Handler *Handler
}

func NewModule(db *sqlx.DB) *Module {
	repo := NewPostgresRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	return &Module{
		Service: service,
		Handler: handler,
	}
}
