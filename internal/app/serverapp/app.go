package serverapp

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
	"github.com/leonf08/metrics-yp.git/internal/handlers"
	"github.com/leonf08/metrics-yp.git/internal/logger"
	"github.com/leonf08/metrics-yp.git/internal/storage"
	"go.uber.org/zap"
)

func StartApp() error {
	l, err := initLogger()
	if err != nil {
		return err
	}

	log := logger.NewLogger(l)

	address := *flag.String("a", ":8080", "Host address of the server")
	flag.Parse()

	cfg := serverconf.NewConfig(address)
	err = env.Parse(cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg)
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
	h := logger.LoggingMiddleware(log)

	return server.Run(h(router), log)
}

func initLogger() (logger.Logger, error) {
    lvl, err := zap.ParseAtomicLevel("info")
    if err != nil {
        return nil, err
    }
    
    cfg := zap.NewDevelopmentConfig()
    cfg.Level = lvl
    zl, err := cfg.Build()
    if err != nil {
        return nil, err
    }
    
    return zl.Sugar(), nil
}