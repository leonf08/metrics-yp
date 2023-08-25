package serverapp

import (
	"net/http"
	"log/slog"

	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
	"github.com/leonf08/metrics-yp.git/internal/storage"
)

type Server struct {
	storage storage.Repository
	config *serverconf.Config
}

func NewServer(st storage.Repository, cfg *serverconf.Config) *Server {
	return &Server{
		storage: st,
		config: cfg,
	}
}

func (server Server) Run(h http.Handler) error {
	slog.Info("Running server", "address", server.config.Addr)
	return http.ListenAndServe(server.config.Addr, h)
}

