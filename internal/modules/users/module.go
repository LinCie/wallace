package users

import "github.com/jmoiron/sqlx"

type Module struct {
	Service Service
}

func NewModule(db *sqlx.DB) *Module {
	repository := NewPostgresRepository(db)
	service := NewService(repository)

	return &Module{
		Service: service,
	}
}
