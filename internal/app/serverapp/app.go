package serverapp

import (
	"context"
	"flag"

	"go.uber.org/zap"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
	"github.com/leonf08/metrics-yp.git/internal/server/httpserver"
	"github.com/leonf08/metrics-yp.git/internal/server/logger"
	"github.com/leonf08/metrics-yp.git/internal/storage"
)

func StartApp() error {
	server, err := initServer()
	if err != nil {
		return err
	}

	defer server.CloseFile()

	router := chi.NewRouter()
	router.Get("/", server.Default)
	router.Post("/", server.Default)
	router.Route("/", func(r chi.Router) {
		r.Get("/", server.Default)
		r.Get("/ping", server.PingDB)
	})
	router.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", server.GetMetric)
		r.Post("/", server.GetMetricJSON)
	})
	router.Route("/update", func(r chi.Router) {
		r.Post("/", server.UpdateMetricJSON)
		r.Post("/{type}/{name}/{val}", server.UpdateMetric)
	})

	handler := server.LoggingMiddleware(server.CompressMiddleware(router))
	server.RegisterHandler(handler)

	return server.Run()
}

func initServer() (*httpserver.Server, error) {
	l, err := initLogger()
	if err != nil {
		return nil, err
	}
	log := logger.NewLogger(l)

	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	repo, err := initRepo(config)
	if err != nil {
		return nil, err
	}

	var s *httpserver.Server
	if config.DataBaseAddr == "" {
		config.UseDB(false)
		s, err = httpserver.NewServer(repo, config, log,
			httpserver.WithSaverOpt(), httpserver.WithLoaderOpt())
		if err != nil {
			return nil, err
		}
	} else {
		config.UseDB(true)
		s, err = httpserver.NewServer(repo, config, log)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
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

func getConfig() (*serverconf.Config, error) {
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

func initRepo(cfg *serverconf.Config) (httpserver.Repository, error) {
	var repo httpserver.Repository
	if cfg.DataBaseAddr == "" {
		st := storage.NewStorage()
		repo = st
	} else {
		db, err := storage.NewDB(cfg.DataBaseAddr)
		if err != nil {
			return nil, err
		}

		if err = db.CreateTable(context.Background()); err != nil {
			return nil, err
		}

		repo = db
	}

	return repo, nil
}
