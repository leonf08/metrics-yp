package httpserver

import (
	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	"github.com/leonf08/metrics-yp.git/internal/httpserver/middleware"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"log/slog"
)

func NewRouter(
	s *services.HashSigner,
	repo services.Repository,
	fs services.FileStore,
	l *slog.Logger,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logging(l), middleware.Auth(s), middleware.Compress, chiMw.Recoverer)

	newHandler(r, repo, fs, l)

	return r
}
