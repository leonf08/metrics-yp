package serverapp

import (
	"flag"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/leonf08/metrics-yp.git/internal/auth"

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

	router := chi.NewRouter()
	router.Use(server.LoggingMiddleware, server.AuthMiddleware, server.CompressMiddleware, middleware.Recoverer)
	router.Route("/", func(r chi.Router) {
		r.Get("/", server.Default)
		r.Post("/", server.Default)
		r.Get("/ping", server.PingDB)
		r.Post("/updates/", server.UpdateMetricsBatch)
		r.Route("/value", func(r chi.Router) {
			r.Get("/{type}/{name}", server.GetMetric)
			r.Post("/", server.GetMetricJSON)
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", server.UpdateMetricJSON)
			r.Post("/{type}/{name}/{val}", server.UpdateMetric)
		})
	})
	server.RegisterHandler(router)

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

	signer := auth.NewHashSigner(config.Key)

	s := httpserver.NewServer(repo, config, log, signer)

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
	dbAddr := flag.String("d", "", "Database address")
	restore := flag.Bool("r", true, "Load previously saved metrics at the server start")
	key := flag.String("k", "", "Authentication key")
	flag.Parse()

	cfg := serverconf.NewConfig(*storeInt, *address, *filePath, *dbAddr, *key, *restore)
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func initRepo(cfg *serverconf.Config) (httpserver.Repository, error) {
	var repo httpserver.Repository
	if cfg.IsInMemStorage() {
		st := storage.NewStorage()

		if cfg.IsFileStorage() {
			err := st.WithFileStorage(cfg.FileStoragePath)
			if err != nil {
				return nil, err
			}
		}
		repo = st
	} else {
		db, err := storage.NewDB(cfg.DatabaseAddr)
		if err != nil {
			return nil, err
		}

		repo = db
	}

	return repo, nil
}
