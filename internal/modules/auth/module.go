package auth

import "wallace/internal/config"

type Module struct {
	Service Service
	Handler *Handler
}

func NewModule(users UsersService, refresh RefreshStore, cfg config.Config) *Module {
	service := NewService(users, NewArgon2idHasher(), refresh, cfg)

	return &Module{
		Service: service,
		Handler: NewHandler(service, cfg),
	}
}
