package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/handlers"
	"github.com/leonf08/metrics-yp.git/internal/logger"
	"github.com/leonf08/metrics-yp.git/internal/storage"
	"go.uber.org/zap"
)

func main() {
	err := run()
	if err != nil {
		logger.Log.Fatal("Failed to run server", zap.Error(err))
	}
}

func metricsRouter(st storage.Repository) chi.Router {
	r := chi.NewRouter()

	r.Get("/", handlers.DefaultHandler(st))
	r.Post("/", handlers.DefaultHandler(st))

	r.Get("/value/{type}/{name}", handlers.GetMetric(st))
	r.Post("/update/{type}/{name}/{val}", handlers.UpdateMetric(st))

	return r
}

func run() error {
	addr := parseFlags()

	serverStorage := storage.NewStorage()
	r := metricsRouter(serverStorage)

	if err := logger.InitLogger(); err != nil {
		return err
	}

	logger.Log.Info("Running server", zap.String("address", addr))

	return http.ListenAndServe(addr, logger.Logger(r))
}