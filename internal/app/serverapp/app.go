package serverapp

import (
	"flag"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
	"github.com/leonf08/metrics-yp.git/internal/handlers"
	"github.com/leonf08/metrics-yp.git/internal/logger"
	"github.com/leonf08/metrics-yp.git/internal/storage"
)

func StartApp() error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	address := flag.String("a", ":8080", "Host address of the server")
	flag.Parse()

	cfg := serverconf.NewConfig(*address)
	err := env.Parse(cfg)
	if err != nil {
		panic(err)
	}

	storage := storage.NewStorage()

	router := chi.NewRouter()
	router.Get("/", handlers.DefaultHandler(storage))
	router.Post("/", handlers.DefaultHandler(storage))
	router.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", handlers.GetMetric(storage))
		r.Post("/", handlers.GetMetricJSON(storage))
	})
	router.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.UpdateMetricJSON(storage))
		r.Post("/{type}/{name}/{val}", handlers.UpdateMetric(storage))
	})

	server := NewServer(storage, cfg)
	return server.Run(logger.LoggingMiddleware(router))
}