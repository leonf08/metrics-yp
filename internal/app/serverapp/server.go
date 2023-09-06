package serverapp

import (
	"net/http"

	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
	"github.com/leonf08/metrics-yp.git/internal/handlers"
	"github.com/leonf08/metrics-yp.git/internal/logger"
	"github.com/leonf08/metrics-yp.git/internal/storage"
)

type Server struct {
	storage storage.Repository
	config *serverconf.Config
	saver *Saver
	loader *Loader
	logger logger.Logger
}

func NewServer(st storage.Repository, cfg *serverconf.Config, 
				sv *Saver, ld *Loader, logger logger.Logger) *Server {
	return &Server{
		storage: st,
		config: cfg,
		saver: sv,
		loader: ld,
		logger: logger,
	}
}

func (server Server) Run(h http.Handler) error {
	server.logger.Infoln("Running server", "address", server.config.Addr)
	return http.ListenAndServe(server.config.Addr, h)
}

func (server Server) GetMetricHandler() http.HandlerFunc {
	return handlers.GetMetric(server.storage)
}

func (server Server) UpdateMetricHandler() http.HandlerFunc {
	return handlers.UpdateMetric(server.storage)
}

func (server Server) DefaultHandler() http.HandlerFunc {
	return handlers.Default(server.storage)
}

func (server Server) GetMetricJSONHandler() http.HandlerFunc {
	return handlers.GetMetricJSON(server.storage)
}

func (server Server) UpdateMetricJSONHandler() http.HandlerFunc {
	return handlers.UpdateMetricJSON(server.storage)
}

func (server Server) LoggingMiddleware() func(http.Handler) http.Handler {
	return logger.Logging(server.logger)
}

func (server Server) CompressMiddleware() func(http.Handler) http.Handler {
	return handlers.Compress
}