package organizations

import "github.com/jmoiron/sqlx"

type Module struct {
	Service Service
	Handler *Handler
}

func NewModule(db *sqlx.DB) *Module {
	repo := NewPostgresRepository(db)
	service := NewService(repo)
	return &Module{Service: service, Handler: NewHandler(service)}
}
