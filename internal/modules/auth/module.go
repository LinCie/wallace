package auth

import "wallace/internal/config"

type Module struct {
	Service Service
	Handler *Handler
}

func NewModule(users UsersService, cfg config.Config) *Module {
	refresh := NewMemoryRefreshStore()
	hasher := NewArgon2idHasher()
	service := NewService(users, hasher, refresh, cfg)
	handler := NewHandler(service, cfg)

	return &Module{
		Service: service,
		Handler: handler,
	}
}
