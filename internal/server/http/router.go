package http

import (
	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	middleware2 "github.com/leonf08/metrics-yp.git/internal/server/http/middleware"
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
	r.Use(middleware2.Logging(l), middleware2.IPCheck(ip), middleware2.Auth(s),
		middleware2.Crypto(cr), middleware2.Compress, chiMw.Recoverer)

	newHandler(r, repo, fs, l)

	return r
}
