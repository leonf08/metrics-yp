package serverapp

import (
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
	"github.com/leonf08/metrics-yp.git/internal/server/httpserver"
	"github.com/leonf08/metrics-yp.git/internal/server/logger"
	"github.com/leonf08/metrics-yp.git/internal/storage"
	"go.uber.org/zap"
)

func StartApp() error {
	server, err := initServer()
	if err != nil {
		return err
	}

	defer server.Saver.Close()
	defer server.Loader.Close()

	router := chi.NewRouter()
	router.Get("/", server.Default)
	router.Post("/", server.Default)
	router.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", server.GetMetric)
		r.Post("/", server.GetMetricJSON)
	})
	router.Route("/update", func(r chi.Router) {
		r.Post("/", server.UpdateMetricJSON)
		r.Post("/{type}/{name}/{val}", server.UpdateMetric)
	})

	handler := server.LoggingMiddleware(server.CompressMiddleware(router))

	if server.Config.FileStoragePath != "" {
		if server.Config.Restore {
			m, err := server.Loader.LoadMetrics()
			if err != nil {
				return err
			}

			for k, v := range m.Storage {
				if err := server.Storage.SetVal(k, v); err != nil {
					return err
				}
			}
		}

		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt)
		defer close(shutdown)

		if server.Config.StoreInt > 0 {
			timer := time.NewTicker(time.Duration(server.Config.StoreInt) * time.Second)
			defer timer.Stop()

			go func() {
				for {
					select {
					case <-timer.C:
						server.Logger.Infoln("Save current metrics")
						if err := server.Saver.SaveMetrics(); err != nil {
							server.Logger.Errorln(err)
						}
					case <-shutdown:
						server.Logger.Infoln("Save current metrics and shut down the server")
						if err := server.Saver.SaveMetrics(); err != nil {
							server.Logger.Fatalln(err)
						}
						os.Exit(0)
					}
				}
			}()
		} else {
			go func() {
				<-shutdown
				server.Logger.Infoln("Save current metrics and shut down the server")
				if err := server.Saver.SaveMetrics(); err != nil {
					server.Logger.Fatalln(err)
				}
				os.Exit(0)
			}()
		}
	}

	return server.Run(handler)
}

func initServer() (*httpserver.Server, error) {
	l, err := initLogger()
	if err != nil {
		return nil, err
	}
	log := logger.NewLogger(l)

	config, err := processFlagsEnv()
	if err != nil {
		return nil, err
	}

	database, err := storage.NewDb(config.DataBaseAddr)
	if err != nil {
		return nil, err
	}

	storage := storage.NewStorage()

	metricsSaver, err := httpserver.NewSaver(config.FileStoragePath, storage)
	if err != nil {
		return nil, err
	}

	metricsLoader, err := httpserver.NewLoader(config.FileStoragePath)
	if err != nil {
		return nil, err
	}

	return httpserver.NewServer(storage, config, metricsSaver, metricsLoader, log, database), nil
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
	filePath := flag.String("f", "tmp/metrics-db.json", "Path to file for metrics storage")
	dbAddr := flag.String("d", "", "DataBase address")
	restore := flag.Bool("r", true, "Load previously saved metrics at the server start")
	flag.Parse()

	cfg := serverconf.NewConfig(*storeInt, *address, *filePath, *dbAddr, *restore)
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
