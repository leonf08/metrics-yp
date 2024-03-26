package httpserver

import (
	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	"github.com/leonf08/metrics-yp.git/internal/httpserver/middleware"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/leonf08/metrics-yp.git/internal/services/repo"
	"github.com/rs/zerolog"
)

// NewRouter creates a new router and adds middleware.
func NewRouter(
	s *services.HashSigner,
	cr services.Crypto,
	repo repo.Repository,
	fs services.FileStore,
	ip services.IPChecker,
	l zerolog.Logger,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logging(l), middleware.IPCheck(ip), middleware.Auth(s),
		middleware.Crypto(cr), middleware.Compress, chiMw.Recoverer)

	newHandler(r, repo, fs, l)

	return r
}
