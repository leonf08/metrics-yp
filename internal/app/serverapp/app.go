package serverapp

import (
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
	"github.com/leonf08/metrics-yp.git/internal/logger"
	"github.com/leonf08/metrics-yp.git/internal/storage"
	"go.uber.org/zap"
)

func StartApp() error {
	server, err := initServer()
	if err != nil {
		return err
	}

	defer server.saver.Close()
	defer server.loader.Close()

	router := chi.NewRouter()
	router.Get("/", server.DefaultHandler())
	router.Post("/", server.DefaultHandler())
	router.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", server.GetMetricHandler())
		r.Post("/", server.GetMetricJSONHandler())
	})
	router.Route("/update", func(r chi.Router) {
		r.Post("/", server.UpdateMetricJSONHandler())
		r.Post("/{type}/{name}/{val}", server.UpdateMetricHandler())
	})

	logMw := server.LoggingMiddleware()
	handler := logMw(server.CompressMiddleware()(router))

	if server.config.FileStoragePath != "" {
		if server.config.Restore {
			m, err := server.loader.LoadMetrics()
			if err != nil {
				return err
			}
	
			for k, v := range m.Storage {
				if err := server.storage.SetVal(k, v); err != nil {
					return err
				}
			}
		}

		timer := time.NewTicker(time.Duration(server.config.StoreInt)*time.Second)
		defer timer.Stop()

		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt)
		go func() {
			for {
				select {
				case <- timer.C:
					server.logger.Infoln("Save current metrics")
					server.saver.SaveMetrics()
				case <- shutdown:
					server.logger.Infoln("Save current metrics and shut down the server")
					server.saver.SaveMetrics()
					os.Exit(0)
				}
			}
		}()
	}

	return server.Run(handler)
}

func initServer() (*Server, error) {
	l, err := initLogger()
	if err != nil {
		return nil, err
	}
	log := logger.NewLogger(l)

	config, err := processFlagsEnv()
	if err != nil {
		return nil, err
	}

	storage := storage.NewStorage()

	metricsSaver, err := NewSaver(config.FileStoragePath, storage)
	if err != nil {
		return nil, err
	}

	metricsLoader, err := NewLoader(config.FileStoragePath)
	if err != nil {
		return nil, err
	}

	return NewServer(storage, config, metricsSaver, metricsLoader, log), nil
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
	storeInt := flag.Int("i", 300, "Store interval for the metrics")
	filePath := flag.String("f", "/tmp/metrics-db.json", "Path to file for metrics storage")
	restore := flag.Bool("r", true, "Load previously saved metrics at the server start")
	flag.Parse()

	cfg := serverconf.NewConfig(*storeInt, *address, *filePath, *restore)
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	
	return cfg, nil
}