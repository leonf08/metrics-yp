package agentapp

import (
	"flag"
	"net/http"

	"github.com/caarlos0/env/v6"
	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/server/logger"
	"github.com/leonf08/metrics-yp.git/internal/storage"
	"go.uber.org/zap"
)

func StartApp() error {
	l, err := initLogger()
	if err != nil {
		return err
	}

	log := logger.NewLogger(l)

	address := flag.String("a", "localhost:8080", "Host address of the server")
	reportInt := flag.Int("r", 10, "Report interval to server")
	pollInt := flag.Int("p", 2, "Poll interval for metrics")
	flag.Parse()

	cfg := agentconf.NewConfig(*address, *reportInt, *pollInt)
	err = env.Parse(cfg)
	if err != nil {
		return err
	}

	client := &http.Client{}
	storage := storage.NewStorage()

	agent := NewAgent(client, storage, log, cfg)
	return agent.Run()
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
