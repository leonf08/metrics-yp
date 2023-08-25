package serverapp

import (
	"net/http"

	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
	"github.com/leonf08/metrics-yp.git/internal/logger"
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

func (server Server) Run(h http.Handler, log logger.Logger) error {
	log.Infoln("Running server", "address", server.config.Addr)
	return http.ListenAndServe(server.config.Addr, h)
}

