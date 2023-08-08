package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/handlers"
	"github.com/leonf08/metrics-yp.git/internal/storage"
)

func main() {
	serverStorage := storage.NewStorage()
	r := MetricsRouter(serverStorage)

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}

func MetricsRouter(st storage.Repository) chi.Router {
	r := chi.NewRouter()

	r.Get("/", handlers.DefaultHandler(st))
	r.Post("/", handlers.DefaultHandler(st))

	r.Get("/value/{type}/{name}", handlers.GetMetric(st))
	r.Post("/update/{type}/{name}/{val}", handlers.UpdateMetric(st))

	return r
}
