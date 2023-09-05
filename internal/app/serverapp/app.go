package serverapp

import (
	"flag"
	"net/http"
	"time"
	"os"

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

	config, err := processFlagsEnv()
	if err != nil {
		return err
	}

	metricsSaver, err := NewSaver(config.FileStoragePath)
	if err != nil {
		return err
	}
	defer metricsSaver.Close()

	metricsLoader, err := NewLoader(config.FileStoragePath)
	if err != nil {
		return err
	}
	defer metricsLoader.Close()

	repo := storage.NewStorage()
	router := createRouter(repo)

	logging := logger.LoggingMiddleware(log)
	handler :=  logging(handlers.CompressMiddleware(router))

	s := &http.Server{
		Addr: config.Addr,
		Handler: handler,
	}

	if config.FileStoragePath != "" {
		if config.Restore {
			m, err := metricsLoader.Load()
			if err != nil {
				return err
			}
	
			repo = m
		}

		timer := time.NewTicker(time.Duration(config.StoreInt)*time.Second)
		defer timer.Stop()

		shutdown := make(chan os.Signal, 1)
		go func() {
			for {
				select {
				case <- timer.C:
					log.Infoln("Save current metrics")
					metricsSaver.SaveMetrics(repo)
				case <- shutdown:
					log.Infoln("Save current metrics and shut down the server")
					metricsSaver.SaveMetrics(repo)
					return
				}
			}
		}()
	}

	return s.ListenAndServe()
}

func initLogger() (logger.Logger, error) {
    lvl, err := zap.ParseAtomicLevel("info")
    if err != nil {
        return nil, err
    }
    
    cfg := zap.NewProductionConfig()
    cfg.Level = lvl
    zl, err := cfg.Build()
    if err != nil {
        return nil, err
    }
    
    return zl.Sugar(), nil
}

func processFlagsEnv() (*serverconf.Config, error) {
	address := flag.String("a", ":8080", "Host address of the server")
	storeInt := flag.Int("i", 10, "Store interval for the metrics")
	filePath := flag.String("f", "tmp/metrics-db.json", "Path to file for metrics storage")
	restore := flag.Bool("r", true, "Load previously save metrics at the server start")
	flag.Parse()

	cfg := serverconf.NewConfig(*storeInt, *address, *filePath, *restore)
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	
	return cfg, nil
}

func createRouter(st storage.Repository) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", handlers.DefaultHandler(st))
	router.Post("/", handlers.DefaultHandler(st))
	router.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", handlers.GetMetric(st))
		r.Post("/", handlers.GetMetricJSON(st))
	})
	router.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.UpdateMetricJSON(st))
		r.Post("/{type}/{name}/{val}", handlers.UpdateMetric(st))
	})

	return router
}