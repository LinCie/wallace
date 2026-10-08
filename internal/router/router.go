package router

import (
	"net/http"

	"github.com/jmoiron/sqlx"

	"wallace/internal/config"
	"wallace/internal/httpx"
	"wallace/internal/middleware"
)

type Router struct {
	mux *http.ServeMux
	cfg config.Config
}

func NewRouter(db *sqlx.DB, cfg config.Config) Router {
	mux := http.NewServeMux()

	return Router{
		mux: mux,
		cfg: cfg,
	}
}

func (r Router) Handler() http.Handler {
	return httpx.NewChain(
		middleware.Logger,
		middleware.Recovery,
		middleware.WithCORS(r.cfg),
	).Then(r.mux)
}
